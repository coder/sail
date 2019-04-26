package extensions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"golang.org/x/xerrors"
)

type packageMeta struct {
	Name      string `json:"name"`
	Publisher string `json:"publisher"`
}

// ParseExtensionList takes the path to the extensions folder and returns a slice
// of all extensions existing within that folder.
func ParseExtensionList(path string) ([]string, error) {
	pkgs := []string{}

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.Name() != "package.json" {
			return nil
		}

		raw, err := ioutil.ReadFile(path)
		if err != nil {
			return xerrors.Errorf("failed to read package.json: %w", err)
		}

		meta := new(packageMeta)
		err = json.Unmarshal(raw, meta)
		if err != nil {
			return xerrors.Errorf("failed to unmarshal package.json: %w", err)
		}

		if meta.Name == "" || meta.Publisher == "" {
			return nil
		}

		pkgs = append(pkgs, meta.Publisher+"."+meta.Name)
		return nil
	})

	return pkgs, err
}

const extensionIndicator = "# DO NOT EDIT. EXTENSIONS MANAGED BY SAIL."

// DockerfileSetExtensions takes a raw Dockerfile and replaces all extensions with the provided slice.
func DockerfileSetExtensions(fi []byte, exts []string) ([]byte, error) {
	var (
		sep    = splitNewline(fi)
		fmtted = FmtExtensions(exts)
	)

	if len(fi) == 0 {
		buf := bytes.NewBuffer(nil)
		buf.WriteString("FROM codercom/ubuntu-dev\n")
		buf.Write(joinNewline(fmtted))
		return buf.Bytes(), nil
	}

	// cut out enclosing newlines, not needed when we're replacing
	fmtted = fmtted[1 : len(fmtted)-1]
	start, end, err := extensionBounds(sep)
	if err != nil {
		return nil, err
	}

	return joinNewline(append(sep[:start], append(fmtted, sep[end+1:]...)...)), nil
}

// DockerfileAddExtensions takes a raw Dockerfile and adds the given extensions ignoring duplicates,
// returning the resulting Dockerfile.
func DockerfileAddExtensions(fi []byte, toAdd []string) ([]byte, error) {
	existing := extensionsFromDockerfile(splitNewline(fi))
	merged := merge(existing, toAdd)

	return DockerfileSetExtensions(fi, merged)
}

// DockerfileRemoveExtensions takes a raw Dockerfile and removes all given extensions,
// returning the resulting Dockerfile.
func DockerfileRemoveExtensions(fi []byte, toRemove []string) ([]byte, error) {
	existing := extensionsFromDockerfile(splitNewline(fi))
	merged := diff(existing, toRemove)

	return DockerfileSetExtensions(fi, merged)
}

func diff(s1, s2 []string) []string {
	for _, e := range s2 {
		i := 0
		exists := false
		for ii, ee := range s1 {
			if e == ee {
				exists = true
				i = ii
			}

		}

		if exists {
			s1 = append(s1[:i], s1[i+1:]...)
		}
	}

	return s1
}

func merge(s1, s2 []string) []string {
	for _, e := range s2 {
		exists := false
		for _, ee := range s1 {
			if e == ee {
				exists = true
				break
			}
		}

		if !exists {
			s1 = append(s1, e)
		}
	}

	return s1
}

// extensionBounds finds the starting and ending lines of the sail managed
// extension block. Returns an error if one could not be found or is partial.
// fi is a raw Dockerfile split by newlines.
func extensionBounds(fi [][]byte) (start int, end int, err error) {
	var (
		inBlock bool
	)

	for i, e := range fi {
		cnt := bytes.Contains(e, []byte(extensionIndicator))
		if !cnt && !inBlock {
			continue

		} else if cnt {
			inBlock = !inBlock
			if inBlock {
				start = i
			} else {
				end = i
			}

			continue
		}

	}

	if start == 0 || end == 0 {
		err = xerrors.Errorf("failed to find extension bounds. found %v", []int{start, end})
	}

	return
}

var matchExt = regexp.MustCompile(`\S+\.\S+\b`)

// extensionsFromDockerfile parses the currently installed extensions in a Dockerfile
// and returns them as a slice of strings.
// fi is a raw Dockerfile split by newlines.
func extensionsFromDockerfile(fi [][]byte) []string {
	upper, lower, err := extensionBounds(fi)
	if err != nil {
		// if an error is returned, we couldn't find the extension block
		// continue as if there are none
		return []string{}
	}

	exts := []string{}
	for i := upper + 1; i < lower; i++ {
		match := matchExt.Find(fi[i])
		if len(match) != 0 {
			exts = append(exts, string(match))
		}
	}

	return exts
}

// FmtExtensions formats a slice of extensions into a Docker RUN statement split by newlines.
func FmtExtensions(exts []string) [][]byte {
	b := new(bytes.Buffer)
	b.WriteString(fmt.Sprintf("\n%s\n", extensionIndicator))
	b.WriteString("RUN ")
	for i, e := range exts {
		if i == 0 {
			fmt.Fprintf(b, "code-server --install-extension %s", e)
			continue
		}

		fmt.Fprintf(b, " && \\\n\tcode-server --install-extension %s", e)
	}

	b.WriteString("\n")
	b.WriteString(extensionIndicator)
	b.WriteString("\n")
	return splitNewline(b.Bytes())

}

func joinNewline(b [][]byte) []byte {
	return bytes.Join(b, []byte{10})
}

func splitNewline(b []byte) [][]byte {
	return bytes.Split(b, []byte{10})
}

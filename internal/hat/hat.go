package hat

import (
	"bufio"
	"bytes"
	"io/ioutil"

	"golang.org/x/xerrors"

	"go.coder.com/sail/internal/xexec"
)

// DockerReplaceFrom replaces the FROM clause in a Dockerfile
// with the provided base.
func DockerReplaceFrom(dockerFile []byte, base string) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, len(dockerFile)))

	sc := bufio.NewScanner(bytes.NewReader(dockerFile))

	for sc.Scan() {
		byt := sc.Bytes()
		if bytes.HasPrefix(byt, []byte("FROM")) {
			byt = []byte("FROM " + base)
		}

		buf.Write(byt)
		buf.WriteByte('\n')
	}
	return bytes.TrimSpace(buf.Bytes())
}

// ResolveGitHubPath takes a path like ammario/dotfiles
// and downloads it into a temporary direcory.
func ResolveGitHubPath(ghPath string) (string, error) {
	dir, err := ioutil.TempDir("", "hat")
	if err != nil {
		return "", xerrors.Errorf("failed to create tempdir: %w", err)
	}

	cmd := xexec.Fmt("git clone git@github.com:%v.git %v", ghPath, dir)
	xexec.Attach(cmd)
	err = cmd.Run()
	if err != nil {
		return "", xerrors.Errorf("failed to clone hat: %w", err)
	}

	return dir, nil
}

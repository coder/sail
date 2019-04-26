package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"golang.org/x/xerrors"

	"go.coder.com/flog"
	"go.coder.com/sail/internal/extensions"
	"go.coder.com/sail/internal/xexec"
)

var _ command = &extCmd{}

type extCmd struct {
	arg int
}

func (c *extCmd) spec() commandSpec {
	return commandSpec{
		name:      "extensions",
		shortDesc: "manage your VS Code extensions",
		longDesc: `list:
		Lists the extensions installed to your local VS Code.
		
	set:
		Appends or replaces an existing RUN statement that adds the locally installed
		extensions to the provided Dockerfile, or creates a new one if it doesn't exist.
		If no Dockerfile is given, it will be sent to stdout.

		Example:
		sail extensions set Dockerfile

	add:
		Adds the provided extension to the provided Dockerfile.

		Example:
		sail extensions add vscodevim.vim Dockerfile

	remove:
		Removes the provided extension from the provided Dockerfile.

		Example:
		sail extensions remove vscodevim.vim Dockerfile
		`,

		usage: "[list | add <extension> <file> | remove <extension> <file> | set <file>]",
	}
}

func (c *extCmd) nextArg(fl *flag.FlagSet) string {
	c.arg++
	return fl.Arg(c.arg - 1)
}

func (c *extCmd) initFlags(fl *flag.FlagSet) {}

// add adds a provided extension to the provided Dockerfile.
func (c *extCmd) add(fl *flag.FlagSet) error {
	var (
		ext = c.nextArg(fl)
		fi  = c.nextArg(fl)
	)
	fmt.Println(ext, fi)

	stat, err := os.Stat(fi)
	if err != nil {
		return xerrors.Errorf("failed to stat Dockerfile: %w", err)
	}

	raw, err := ioutil.ReadFile(fi)
	if err != nil {
		return xerrors.Errorf("failed to read Dockerfile: %w", err)
	}

	r, err := extensions.DockerfileAddExtensions(raw, []string{ext})
	if err != nil {
		return xerrors.Errorf("failed to replace extensions in Dockerfile: %w", err)
	}

	err = ioutil.WriteFile(fi, r, stat.Mode())
	if err != nil {
		return xerrors.Errorf("failed to write Dockerfile: %w", err)
	}

	return nil
}

// remove removes the provided extension from the provided Dockerfile.
func (c *extCmd) remove(fl *flag.FlagSet) error {
	var (
		ext = c.nextArg(fl)
		fi  = c.nextArg(fl)
	)

	stat, err := os.Stat(fi)
	if err != nil {
		return xerrors.Errorf("failed to stat Dockerfile: %w", err)
	}

	raw, err := ioutil.ReadFile(fi)
	if err != nil {
		return xerrors.Errorf("failed to read Dockerfile: %w", err)
	}

	r, err := extensions.DockerfileRemoveExtensions(raw, []string{ext})
	if err != nil {
		return xerrors.Errorf("failed to replace extensions in Dockerfile: %w", err)
	}

	err = ioutil.WriteFile(fi, r, stat.Mode())
	if err != nil {
		return xerrors.Errorf("failed to write Dockerfile: %w", err)
	}

	return nil
}

// set overrides all previously set extensions with the ones in the local VS Code extension folder.
func (c *extCmd) set(fl *flag.FlagSet) error {
	var (
		hatFile   = fl.Arg(1)
		hatExists = false
	)

	stat, err := os.Stat(hatFile)
	if err == nil {
		hatExists = true
	}

	extDir, err := vscodeExtensionDir()
	if err != nil {
		return xerrors.Errorf("failed to find vscode edition: %s", err)
	}

	exts, err := extensions.ParseExtensionList(extDir)
	if err != nil {
		return xerrors.Errorf("failed to parse extension list: %s", err)
	}

	if len(exts) == 0 {
		fmt.Println("No extensions found.")
		fmt.Println("Is this a mistake? Edit your sail.toml with the correct VS Code path.")
		return nil
	}

	if hatExists {
		raw, err := ioutil.ReadFile(hatFile)
		if err != nil {
			return xerrors.Errorf("failed to read hat file: %s", err)
		}

		r, err := extensions.DockerfileSetExtensions(raw, exts)
		// we only care if this executes successfully beacuse if it errors,
		// that means we can't find the extension block and it'll be added
		// below.
		if err == nil {
			err = ioutil.WriteFile(hatFile, r, stat.Mode())
			if err != nil {
				return xerrors.Errorf("failed to write hat file: %s", err)
			}

			return nil
		}

	}

	buf := bytes.NewBuffer(nil)
	if !hatExists && hatFile != "" {
		r, err := extensions.DockerfileSetExtensions(nil, exts)
		if err != nil {
			return xerrors.Errorf("failed to set extensions: %s", err)
		}
		buf.Write(r)
	} else {
		buf.Write(bytes.Join(extensions.FmtExtensions(exts), []byte{10}))
	}

	if hatFile == "" {
		io.Copy(os.Stdout, buf)
		return nil
	}

	fi, err := os.OpenFile(hatFile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return xerrors.Errorf("failed to open hat file: %s", err)
	}

	_, err = io.Copy(fi, buf)
	if err != nil {
		return xerrors.Errorf("failed to write to hat file: %s", err)
	}

	err = fi.Close()
	if err != nil {
		return xerrors.Errorf("failed to close hat file: %s", err)
	}

	return nil
}

func (c *extCmd) handle(gf globalFlags, fl *flag.FlagSet) {
	defer os.Exit(1)

	switch c.nextArg(fl) {
	case "add":
		err := c.add(fl)
		if err != nil {
			flog.Error("%s", err.Error())
			return
		}

	case "remove":
		err := c.remove(fl)
		if err != nil {
			flog.Error("%s", err.Error())
			return
		}

	case "set":
		err := c.set(fl)
		if err != nil {
			flog.Error("%s", err.Error())
			return
		}

	case "list":
		err := c.list(fl)
		if err != nil {
			flog.Error("%s", err.Error())
			return
		}

	default:
		flog.Error("command %q not found", fl.Arg(0))
	}

	os.Exit(0)
}

func (c *extCmd) list(fl *flag.FlagSet) error {
	exts, err := extensions.ParseExtensionList("/home/colin/.vscode-oss/extensions")
	if err != nil {
		return xerrors.Errorf("failed to parse extension list: %w", err)
	}

	if len(exts) == 0 {
		fmt.Println("No extensions found.")
		fmt.Println("Is this a mistake? Edit your sail.toml with the correct VS Code path.")
		return nil
	}

	for _, e := range exts {
		fmt.Println(e)
	}

	return nil
}

// vscodeExtensionDir finds the extension directory of the default
// VS Code installation.
func vscodeExtensionDir() (string, error) {
	cmdStr := "code -h"
	cmd := xexec.Fmt(cmdStr)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", xerrors.Errorf("failed to run %q:, %w", cmdStr, err)
	}

	e, err := extractVSCodeEdition(out)
	if err != nil {
		return "", xerrors.Errorf("failed to extract vscode edition: %w", err)
	}

	return e.extensionDir(), nil
}

type vscodeEdition string

const (
	vscode         = "Visual Studio Code"
	vscodeOSS      = "Code - OSS"
	vscodeInsiders = "Visual Studio Code - Insiders"
)

func (e vscodeEdition) extensionDir() string {
	var (
		home = os.Getenv("HOME")
		dir  string
	)

	switch e {
	case vscode:
		dir = ".vscode"
	case vscodeOSS:
		dir = ".vscode-oss"
	case vscodeInsiders:
		dir = ".vscode-insiders"
	}

	return path.Join(home, dir, "extensions")
}

// extractVSCodeEdition takes the output of `code -h` and parses it
// for the current edition.
// There are currently 3 editions: code, code-oss, and code-insiders.
func extractVSCodeEdition(out []byte) (vscodeEdition, error) {
	// newline character
	sp := bytes.Split(out, []byte{10})
	if len(sp) == 0 {
		return "", xerrors.New("invalid input: no newlines found")
	}

	switch {
	case bytes.Contains(sp[0], []byte(vscode)):
		return vscode, nil
	case bytes.Contains(sp[0], []byte(vscodeOSS)):
		return vscodeOSS, nil
	case bytes.Contains(sp[0], []byte(vscodeInsiders)):
		return vscodeInsiders, nil
	}

	return "", xerrors.New("failed to find primary vscode version")
}

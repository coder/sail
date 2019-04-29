package main

import (
	"bytes"
	"flag"
	"os"
	"path"

	"golang.org/x/xerrors"

	"go.coder.com/cli"
	"go.coder.com/sail/internal/xexec"
)

var _ cli.Command = &extCmd{}

type extCmd struct {
	arg int
}

func (c *extCmd) Spec() cli.CommandSpec {
	return cli.CommandSpec{
		Name: "extensions",
		Desc: "Manage extensions in a Dockerfile from the command line.",
	}
}

func (c *extCmd) Subcommands() []cli.Command {
	return []cli.Command{
		&extAddCmd{},
		&extListCmd{},
		&extRemoveCmd{},
		&extSetCmd{},
	}
}

func (c *extCmd) RegisterFlags(fl *flag.FlagSet) {}

func (c *extCmd) Run(fl *flag.FlagSet) {
	fl.Usage()
	os.Exit(0)
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
	sp := splitNewline(out)
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

func joinNewline(b [][]byte) []byte {
	return bytes.Join(b, []byte{10})
}

func splitNewline(b []byte) [][]byte {
	return bytes.Split(b, []byte{10})
}

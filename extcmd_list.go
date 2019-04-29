package main

import (
	"flag"
	"fmt"
	"os"

	"go.coder.com/cli"
	"go.coder.com/flog"
	"go.coder.com/sail/internal/extensions"
	"golang.org/x/xerrors"
)

var _ cli.Command = &extListCmd{}

type extListCmd struct{}

func (c *extListCmd) Spec() cli.CommandSpec {
	return cli.CommandSpec{
		Name:   "list",
		Usage:  "list",
		Desc:   "Lists the extensions installed to your local VS Code.",
		Hidden: false,
	}
}

func (c *extListCmd) Run(fl *flag.FlagSet) {
	err := c.list(fl)
	if err != nil {
		flog.Error("failed to list extensions: %s", err)
	}
}

func (c *extListCmd) list(fl *flag.FlagSet) error {
	vscdir, err := vscodeExtensionDir()
	if err != nil {
		return xerrors.Errorf("failed to find VS Code extension directory: %w", err)
	}

	exts, err := extensions.ParseExtensionList(vscdir)
	if err != nil {
		return xerrors.Errorf("failed to parse extension list: %w", err)
	}

	if len(exts) == 0 {
		fmt.Fprintln(os.Stderr, "No extensions found.")
		fmt.Fprintln(os.Stderr, "Is this a mistake? Edit your sail.toml with the correct VS Code path.")
		return nil
	}

	for _, e := range exts {
		fmt.Println(e)
	}

	return nil
}

package main

import (
	"flag"
	"io/ioutil"
	"os"

	"go.coder.com/cli"
	"go.coder.com/flog"
	"go.coder.com/sail/internal/extensions"
	"golang.org/x/xerrors"
)

type extAddCmd struct{}

func (c *extAddCmd) Spec() cli.CommandSpec {
	return cli.CommandSpec{
		Name:  "add",
		Usage: "add <extension> <file>",
		Desc: `Adds the provided extension to the provided Dockerfile.

Example:
sail extensions add vscodevim.vim Dockerfile
		`,
		Hidden: false,
	}
}

func (c *extAddCmd) Run(fl *flag.FlagSet) {
	err := c.add(fl)
	if err != nil {
		flog.Error("failed to add extension to file: %s", err)
	}
}

// add adds a provided extension to the provided Dockerfile.
func (c *extAddCmd) add(fl *flag.FlagSet) error {
	var (
		ext = fl.Arg(0)
		fi  = fl.Arg(0)
	)

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

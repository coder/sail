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

type extRemoveCmd struct{}

func (c *extRemoveCmd) Spec() cli.CommandSpec {
	return cli.CommandSpec{
		Name:   "remove",
		Usage:  "remove <extension> <file>",
		Desc:   "List locally installed extensions",
		Hidden: false,
	}
}

func (c *extRemoveCmd) Run(fl *flag.FlagSet) {
	err := c.remove(fl)
	if err != nil {
		flog.Error("failed to remove extension(s): %s", err)
	}
}

// remove removes the provided extension from the provided Dockerfile.
func (c *extRemoveCmd) remove(fl *flag.FlagSet) error {
	var (
		ext = fl.Arg(0)
		fi  = fl.Arg(1)
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

package main

import (
	"flag"
	"os"

	"go.coder.com/flog"
)

type opencmd struct {
}

func (c *opencmd) spec() commandSpec {
	return commandSpec{
		name:      "open",
		shortDesc: "Opens a project.",
		longDesc:  ``,
		usage:     "[project]",
	}
}

func (c *opencmd) initFlags(fl *flag.FlagSet) {
}

func (c *opencmd) handle(gf globalFlags, fl *flag.FlagSet) {
	proj := gf.project(fl)
	err := proj.open()
	if err != nil {
		flog.Fatal("failed to open project: %v", err)
	}
	os.Exit(0)
}

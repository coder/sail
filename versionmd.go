package main

import (
	"flag"
	"fmt"

	"go.coder.com/cli"
)

var version string

type versioncmd struct {
	print bool
}

func (v *versioncmd) Spec() cli.CommandSpec {
	return cli.CommandSpec{
		Name: "version",
		Desc: fmt.Sprintf("Retrieve the current version"),
	}
}

func (v *versioncmd) RegisterFlags(fl *flag.FlagSet) {
	fl.BoolVar(&v.print, "print", false, "Print the current version of Sail")
}

func (v *versioncmd) Run(fl *flag.FlagSet) {
	fmt.Printf("Sail version: %s", version)
}

package main

import (
	"flag"
	"fmt"

	"go.coder.com/cli"
)

var version string

type versioncmd struct{}

func (v *versioncmd) Spec() cli.CommandSpec {
	return cli.CommandSpec{
		Name: "version",
		Desc: fmt.Sprintf("Retrieve the current version."),
	}
}

func (v *versioncmd) Run(fl *flag.FlagSet) {
	fmt.Println(version)
}

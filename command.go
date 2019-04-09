package main

import "flag"

// commandSpec describes how a command should be used.
// It should not list flags.
type commandSpec struct {
	name      string
	shortDesc string
	longDesc  string
	usage     string
}

type command interface {
	spec() commandSpec
	handle(gf globalFlags, fl *flag.FlagSet)
	// initFlags allows the command to register command-specific flags.
	initFlags(fl *flag.FlagSet)
}

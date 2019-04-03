package main

import (
	"flag"
	"go.coder.com/flog"
)

// handleExec handles the -handleExec command.
func handleExec(repo repo, cnt container) int {
	if flag.NArg() < 2 {
		flog.Fatal("command not provided")
	}

	cmd := cnt.ExecTTY(flag.Args()[1], flag.Args()[2:]...)

	attach(cmd)
	err := cmd.Run()
	if err != nil {
		return 1
	}
	return 0
}

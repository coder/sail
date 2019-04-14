package main

import (
	"os"
	"os/exec"

	"go.coder.com/flog"
)

// handlePrune removes all narwhal container from host.
func handlePrune(flg globalFlags) int {
	list, err := listContainers(true, "")
	if err != nil {
		flog.Fatal("failed to list containers: %v", err)
	}
	if len(list) == 0 {
		return 0
	}

	for _, cnt := range list {
		cmd := exec.Command("docker", "rm", "-f", cnt)
		cmd.Stderr = os.Stderr
		// Stdout is just the container names.

		err = cmd.Run()
		if err != nil {
			flog.Error("failed to remove %v", cnt)
		} else {
			flog.Success("removed %v", cnt)
		}
	}
	return 0
}

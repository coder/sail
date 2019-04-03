package main

import (
	"fmt"
	"os"
	"os/exec"
)

func fmtExec(cmdFmt string, args ...interface{}) *exec.Cmd {
	return exec.Command("bash", "-c", fmt.Sprintf(cmdFmt, args...))
}

func attach(cmd *exec.Cmd) {
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
}

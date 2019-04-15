// Package xexec contains extended os/exec utilities.
package xexec

import (
	"fmt"
	"os"
	"os/exec"
)

func Fmt(cmdFmt string, args ...interface{}) *exec.Cmd {
	return exec.Command("bash", "-c", fmt.Sprintf(cmdFmt, args...))
}

func Attach(cmd *exec.Cmd) {
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
}

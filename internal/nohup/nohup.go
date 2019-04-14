// Package nohup provides the ability to daemonize or "disown" a process
// so that when the current Go program exits, the process still runs as usual.
package nohup

import (
	"os/exec"
)

// Start runs cmd with args.
// It returns an error if it fails to start.
func Start(cmd string, args ...string) error {
	c := exec.Command("nohup", append([]string{cmd}, args...)...)
	return c.Start()
}

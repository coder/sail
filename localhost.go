package main

import (
	"os/exec"
	"runtime"
)

type localhost struct {
}

func (l *localhost) Command(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

func (l *localhost) OS() string {
	return runtime.GOOS
}

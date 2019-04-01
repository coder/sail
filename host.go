package main

import (
	"io"
	"os/exec"
)

type Execer interface {
	Command(name string, args ...string) *exec.Cmd
}

// Host describes a remote or local host.
type Host interface {
	Execer
	// OS returns one of `linux, `darwin`, or `freebsd`.
	OS() string
}

// Clone clones a git repository on h.
// It returns a path to the repository.
func Clone(log io.Writer, h Host, repo Repo) error {
	cmd := h.Command("git", "clone", repo.String(), ".")
	cmd.Stdout = log
	cmd.Stderr = log
	return cmd.Run()
}

// Dial forms a Host from a hostname.
func Dial(hostname string) Host {
	switch hostname {
	case "localhost":
		return &localhost{}
	default:
		panic("No host")
	}
}

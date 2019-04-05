// Package editor works with the OS to find a suitable editor.
// +build linux darwin

package editor

import (
	"os"
)

// Env returns the name of a suitable editor.
func Env() (string, error) {
	envEditor := os.Getenv("EDITOR")
	if envEditor != "" {
		return envEditor, nil
	}

	return "vim", nil
}

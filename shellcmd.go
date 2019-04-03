package main

import (
	"bytes"
	"go.coder.com/flog"
)

// handleShell handles the -handleShell command.
func handleShell(repo repo, cnt container) int {
	out, err := cnt.FmtExec("grep ^.*:.*:$(id -u): /etc/passwd | cut -d : -f 7-").CombinedOutput()
	if err != nil {
		flog.Fatal("failed to get default shell: %v\n%s", err, out)
	}

	cmd := cnt.ExecTTY(string(bytes.TrimSpace(out)))
	attach(cmd)
	err = cmd.Run()
	if err != nil {
		return 1
	}
	return 0
}

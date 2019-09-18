package codeserver

import (
	"strconv"
	"strings"

	"golang.org/x/xerrors"

	"go.coder.com/sail/internal/dockutil"
)

var (
	// PortNotFoundError is returned whenever the port isn't found.
	// This can happen if code-server hasn't started it's listener yet
	// or if the code-server process failed for any reason.
	PortNotFoundError = xerrors.New("failed to find port")
)

// PID returns the pid of code-server running inside of the container.
func PID(containerName string) (int, error) {
	out, err := dockutil.FmtExec(containerName, "pgrep -P 1 code-server").CombinedOutput()
	if err != nil {
		return 0, xerrors.Errorf("%s: %w", out, err)
	}

	return strconv.Atoi(strings.TrimSpace(string(out)))
}

// Port returns the port that code-server is listening on, or
// PortNotFoundError if code-server isn't listening on any port.
func Port(containerName string) (string, error) {
	// Example output:
	// Proto Recv-Q Send-Q Local Address           Foreign Address         State       PID/Program name
	// tcp        0      0 localhost:4774          0.0.0.0:*               LISTEN      6/code-server
	out, err := dockutil.Exec(containerName, "netstat", "-ntpl").CombinedOutput()
	if err != nil {
		return "", xerrors.Errorf("failed to netstat: %s, %w", out, err)
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "code-server") {
			fields := strings.Fields(line)
			const localAddrIndex = 3
			localAddrField := strings.TrimPrefix(fields[localAddrIndex], "127.0.0.1:")
			return strings.TrimPrefix(localAddrField, "localhost:"), nil
		}
	}

	return "", PortNotFoundError
}

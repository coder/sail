package main

import (
	"bytes"
	"errors"
	"go.coder.com/narwhal/internal/xexec"
	"golang.org/x/xerrors"
)

const narwhalLabel = "com.coder.narwhal"

// listContainers lists containers with the given prefix.
// Names are returned in descending order with respect to when it
// was created.
func listContainers(all bool, prefix string) ([]string, error) {
	var allFlag string
	if all {
		allFlag = "-a"
	}

	cmd := xexec.Fmt("docker ps %v --format '{{ .Names }}' --filter name=%v --filter label=%v",
		allFlag, prefix, narwhalLabel,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, xerrors.Errorf("failed to get list of containers: %w", err)
	}

	var names []string
	for _, v := range bytes.Split(bytes.TrimSpace(out), []byte("\n")) {
		v = bytes.TrimSpace(v)
		if string(v) == "" {
			continue
		}
		names = append(names, string(v))
	}
	return names, nil
}

// containerLogPath is the location of the code-server log.
const containerLogPath = "/tmp/code-server.log"

var (
	errCodeServerRunning = errors.New("code-server is already running")
)

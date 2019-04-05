package main

import (
	"bytes"
	"errors"
	"go.coder.com/narwhal/internal/xexec"
	"golang.org/x/xerrors"
	"math/rand"
	"net"
	"strconv"
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

// CodeServerLogLocation is the location of the code-server log.
const CodeServerLogLocation = "/tmp/code-server.log"

// checkPort returns true if the port is bound.
// We want to run this on the host and not in the container
func checkPort(port string) bool {
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return false
	}
	_ = l.Close()
	return true
}

func findAvailablePort() (string, error) {
	const (
		min = 8000
		max = 9000
	)
	for _, tryPort := range rand.Perm(int(max - min)) {
		tryPort += int(min)

		strport := strconv.Itoa(tryPort)
		if checkPort(strport) {
			return strport, nil
		}
	}
	return "", xerrors.New("no availabe ports")
}

var (
	errCodeServerRunning  = errors.New("code-server is already running")
	errCodeServerTimedOut = errors.New("code-server took too long to start")
	errCodeServerFailed   = errors.New("code-server failed to start")
)

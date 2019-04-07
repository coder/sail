package xnet

import (
	"golang.org/x/xerrors"
	"math/rand"
	"net"
	"strconv"
)

// PortFree returns true if the port is bound.
// We want to run this on the host and not in the container
func PortFree(port string) bool {
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return false
	}
	_ = l.Close()
	return true
}

func FindAvailablePort() (string, error) {
	const (
		min = 8000
		max = 9000
	)
	for _, tryPort := range rand.Perm(int(max - min)) {
		tryPort += int(min)

		strport := strconv.Itoa(tryPort)
		if PortFree(strport) {
			return strport, nil
		}
	}
	return "", xerrors.New("no availabe ports")
}

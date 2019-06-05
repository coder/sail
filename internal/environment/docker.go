package environment

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/client"
	"go.coder.com/flog"
	"go.coder.com/sail/internal/randstr"
	"go.coder.com/sail/internal/sshforward"
	"golang.org/x/xerrors"
)

// dockerClient returns an instantiated docker client that
// is using the correct API version. If the client can't be
// constructed, this will panic.
func dockerClient() *client.Client {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(fmt.Sprintf("failed to make docker client: %v", err))
	}

	// Update the API version of the client to match
	// what the server is running.
	cli.NegotiateAPIVersion(context.Background())

	return cli
}

// EnsureRemoteContainerSock ensures that the docker socket it forwarded to
// remote host.
func EnsureRemoteContainerSock(host string) (closeFn func() error, err error) {
	localSock := fmt.Sprintf("/tmp/sail-docker-%s.sock", randstr.Make(5))
	const (
		// Default docker sock.
		remoteSock = "/var/run/docker.sock"
	)

	err = os.Setenv("DOCKER_HOST", "unix:///"+localSock)
	if err != nil {
		return nil, xerrors.Errorf("failed to set DOCKER_HOST env: %w", err)
	}
	flog.Info("forwarding DOCKER_HOST through %s", localSock)

	forwarder := sshforward.NewLocalSocketForwarder(localSock, remoteSock, host)
	err = forwarder.Forward()
	if err != nil {
		return nil, err
	}

	return forwarder.Close, nil
}

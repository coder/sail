package environment

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
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

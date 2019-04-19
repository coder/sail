package dockutil

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"golang.org/x/xerrors"

	"github.com/docker/docker/client"
)

func isNetworkNotFoundError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "No such network")
}

// EnsureNetwork ensures that the network used for sail containers is defined.
func EnsureNetwork(ctx context.Context, cli *client.Client, name, subnet string) error {
	_, err := cli.NetworkInspect(ctx, name, types.NetworkInspectOptions{})
	if err == nil {
		return nil
	}

	_, err = cli.NetworkCreate(ctx, name, types.NetworkCreate{
		Driver: "bridge",
		IPAM: &network.IPAM{
			Driver: "default",
			Config: []network.IPAMConfig{
				{
					Subnet: subnet,
				},
			},
		},
	})
	if err != nil {
		return xerrors.Errorf("failed to create network: %w", err)
	}

	return nil
}

// ContainerIP returns the IP address of the container.
func ContainerIP(ctx context.Context, cli *client.Client, name string) (string, error) {
	insp, err := cli.ContainerInspect(ctx, name)
	if err != nil {
		return "", xerrors.Errorf("failed to inspect container: %w", err)
	}

	if insp.NetworkSettings == nil {
		return "", xerrors.Errorf("network settings are nil, unable to determine container %s IP", name)
	}

	return IPFromEndpointSettings(insp.NetworkSettings.Networks)
}

// IPFromEndpointSettings returns the IP address of the network that a container is using
// based on it's endpoint settings.
func IPFromEndpointSettings(networks map[string]*network.EndpointSettings) (string, error) {
	// We only assign one network endpoint to the container, so it should always
	// be last in the networks map.
	var epSettings *network.EndpointSettings
	for _, settings := range networks {
		epSettings = settings
	}

	if epSettings == nil {
		return "", xerrors.New("unable to find IP address")
	}

	return epSettings.IPAddress, nil
}

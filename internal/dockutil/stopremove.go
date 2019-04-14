package dockutil

import (
	"context"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"golang.org/x/xerrors"
)

func DurationPtr(dur time.Duration) *time.Duration {
	return &dur
}

// StopRemove stops a container and then removes it.
// It is an equivalent to `docker rm -f`.
func StopRemove(ctx context.Context, cli *client.Client, cntName string) error {
	err := cli.ContainerStop(ctx, cntName, DurationPtr(time.Second))
	if err != nil {
		return xerrors.Errorf("failed to stop container %v: %w", cntName, err)
	}

	err = cli.ContainerRemove(ctx, cntName, types.ContainerRemoveOptions{})
	if err != nil {
		return xerrors.Errorf("failed to remove container %v: %w", cntName, err)
	}

	return nil
}

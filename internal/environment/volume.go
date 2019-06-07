package environment

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/volume"
	"golang.org/x/xerrors"
)

// TODO: Some sort of abstraction will need to be in place in order to utilize
// alternative volume drivers.

// localVolume represents a docker volume using the "local" driver.
type localVolume struct {
	vol types.Volume
}

var errMissingVolume = xerrors.Errorf("missing volume")

// findLocalVolume tries to find a docker volume by name. If the volume cannot
// be found, errMissingVolume will be returned.
func findLocalVolume(ctx context.Context, name string) (*localVolume, error) {
	cli := dockerClient()
	defer cli.Close()

	vol, err := cli.VolumeInspect(ctx, name)
	if isVolumeNotFoundError(err) {
		return nil, errMissingVolume
	}
	if err != nil {
		return nil, xerrors.Errorf("failed to inspect volume: %w", err)
	}

	return &localVolume{
		vol: vol,
	}, nil
}

func createLocalVolume(ctx context.Context, name string) (*localVolume, error) {
	cli := dockerClient()
	defer cli.Close()

	volReq := volume.VolumeCreateBody{
		Driver:     "local",
		DriverOpts: map[string]string{},
		Labels:     map[string]string{},
		Name:       name,
	}
	vol, err := cli.VolumeCreate(ctx, volReq)
	if err != nil {
		return nil, xerrors.Errorf("failed to create volume: %w", err)
	}

	return &localVolume{
		vol: vol,
	}, nil
}

func deleteLocalVolume(ctx context.Context, lv *localVolume) error {
	cli := dockerClient()
	defer cli.Close()

	err := cli.VolumeRemove(ctx, lv.vol.Name, false)
	if err != nil {
		return xerrors.Errorf("failed to delete local volume: %v", err)
	}

	return nil
}

func isVolumeNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "No such volume")
}

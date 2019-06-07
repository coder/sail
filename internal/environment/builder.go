package environment

import (
	"context"
	"io"
	"io/ioutil"
	"math/rand"
	"os/user"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"golang.org/x/xerrors"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type BuildConfig struct {
	Name   string
	Image  string
	Envs   []string
	Mounts []mount.Mount
}

func NewDefaultBuildConfig(name string) *BuildConfig {
	var envs []string
	// TODO: Should this be handled better?
	// Ensures that the vscode remote extension works correctly.
	envs = append(envs, "HOME=/home/user")

	return &BuildConfig{
		Name:  name,
		Image: "codercom/ubuntu-dev",
		Envs:  envs,
	}
}

// Build builds an environment using the provided config.
func Build(ctx context.Context, cfg *BuildConfig) (*Environment, error) {
	cli := dockerClient()
	defer cli.Close()

	err := ensureImage(ctx, cfg.Image)
	if err != nil {
		return nil, xerrors.Errorf("failed to ensure image: %w", err)
	}

	cmd := "tail -f /dev/null"

	u, err := user.Current()
	if err != nil {
		return nil, xerrors.Errorf("failed to get current user: %w", err)
	}

	// TODO: Will need to add envs for forwarding ssh agent.
	containerConfig := &container.Config{
		Hostname: cfg.Name,
		Cmd: strslice.StrSlice{
			"bash", "-c", cmd,
		},
		Image:  cfg.Image,
		Labels: map[string]string{},
		User:   u.Uid + ":user",
		Env:    cfg.Envs,
	}

	hostConfig := &container.HostConfig{
		NetworkMode: "host",
		Privileged:  true,
		ExtraHosts: []string{
			containerConfig.Hostname + ":127.0.0.1",
		},
		Mounts: cfg.Mounts,
	}

	_, err = cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, cfg.Name)
	if err != nil {
		return nil, xerrors.Errorf("failed to create container: %w", err)
	}

	cnt, err := cli.ContainerInspect(ctx, cfg.Name)
	if err != nil {
		return nil, xerrors.Errorf("failed to inspect container after create: %w", err)
	}

	env := &Environment{
		name: cfg.Name,
		cnt:  cnt,
	}

	err = Start(ctx, env)
	if err != nil {
		return nil, err
	}

	return env, nil
}

// ensureImage ensures that the image exists on the docker host. If the image
// doesn't exist, the image will be pulled.
func ensureImage(ctx context.Context, image string) error {
	cli := dockerClient()
	defer cli.Close()

	_, _, err := cli.ImageInspectWithRaw(ctx, image)
	if isImageNotFoundError(err) {
		opts := types.ImagePullOptions{}
		rdr, err := cli.ImagePull(ctx, image, opts)
		if err != nil {
			return xerrors.Errorf("failed to pull image: %w", err)
		}
		defer rdr.Close()

		io.Copy(ioutil.Discard, rdr)
	} else if err != nil {
		return xerrors.Errorf("failed to inspect image: %w", err)
	}

	return nil
}

func isImageNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "No such image")
}

package environment

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/user"
	"time"

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
	var (
		envs   []string
		mounts []mount.Mount
	)

	// TODO: This will only work for local sail containers. Some coordination
	// between forwarding the socket over ssh will need to happen.
	sshAuthSock, exists := os.LookupEnv("SSH_AUTH_SOCK")
	if exists {
		env := fmt.Sprintf("SSH_AUTH_SOCK=%s", sshAuthSock)
		envs = append(envs, env)

		sockMount := mount.Mount{
			Type:   mount.TypeBind,
			Source: sshAuthSock,
			Target: sshAuthSock,
		}
		mounts = append(mounts, sockMount)
	}

	// TODO: Should this be handled better?
	// Ensures that the vscode remote extension works correctly.
	envs = append(envs, "HOME=/home/user")

	return &BuildConfig{
		Name:   name,
		Image:  "codercom/ubuntu-dev",
		Envs:   envs,
		Mounts: mounts,
	}
}

// Build builds an environment using the provided config.
func Build(ctx context.Context, cfg *BuildConfig) (*Environment, error) {
	cli := dockerClient()
	defer cli.Close()

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

package environment

import (
	"context"
	"math/rand"
	"os/user"
	"strconv"
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
	return &BuildConfig{
		Name:  name,
		Image: "codercom/ubuntu-dev",
	}
}

func Build(ctx context.Context, cfg *BuildConfig) (*Environment, error) {
	cli := dockerClient()
	defer cli.Close()

	var (
		projDir = "project" // TODO
		port    = codeServerPort()
	)

	// TODO: Remove code-server command, potentially put it into its own hat.
	cmd := "cd " + projDir + "; code-server --host 127.0.0.1" +
		" --port " + port +
		" --data-dir ~/.config/Code --extensions-dir ~/.vscode/extensions --allow-http --no-auth 2>&1"
	cmd = "tail -f /dev/null"

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
		// TODO: Do something about this. This is just to get the vscode remote
		// extension working.
		Env: []string{"HOME=/home/user"},
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

func codeServerPort() string {
	return strconv.Itoa(rand.Intn(65535-1024+1) + 1024) // TODO
}

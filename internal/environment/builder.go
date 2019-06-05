package environment

import (
	"context"
	"fmt"
	"math/rand"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"golang.org/x/xerrors"
)

type BuildConfig struct {
	Name  string
	Image string
	Envs  []string
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
		projDir = defaultDirForRepo(nil) // TODO
		port    = codeServerPort()
	)

	cmd := "cd " + projDir + "; code-server --host 127.0.0.1" +
		" --port " + port +
		" --data-dir ~/.config/Code --extensions-dir ~/.vscode/extensions --allow-http --no-auth 2>&1"

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

	// repoVol, err := ensureVolumeForRepo(ctx, b.repo)
	// if err != nil {
	// 	return nil, err
	// }

	mounts := []mount.Mount{
		{
			Type: mount.TypeVolume,
			// Source: repoVol.vol.Name,
			Target: projDir,
		},
	}

	hostConfig := &container.HostConfig{
		NetworkMode: "host",
		Privileged:  true,
		ExtraHosts: []string{
			containerConfig.Hostname + ":127.0.0.1",
		},
		Mounts: mounts,
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

	// if !b.skipClone {
	// 	err = env.clone(ctx, projDir)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	return env, nil
}

// ensureVolumeForRepo ensures that there's a volume dedicated to storing the
// repo.
func ensureVolumeForRepo(ctx context.Context, r *Repo) (*localVolume, error) {
	name := formatRepoVolumeName(r)
	lv, err := findLocalVolume(ctx, name)
	if xerrors.Is(err, errMissingVolume) {
		lv, err = createLocalVolume(ctx, name)
	}
	if err != nil {
		return nil, err
	}

	return lv, nil
}

func defaultDirForRepo(r *Repo) string {
	// TODO: Correct path.
	return filepath.Join("/home/user/Projects/project")
}

func formatRepoVolumeName(r *Repo) string {
	// TODO: Correct name.
	return fmt.Sprintf("%s_%s", "test", "test")
}

func codeServerPort() string {
	return strconv.Itoa(rand.Intn(65535-1024+1) + 1024) // TODO
}

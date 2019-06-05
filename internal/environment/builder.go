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

// Builder is able to build environments for a repo.
type Builder struct {
	repo  *Repo
	image string
	envs  []string
}

func NewDefaultBuilder(r *Repo) *Builder {
	return &Builder{
		image: "codercom/ubuntu-dev",
		repo:  r,
	}
}

func (b *Builder) Build(ctx context.Context) (*Environment, error) {
	cli := dockerClient()
	defer cli.Close()

	var (
		projDir = defaultDirForRepo(b.repo)
		port    = codeServerPort()
	)

	cmd := "cd " + projDir + "; code-server --host 127.0.0.1" +
		" --port " + port +
		" --data-dir ~/.config/Code --extensions-dir ~/.vscode/extensions --allow-http --no-auth 2>&1"

	u, err := user.Current()
	if err != nil {
		return nil, xerrors.Errorf("failed to get current user: %w", err)
	}

	name := b.repo.DockerName()
	// TODO: Will need to add envs for forwarding ssh agent.
	containerConfig := &container.Config{
		Hostname: name,
		Cmd: strslice.StrSlice{
			"bash", "-c", cmd,
		},
		Image:  b.image,
		Labels: map[string]string{},
		User:   u.Uid + ":user",
	}

	repoVol, err := ensureVolumeForRepo(ctx, b.repo)
	if err != nil {
		return nil, err
	}

	mounts := []mount.Mount{
		{
			Type:   mount.TypeVolume,
			Source: repoVol.vol.Name,
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

	_, err = cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, name)
	if err != nil {
		return nil, xerrors.Errorf("failed to create container: %w", err)
	}

	cnt, err := cli.ContainerInspect(ctx, name)
	if err != nil {
		return nil, xerrors.Errorf("failed to inspect container after create: %w", err)
	}

	env := &Environment{
		repo: b.repo,
		name: name,
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

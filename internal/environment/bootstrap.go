package environment

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/mount"
	"go.coder.com/flog"
	"golang.org/x/xerrors"
)

// Bootstrap will attempt to create an environment using the docker file
// located at '.sail/Dockerfile' within the repo.
//
// An initial default environment and volume will be created, then the repo
// will be cloned into the volume. If the sail docker file does exists, an
// image will be created and the environment will be rebuilt using that new
// image.
//
// The default environment will be returned if no sail docker file exists.
func Bootstrap(ctx context.Context, cfg *BuildConfig, repo *Repo) (*Environment, error) {
	// TODO: Should this always try to create?
	lv, err := ensureVolumeForRepo(ctx, repo)
	if err != nil {
		return nil, err
	}

	projectPath := defaultDirForRepo(repo)
	projectMount := mount.Mount{
		Type:   mount.TypeVolume,
		Source: lv.vol.Name,
		Target: projectPath,
	}
	cfg.Mounts = append(cfg.Mounts, projectMount)

	env, err := Build(ctx, cfg)
	if err != nil {
		return nil, err
	}

	err = cloneInto(ctx, env, repo, projectPath)
	if err != nil {
		return nil, err
	}

	sailPath := filepath.Join(projectPath, ".sail")
	_, err = env.readPath(ctx, sailPath)
	if xerrors.Is(err, errNoSuchFile) {
		flog.Info("no '.sail/Dockerfile' for repo")
		return env, nil
	}
	if err != nil {
		return nil, err
	}

	prov := &EnvPathProvider{
		Env:  env,
		Path: sailPath,
	}
	env, err = Modify(ctx, prov, env)
	if err != nil {
		return nil, xerrors.Errorf("failed to modify: %w", err)
	}

	return env, nil
}

// ensureVolumeForRepo ensures that there's a volume dedicated to storing the
// repo.
func ensureVolumeForRepo(ctx context.Context, r *Repo) (*localVolume, error) {
	name := r.DockerName()
	lv, err := findLocalVolume(ctx, name)
	if xerrors.Is(err, errMissingVolume) {
		lv, err = createLocalVolume(ctx, name)
	}
	if err != nil {
		return nil, err
	}

	return lv, nil
}

// cloneInto clones the repo into the environment.
//
// The enviroment should have its auth set up in such a way that would allow the
// user to clone a private repo.
//
// TODO: Should this handle creating/attaching the volume?
func cloneInto(ctx context.Context, env *Environment, repo *Repo, path string) error {
	out, err := env.exec(ctx, "sudo", "mkdir", "-p", path).CombinedOutput()
	if err != nil {
		return xerrors.Errorf("failed to create dir: %s: %w", out, err)
	}

	out, err = env.exec(ctx, "sudo", "chown", "-R", "user", path).CombinedOutput()
	if err != nil {
		return xerrors.Errorf("failed to chown: %s: %w", out, err)
	}

	uri := repo.CloneURI()
	flog.Info("cloning from %s", uri)
	cloneStr := fmt.Sprintf("cd %s; git clone %s .", path, uri)
	cmd := env.execTTY(ctx, "bash", []string{"-c", cloneStr}...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return xerrors.Errorf("failed to clone: %w", err)
	}

	return nil
}

var errMissingDockerfile = xerrors.Errorf("missing dockerfile")

// buildImage builds an image from the given directory. The dir should contain a
// valid dockerfile at its root.
func buildImage(ctx context.Context, buildContext io.Reader, name string) error {
	cli := dockerClient()
	defer cli.Close()

	flog.Info("building image: %s", name)

	opts := types.ImageBuildOptions{
		Tags: []string{name},
	}
	resp, err := cli.ImageBuild(ctx, buildContext, opts)
	if err != nil {
		return xerrors.Errorf("failed to build image: %w", err)
	}
	defer resp.Body.Close()

	// Mostly for debugging.
	io.Copy(os.Stdout, resp.Body)

	return nil
}

func defaultDirForRepo(r *Repo) string {
	return filepath.Join("/home/user/Projects", r.BaseName())
}

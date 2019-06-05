package environment

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"go.coder.com/flog"
	"go.coder.com/sail/internal/randstr"
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
	env, err := Build(ctx, cfg)
	if err != nil {
		return nil, err
	}

	err = cloneInto(ctx, env, repo, "/home/user/Projects/project")
	if err != nil {
		return nil, err
	}

	// TODO: Actually do the other stuff too.
	return env, nil

	imgName, err := imageFromEnvRepo(ctx, env)
	if xerrors.Is(err, errMissingDockerfile) {
		flog.Info("no '.sail/Dockerfile' for repo")
		return env, nil
	}
	if err != nil {
		return nil, err
	}

	flog.Info("created new image from '.sail/Dockerfile': %s", imgName)

	// New image for environment was created, stop and remove the previous
	// container to make room for the new one.
	err = Stop(ctx, env)
	if err != nil {
		return nil, err
	}
	err = Remove(ctx, env)
	if err != nil {
		return nil, err
	}

	// Set new image to build with.
	cfg.Image = imgName

	// Rebuild with image specific to the repo.
	env, err = Build(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return env, nil
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

	cloneStr := fmt.Sprintf("cd %s; git clone %s .", path, repo.CloneURI())
	out, err = env.exec(ctx, "bash", []string{"-c", cloneStr}...).CombinedOutput()
	if err != nil {
		return xerrors.Errorf("failed to clone: %s: %w", out, err)
	}

	return nil
}

var errMissingDockerfile = xerrors.Errorf("missing dockerfile")

// imageFromEnvRepo will attempt to create an image from repo inside the
// environment if the repo contains a docker file at '.sail/Dockerfile'.
// errMissingDockerfile will be returned if a docker file cannot be found.
func imageFromEnvRepo(ctx context.Context, env *Environment) (string, error) {
	const relPath = ".sail"

	// Read the file from the docker container and not directly from the path
	// on the host machine. The path on the host machine is owned by root.
	// TODO: Fix repo path.
	rdr, err := env.readPath(ctx, filepath.Join(defaultDirForRepo(nil), relPath))
	if xerrors.Is(err, errNoSuchFile) {
		return "", errMissingDockerfile
	}
	if err != nil {
		return "", err
	}

	// TODO: Better image name
	imgName := env.name + "-bootstrap-" + randstr.MakeCharset(randstr.Lower, 5)
	err = buildImage(ctx, rdr, imgName)
	if err != nil {
		return "", err
	}

	return imgName, nil
}

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

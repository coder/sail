package environment

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
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
func Bootstrap(ctx context.Context, b *Builder) (*Environment, error) {
	env, err := b.Build(ctx)
	if err != nil {
		return nil, err
	}

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

	// Update builder to not clone repo since we already have a volume
	// containing the repo.
	// b.skipClone = true
	// Set new image to build with.
	b.image = imgName

	// Rebuild with image specific to the repo.
	env, err = b.Build(ctx)
	if err != nil {
		return nil, err
	}

	return env, nil
}

var errMissingDockerfile = xerrors.Errorf("missing dockerfile")

// imageFromEnvRepo will attempt to create an image from repo inside the
// environment if the repo contains a docker file at '.sail/Dockerfile'.
// errMissingDockerfile will be returned if a docker file cannot be found.
func imageFromEnvRepo(ctx context.Context, env *Environment) (string, error) {
	const relPath = ".sail"

	// Read the file from the docker container and not directly from the path
	// on the host machine. The path on the host machine is owned by root.
	rdr, err := env.readPath(ctx, filepath.Join(defaultDirForRepo(env.repo), relPath))
	if xerrors.Is(err, errNoSuchFile) {
		return "", errMissingDockerfile
	}
	if err != nil {
		return "", err
	}

	imgName := env.repo.DockerName()
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

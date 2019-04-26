package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/client"
	"golang.org/x/xerrors"

	"go.coder.com/flog"
	"go.coder.com/sail/internal/hat"
	"go.coder.com/sail/internal/xexec"
)

// hatBuilder is responsible for applying a hat to a base image.
// The hatBuilder passes any sail labels to the runner through
// setting them in the image. Besides this, the runner should
// have no knowledge of the hatBuilder existing.
type hatBuilder struct {
	// hatPath is the path to the hat file.
	hatPath string
	// baseImage is the image before the hat is applied.
	baseImage string
}

// dockerClient returns an instantiated docker client that
// is using the correct API version. If the client can't be
// constructed, this will panic.
func dockerClient() *client.Client {
	cli, err := client.NewEnvClient()
	if err != nil {
		panicf("failed to make docker client: %v", err)
	}

	// Update the API version of the client to match
	// what the server is running.
	cli.NegotiateAPIVersion(context.Background())

	return cli
}

func (b *hatBuilder) resolveHatPath() (string, error) {
	const ghPrefix = "github:"

	hatPath := b.hatPath
	if strings.HasPrefix(b.hatPath, ghPrefix) {
		hatPath = strings.TrimLeft(b.hatPath, ghPrefix)
		return hat.ResolveGitHubPath(hatPath)
	}

	return hatPath, nil
}

// applyHat applies the hat to the base image.
func (b *hatBuilder) applyHat() (string, error) {
	if b.hatPath == "" {
		return "", xerrors.New("unable to apply hat, none specified")
	}

	hatPath, err := b.resolveHatPath()
	if err != nil {
		return "", xerrors.Errorf("failed to resolve hat path: %w", err)
	}

	dockerFilePath := filepath.Join(hatPath, "Dockerfile")

	dockerFileByt, err := ioutil.ReadFile(dockerFilePath)
	if err != nil {
		return "", xerrors.Errorf("failed to read %v: %w", dockerFilePath, err)
	}
	dockerFileByt = hat.DockerReplaceFrom(dockerFileByt, b.baseImage)

	fi, err := ioutil.TempFile("", "hat")
	if err != nil {
		return "", xerrors.Errorf("failed to create temp file: %w", err)
	}
	defer fi.Close()
	defer os.Remove(fi.Name())

	_, err = fi.Write(dockerFileByt)
	if err != nil {
		return "", xerrors.Errorf("failed to write to %v: %w", fi.Name(), err)
	}

	// We tag based on the checksum of the Dockerfile to avoid spamming
	// images.
	csm := sha256.Sum256(dockerFileByt)
	imageName := b.baseImage + "-hat-" + hex.EncodeToString(csm[:])[:16]

	flog.Info("building hat image %v", imageName)
	cmd := xexec.Fmt("docker build --network=host -t %v -f %v %v --label %v=%v --label %v=%v",
		imageName, fi.Name(), hatPath, baseImageLabel, b.baseImage, hatLabel, b.hatPath,
	)
	xexec.Attach(cmd)
	err = cmd.Run()
	if err != nil {
		return "", xerrors.Errorf("failed to build hatted baseImage: %w", err)
	}

	return imageName, nil
}

// hatBuilderFromContainer gets a hatBuilder from container named
// name.
func hatBuilderFromContainer(name string) (*hatBuilder, error) {
	cli := dockerClient()
	defer cli.Close()

	cnt, err := cli.ContainerInspect(context.Background(), name)
	if err != nil {
		return nil, xerrors.Errorf("failed to inspect %v: %w", name, err)
	}

	return &hatBuilder{
		baseImage: cnt.Config.Labels[baseImageLabel],
		hatPath:   cnt.Config.Labels[hatLabel],
	}, nil
}

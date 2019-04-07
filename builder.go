package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"go.coder.com/flog"
	"go.coder.com/narwhal/internal/hat"
	"go.coder.com/narwhal/internal/xexec"
	"golang.org/x/xerrors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Docker labels for Narwhal state.
const (
	baseImageLabel  = narwhalLabel + ".base_image"
	hatLabel        = narwhalLabel + ".hat"
	portLabel       = narwhalLabel + ".port"
	projectDirLabel = narwhalLabel + ".project_dir"
)

// builder holds all the information needed to assemble a new narwhal container.
// The builder stores itself as state on the container.
// It enables quick iteration on a container with small modifications to it's config.
type builder struct {
	name string

	shares []types.MountPoint

	// hatPath is the path to the hat file.
	hatPath string
	// baseImage is the image before the hat is applied.
	baseImage string
	hostname  string

	port       string
	projectDir string

	// hostUser is the uid on the host which is mapped to
	// the container's "user" user.
	hostUser string

	testCmd string
}

func dockerClient() *client.Client {
	cli, err := client.NewEnvClient()
	if err != nil {
		flog.Fatal("failed to make docker client: %v", err)
	}
	return cli
}

func (b *builder) applyHat() string {
	const ghPrefix = "github:"

	hatPath := b.hatPath
	if strings.HasPrefix(b.hatPath, ghPrefix) {
		hatPath = strings.TrimLeft(b.hatPath, ghPrefix)
		hatPath = hat.ResolveGitHubPath(b.hatPath)
	}

	dockerFilePath := filepath.Join(hatPath, "Dockerfile")

	dockerFileByt, err := ioutil.ReadFile(dockerFilePath)
	if err != nil {
		flog.Fatal("failed to read %v: %v", dockerFilePath, err)
	}
	dockerFileByt = hat.DockerReplaceFrom(dockerFileByt, b.baseImage)

	fi, err := ioutil.TempFile("", "hat")
	if err != nil {
		flog.Fatal("failed to create temp file: %v", err)
	}
	defer fi.Close()
	defer os.Remove(fi.Name())

	_, err = fi.Write(dockerFileByt)
	if err != nil {
		flog.Fatal("failed to write to %v: %v", fi.Name(), err)
	}

	// We tag based on the checksum of the Dockerfile to avoid spamming
	// images.
	csm := sha256.Sum256(dockerFileByt)
	imageName := b.baseImage + "-hat-" + hex.EncodeToString(csm[:])[:16]

	flog.Info("building hat image %v", imageName)
	cmd := xexec.Fmt("docker build --network=host -t %v -f %v %v",
		imageName, fi.Name(), hatPath,
	)
	xexec.Attach(cmd)
	err = cmd.Run()
	if err != nil {
		flog.Fatal("failed to build hatted baseImage: %v", err)
	}
	return imageName
}

// imageMounts adds a list of shares to the shares map from the image.
func (b *builder) imageMounts(image string, mounts []mount.Mount) []mount.Mount {
	cli, err := client.NewEnvClient()
	if err != nil {
		flog.Fatal("failed to create docker client: %v", err)
	}

	ins, _, err := cli.ImageInspectWithRaw(context.Background(), b.baseImage)
	if err != nil {
		flog.Fatal("failed to inspect %v: %v", b.baseImage, err)
	}

	hostHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	for k, v := range ins.ContainerConfig.Labels {
		const prefix = "share."
		if !strings.HasPrefix(k, prefix) {
			continue
		}

		tokens := strings.Split(v, ":")
		if len(tokens) != 2 {
			flog.Fatal("invalid share %q", v)
		}

		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: resolvePath(hostHomeDir, tokens[0]),
			Target: resolvePath(guestHomeDir, tokens[1]),
		})
	}
	return mounts
}

func (b *builder) stripDuplicateMounts(mounts []mount.Mount) []mount.Mount {
	rmounts := make([]mount.Mount, 0, len(mounts))

	dests := make(map[string]struct{})
	for _, mnt := range mounts {
		if _, ok := dests[mnt.Target]; ok {
			continue
		}
		dests[mnt.Target] = struct{}{}
		rmounts = append(rmounts, mnt)
	}
	return rmounts
}

// runContainer creates and runs a new container.
// It handles installing code-server, and uses code-server as
// the container's root process.
// We want code-server to be the root process as it gives us the nice guarantee that
// the container is only online when code-server is working.
func (b *builder) runContainer() error {
	cli := dockerClient()

	image := b.baseImage
	if b.hatPath != "" {
		image = b.applyHat()
	}

	var mounts []mount.Mount
	for _, sh := range b.shares {
		mounts = append(mounts, mount.Mount{
			Type:   mount.Type(sh.Type),
			Source: sh.Source,
			Target: sh.Destination,
		})
	}

	// Mount in code-server
	codeServerBinPath, err := loadCodeServer(context.Background())
	if err != nil {
		return xerrors.Errorf("failed to load code-server: %w", err)
	}
	mounts = append(mounts, mount.Mount{
		Type:   mount.TypeBind,
		Source: codeServerBinPath,
		Target: "/usr/bin/code-server",
	})

	// We take the mounts from the final image so that it includes the hat and the baseImage.
	mounts = b.imageMounts(image, mounts)

	// Duplicates can arise from trying to rebuild an existing image.
	mounts = b.stripDuplicateMounts(mounts)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// We want the code-server logs to be available inside the container for easy
	// access during development, but also going to stdout so `docker logs` can be used
	// to debug a failed code-server startup.
	cmd := "cd " + b.projectDir + "; code-server --port " + b.port + " --data-dir ~/.config/Code --extensions-dir ~/.vscode/extensions --allow-http --no-auth 2>&1 | tee " + containerLogPath
	if b.testCmd != "" {
		cmd = b.testCmd + "; exit 1"
	}

	_, err = cli.ContainerCreate(ctx, &container.Config{
		Hostname: b.hostname,
		Cmd: strslice.StrSlice{
			"bash", "-c", cmd,
		},
		Image: image,
		Labels: map[string]string{
			narwhalLabel:    "",
			hatLabel:        b.hatPath,
			baseImageLabel:  b.baseImage,
			portLabel:       b.port,
			projectDirLabel: b.projectDir,
		},
		User: b.hostUser + ":user",
	}, &container.HostConfig{
		Mounts:      mounts,
		NetworkMode: "host",
		Privileged:  true,
	}, nil, b.name)
	if err != nil {
		return xerrors.Errorf("failed to create container: %w", err)
	}

	err = cli.ContainerStart(ctx, b.name, types.ContainerStartOptions{})
	if err != nil {
		return xerrors.Errorf("failed to start container: %w", err)
	}

	return nil
}

// builderFromContainer gets a builder config from container named
// name.
func builderFromContainer(name string) *builder {
	cli := dockerClient()

	cnt, err := cli.ContainerInspect(context.Background(), name)
	if err != nil {
		flog.Fatal("failed to inspect %v: %v", name, err)
	}

	return &builder{
		name:       name,
		shares:     cnt.Mounts,
		hostname:   cnt.Config.Hostname,
		baseImage:  cnt.Config.Labels[baseImageLabel],
		hatPath:    cnt.Config.Labels[hatLabel],
		port:       cnt.Config.Labels[portLabel],
		projectDir: cnt.Config.Labels[projectDirLabel],
		hostUser:   cnt.Config.User,
	}
}

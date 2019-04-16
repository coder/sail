package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"go.coder.com/flog"
	"go.coder.com/narwhal/internal/hat"
	"go.coder.com/narwhal/internal/xexec"
	"golang.org/x/xerrors"
)

// Docker labels for Narwhal state.
const (
	baseImageLabel       = narwhalLabel + ".base_image"
	hatLabel             = narwhalLabel + ".hat"
	portLabel            = narwhalLabel + ".port"
	projectLocalDirLabel = narwhalLabel + ".project_local_dir"
	projectDirLabel      = narwhalLabel + ".project_dir"
	projectNameLabel     = narwhalLabel + ".project_name"
)

// builder holds all the information needed to assemble a new narwhal container.
// The builder stores itself as state on the container.
// It enables quick iteration on a container with small modifications to it's config.
// All mounts should be configured from the image.
type builder struct {
	cntName     string
	projectName string

	// hatPath is the path to the hat file.
	hatPath string
	// baseImage is the image before the hat is applied.
	baseImage string
	hostname  string

	port string

	projectLocalDir string

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

	// Update the API version of the client to match
	// what the server is running.
	cli.NegotiateAPIVersion(context.Background())

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

func (b *builder) projectDir(image string) (string, error) {
	cli := dockerClient()
	defer cli.Close()

	img, _, err := cli.ImageInspectWithRaw(context.Background(), image)
	if err != nil {
		return "", xerrors.Errorf("failed to inspect image: %w", err)
	}

	proot, ok := img.Config.Labels["project_root"]
	if ok {
		return filepath.Join(proot, b.projectName), nil
	}

	return filepath.Join(guestHomeDir, b.projectName), nil
}

// imageDefinedMounts adds a list of shares to the shares map from the image.
func (b *builder) imageDefinedMounts(image string, mounts []mount.Mount) []mount.Mount {
	cli := dockerClient()
	defer cli.Close()

	ins, _, err := cli.ImageInspectWithRaw(context.Background(), b.baseImage)
	if err != nil {
		flog.Fatal("failed to inspect %v: %v", b.baseImage, err)
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
			Source: tokens[0],
			Target: tokens[1],
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

func panicf(fmtStr string, args ...interface{}) {
	panic(fmt.Sprintf(fmtStr, args...))
}

// resolveMounts replaces ~ with appropriate home paths with
// each mount.
func (b *builder) resolveMounts(mounts []mount.Mount) {
	hostHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	for i := range mounts {
		mounts[i].Source, err = filepath.Abs(resolvePath(hostHomeDir, mounts[i].Source))
		if err != nil {
			panicf("failed to resolve %v: %v", mounts[i].Source, err)
		}
		mounts[i].Target = resolvePath(guestHomeDir, mounts[i].Target)
	}
}

func (b *builder) mounts(mounts []mount.Mount, image string) ([]mount.Mount, error) {
	// Mount in VS Code configs.
	mounts = append(mounts, mount.Mount{
		Type:   "bind",
		Source: "~/.config/Code",
		Target: "~/.config/Code",
	})
	mounts = append(mounts, mount.Mount{
		Type:   "bind",
		Source: "~/.vscode/extensions",
		Target: "~/.vscode/extensions",
	})

	projectDir, err := b.projectDir(image)
	if err != nil {
		return nil, err
	}

	mounts = append(mounts, mount.Mount{
		Type:   "bind",
		Source: b.projectLocalDir,
		Target: projectDir,
	})

	// Mount in code-server
	codeServerBinPath, err := loadCodeServer(context.Background())
	if err != nil {
		return nil, xerrors.Errorf("failed to load code-server: %w", err)
	}
	mounts = append(mounts, mount.Mount{
		Type:   mount.TypeBind,
		Source: codeServerBinPath,
		Target: "/usr/bin/code-server",
	})

	// We take the mounts from the final image so that it includes the hat and the baseImage.
	mounts = b.imageDefinedMounts(image, mounts)

	b.resolveMounts(mounts)
	return mounts, nil
}

// runContainer creates and runs a new container.
// It handles installing code-server, and uses code-server as
// the container's root process.
// We want code-server to be the root process as it gives us the nice guarantee that
// the container is only online when code-server is working.
func (b *builder) runContainer() error {
	cli := dockerClient()
	defer cli.Close()

	image := b.baseImage
	if b.hatPath != "" {
		image = b.applyHat()
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	var mounts []mount.Mount

	mounts, err := b.mounts(mounts, image)

	projectDir, err := b.projectDir(image)
	if err != nil {
		return err
	}

	// We want the code-server logs to be available inside the container for easy
	// access during development, but also going to stdout so `docker logs` can be used
	// to debug a failed code-server startup.
	cmd := "cd " + projectDir + "; code-server --port " + b.port + " --data-dir ~/.config/Code --extensions-dir ~/.vscode/extensions --allow-http --no-auth 2>&1 | tee " + containerLogPath
	if b.testCmd != "" {
		cmd = b.testCmd + "; exit 1"
	}

	containerConfig := &container.Config{
		Hostname: b.hostname,
		Cmd: strslice.StrSlice{
			"bash", "-c", cmd,
		},
		Image: image,
		Labels: map[string]string{
			narwhalLabel:         "",
			hatLabel:             b.hatPath,
			baseImageLabel:       b.baseImage,
			portLabel:            b.port,
			projectDirLabel:      projectDir,
			projectLocalDirLabel: b.projectLocalDir,
			projectNameLabel:     b.projectName,
		},
		User: b.hostUser + ":user",
	}

	hostConfig := &container.HostConfig{
		Mounts:      mounts,
		NetworkMode: "host",
		Privileged:  true,
		ExtraHosts: []string{
			b.hostname + ":127.0.0.1",
		},
	}

	_, err = cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, b.cntName)
	if err != nil {
		return xerrors.Errorf("failed to create container: %w\n%s\n%s",
			err,
			spew.Sdump(containerConfig),
			spew.Sdump(hostConfig),
		)
	}

	err = cli.ContainerStart(ctx, b.cntName, types.ContainerStartOptions{})
	if err != nil {
		return xerrors.Errorf("failed to start container: %w", err)
	}

	return nil
}

// builderFromContainer gets a builder config from container named
// name.
func builderFromContainer(name string) *builder {
	cli := dockerClient()
	defer cli.Close()

	cnt, err := cli.ContainerInspect(context.Background(), name)
	if err != nil {
		flog.Fatal("failed to inspect %v: %v", name, err)
	}

	return &builder{
		cntName:         name,
		hostname:        cnt.Config.Hostname,
		baseImage:       cnt.Config.Labels[baseImageLabel],
		hatPath:         cnt.Config.Labels[hatLabel],
		port:            cnt.Config.Labels[portLabel],
		projectLocalDir: cnt.Config.Labels[projectLocalDirLabel],
		projectName:     cnt.Config.Labels[projectNameLabel],
		hostUser:        cnt.Config.User,
	}
}

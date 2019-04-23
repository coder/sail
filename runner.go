package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"golang.org/x/xerrors"
)

// containerLogPath is the location of the code-server log.
const containerLogPath = "/tmp/code-server.log"

// Docker labels for sail state.
const (
	sailLabel = "com.coder.sail"

	baseImageLabel       = sailLabel + ".base_image"
	hatLabel             = sailLabel + ".hat"
	projectLocalDirLabel = sailLabel + ".project_local_dir"
	projectDirLabel      = sailLabel + ".project_dir"
	projectNameLabel     = sailLabel + ".project_name"
)

// runner holds all the information needed to assemble a new sail container.
// The runner stores itself as state on the container.
// It enables quick iteration on a container with small modifications to it's config.
// All mounts should be configured from the image.
type runner struct {
	cntName     string
	projectName string

	hostname string

	port string

	projectLocalDir string

	// hostUser is the uid on the host which is mapped to
	// the container's "user" user.
	hostUser string

	testCmd string
}

// runContainer creates and runs a new container.
// It handles installing code-server, and uses code-server as
// the container's root process.
// We want code-server to be the root process as it gives us the nice guarantee that
// the container is only online when code-server is working.
func (r *runner) runContainer(image string) error {
	cli := dockerClient()
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	var (
		err    error
		mounts []mount.Mount
	)

	mounts, err = r.mounts(mounts, image)
	if err != nil {
		return xerrors.Errorf("failed to assemble mounts: %w", err)
	}

	projectDir, err := r.projectDir(image)
	if err != nil {
		return err
	}

	// We want the code-server logs to be available inside the container for easy
	// access during development, but also going to stdout so `docker logs` can be used
	// to debug a failed code-server startup.
	cmd := "cd " + projectDir +
		"; code-server --host 127.0.0.1" +
		" --port " + r.port +
		" --data-dir ~/.config/Code --extensions-dir ~/.vscode/extensions --allow-http --no-auth 2>&1 | tee " + containerLogPath
	if r.testCmd != "" {
		cmd = r.testCmd + "; exit 1"
	}

	var envs []string
	sshAuthSock, exists := os.LookupEnv("SSH_AUTH_SOCK")
	if exists {
		s := fmt.Sprintf("SSH_AUTH_SOCK=%s", sshAuthSock)
		envs = append(envs, s)
	}

	containerConfig := &container.Config{
		Hostname: r.hostname,
		Env:      envs,
		Cmd: strslice.StrSlice{
			"bash", "-c", cmd,
		},
		Image: image,
		Labels: map[string]string{
			sailLabel:            "",
			projectDirLabel:      projectDir,
			projectLocalDirLabel: r.projectLocalDir,
			projectNameLabel:     r.projectName,
		},
		User: r.hostUser + ":user",
	}

	err = r.addImageDefinedLabels(image, containerConfig.Labels)
	if err != nil {
		return xerrors.Errorf("failed to add image defined labels: %w", err)
	}

	hostConfig := &container.HostConfig{
		Mounts:      mounts,
		NetworkMode: "host",
		Privileged:  true,
		ExtraHosts: []string{
			r.hostname + ":127.0.0.1",
		},
	}

	_, err = cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, r.cntName)
	if err != nil {
		return xerrors.Errorf("failed to create container: %w", err)
	}

	err = cli.ContainerStart(ctx, r.cntName, types.ContainerStartOptions{})
	if err != nil {
		return xerrors.Errorf("failed to start container: %w", err)
	}

	return nil
}

func (r *runner) mounts(mounts []mount.Mount, image string) ([]mount.Mount, error) {
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

	// 'SSH_AUTH_SOCK' is provided by a running ssh-agent. Passing in the
	// socket to the container allows for using the user's existing setup for
	// ssh authentication instead of having to create a new keys or explicity
	// pass them in.
	sshAuthSock, exists := os.LookupEnv("SSH_AUTH_SOCK")
	if exists {
		mounts = append(mounts, mount.Mount{
			Type:   "bind",
			Source: sshAuthSock,
			Target: sshAuthSock,
		})
	}

	localGlobalStorageDir := filepath.Join(metaRoot(), r.cntName, "globalStorage")
	err := os.MkdirAll(localGlobalStorageDir, 0750)
	if err != nil {
		return nil, err
	}

	// globalStorage holds the UI state, and other code-server specific
	// state.
	mounts = append(mounts, mount.Mount{
		Type:   "bind",
		Source: localGlobalStorageDir,
		Target: "~/.local/share/code-server/globalStorage/",
	})

	projectDir, err := r.projectDir(image)
	if err != nil {
		return nil, err
	}

	mounts = append(mounts, mount.Mount{
		Type:   "bind",
		Source: r.projectLocalDir,
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
	mounts, err = r.imageDefinedMounts(image, mounts)
	if err != nil {
		return nil, err
	}

	r.resolveMounts(mounts)

	err = r.ensureMountSources(mounts)
	if err != nil {
		return nil, err
	}

	return mounts, nil
}

// ensureMountSources ensures that the mount's source exists. If the source
// doesn't exist, it will be created as a directory on the host.
func (r *runner) ensureMountSources(mounts []mount.Mount) error {
	for _, mount := range mounts {
		_, err := os.Stat(mount.Source)
		if err == nil {
			continue
		}
		if !os.IsNotExist(err) {
			return xerrors.Errorf("failed to stat mount source %v: %w", mount.Source, err)
		}

		err = os.MkdirAll(mount.Source, 0755)
		if err != nil {
			return xerrors.Errorf("failed to create mount source %v: %w", mount.Source, err)
		}
	}

	return nil
}

// imageDefinedMounts adds a list of shares to the shares map from the image.
func (r *runner) imageDefinedMounts(image string, mounts []mount.Mount) ([]mount.Mount, error) {
	cli := dockerClient()
	defer cli.Close()

	ins, _, err := cli.ImageInspectWithRaw(context.Background(), image)
	if err != nil {
		return nil, xerrors.Errorf("failed to inspect %v: %w", image, err)
	}

	for k, v := range ins.ContainerConfig.Labels {
		const prefix = "share."
		if !strings.HasPrefix(k, prefix) {
			continue
		}

		tokens := strings.Split(v, ":")
		if len(tokens) != 2 {
			return nil, xerrors.Errorf("invalid share %q", v)
		}

		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: tokens[0],
			Target: tokens[1],
		})
	}
	return mounts, nil
}

// addImageDefinedLabels adds any sail labels that were defined on the image onto the container.
func (r *runner) addImageDefinedLabels(image string, labels map[string]string) error {
	cli := dockerClient()
	defer cli.Close()

	ins, _, err := cli.ImageInspectWithRaw(context.Background(), image)
	if err != nil {
		return xerrors.Errorf("failed to inspect %v: %w", image, err)
	}

	for k, v := range ins.ContainerConfig.Labels {
		if !strings.HasPrefix(k, sailLabel) {
			continue
		}

		labels[k] = v
	}

	return nil
}

func (r *runner) stripDuplicateMounts(mounts []mount.Mount) []mount.Mount {
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
func (r *runner) resolveMounts(mounts []mount.Mount) {
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

func (r *runner) projectDir(image string) (string, error) {
	cli := dockerClient()
	defer cli.Close()

	img, _, err := cli.ImageInspectWithRaw(context.Background(), image)
	if err != nil {
		return "", xerrors.Errorf("failed to inspect image: %w", err)
	}

	proot, ok := img.Config.Labels["project_root"]
	if ok {
		return filepath.Join(proot, r.projectName), nil
	}

	return filepath.Join(guestHomeDir, r.projectName), nil
}

// runnerFromContainer gets a runner from container named
// name.
func runnerFromContainer(name string) (*runner, error) {
	cli := dockerClient()
	defer cli.Close()

	cnt, err := cli.ContainerInspect(context.Background(), name)
	if err != nil {
		return nil, xerrors.Errorf("failed to inspect %v: %w", name, err)
	}

	port, err := codeServerPort(name)
	if err != nil {
		return nil, xerrors.Errorf("failed to find code server port: %w", err)
	}

	return &runner{
		cntName:         name,
		hostname:        cnt.Config.Hostname,
		port:            port,
		projectLocalDir: cnt.Config.Labels[projectLocalDirLabel],
		projectName:     cnt.Config.Labels[projectNameLabel],
		hostUser:        cnt.Config.User,
	}, nil
}

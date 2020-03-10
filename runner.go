package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"
	"golang.org/x/xerrors"

	"go.coder.com/flog"
	"go.coder.com/sail/internal/dockutil"
)

// containerLogPath is the location of the code-server log.
const containerLogPath = "/tmp/code-server.log"

// containerHome is the location of the user's home directory
// inside of the container. This is only used in places where
// docker won't expand the `~` path or the `$HOME` variable.
// For example, when setting environment variables for the container.
const containerHome = "/home/user"

// Docker labels for sail state.
const (
	sailLabel = "com.coder.sail"

	baseImageLabel       = sailLabel + ".base_image"
	hatLabel             = sailLabel + ".hat"
	projectLocalDirLabel = sailLabel + ".project_local_dir"
	projectDirLabel      = sailLabel + ".project_dir"
	projectNameLabel     = sailLabel + ".project_name"
	proxyURLLabel        = sailLabel + ".proxy_url"
)

// Docker labels for user configuration.
const (
	onStartLabel     = "on_start"
	projectRootLabel = "project_root"
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

	testCmd string

	proxyURL string
}

// runContainer creates and runs a new container.
// It handles installing code-server, and uses code-server as
// the container's root process.
// We want code-server to be the root process as it gives us the nice guarantee that
// the container is only online when code-server is working.
// Additionally, runContainer also runs the image's `on_start` label as a bash
// command inside of the project directory.
func (r *runner) runContainer(image string) error {
	cli := dockerClient()
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	projectDir, err := r.projectDir(image)
	if err != nil {
		return err
	}

	var envs []string
	envs = r.environment(envs)

	containerConfig := &container.Config{
		Hostname: r.hostname,
		Env:      envs,
		Cmd: strslice.StrSlice{
			"bash", "-c", r.constructCommand(projectDir),
		},
		Image: image,
		Labels: map[string]string{
			sailLabel:            "",
			projectDirLabel:      projectDir,
			projectLocalDirLabel: r.projectLocalDir,
			projectNameLabel:     r.projectName,
			proxyURLLabel:        r.proxyURL,
		},
		// The user inside has uid 1000. This works even on macOS where the default user has uid 501.
		// See https://stackoverflow.com/questions/43097341/docker-on-macosx-does-not-translate-file-ownership-correctly-in-volumes
		// The docker image runs it as uid 1000 so we don't need to set anything.
		User: "",
	}

	err = r.addImageDefinedLabels(image, containerConfig.Labels)
	if err != nil {
		return xerrors.Errorf("failed to add image defined labels: %w", err)
	}

	var mounts []mount.Mount
	mounts = r.addHatMount(mounts, containerConfig.Labels)

	mounts, err = r.mounts(mounts, image)
	if err != nil {
		return xerrors.Errorf("failed to assemble mounts: %w", err)
	}

	hostConfig, err := r.hostConfig(containerConfig, mounts)
	if err != nil {
		return err
	}

	_, err = cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, r.cntName)
	if err != nil {
		return xerrors.Errorf("failed to create container: %w", err)
	}

	err = cli.ContainerStart(ctx, r.cntName, types.ContainerStartOptions{})
	if err != nil {
		return xerrors.Errorf("failed to start container: %w", err)
	}

	err = r.runOnStart(image)
	if err != nil {
		return xerrors.Errorf("failed to run on_start label in container: %w", err)
	}

	return nil
}

// constructCommand constructs the code-server command that will be used
// as the Sail container's init process.
func (r *runner) constructCommand(projectDir string) string {
	containerAddr := "localhost"
	containerPort := r.port
	if runtime.GOOS == "darwin" {
		// See justification in `runner.hostConfig`.
		containerPort = "8443"
		containerAddr = "0.0.0.0"
	}

	// We want the code-server logs to be available inside the container for easy
	// access during development, but also going to stdout so `docker logs` can be used
	// to debug a failed code-server startup.
	//
	// We start code-server such that extensions installed through the UI are placed in the host's extension dir.
	cmd := fmt.Sprintf(`set -euxo pipefail || exit 1
	cd %v
	# This is necessary in case the .vscode directory wasn't created inside the container, as mounting to the host
	# extension dir will create it as root.
	sudo chown user:user ~/.vscode
	/usr/bin/code-server --host %v --port %v --user-data-dir ~/.config/Code --extensions-dir %v --extra-extensions-dir ~/.vscode/extensions --auth=none \
	--allow-http 2>&1 | tee %v
	`, projectDir, containerAddr, containerPort, hostExtensionsDir, containerLogPath)

	if r.testCmd != "" {
		cmd = r.testCmd + "\n exit 1"
	}

	return cmd
}

// hostConfig constructs the container.HostConfig required for starting the sail container.
func (r *runner) hostConfig(containerConfig *container.Config, mounts []mount.Mount) (*container.HostConfig, error) {
	hostConfig := &container.HostConfig{
		Mounts:      mounts,
		NetworkMode: "host",
		Privileged:  true,
		ExtraHosts: []string{
			r.hostname + ":127.0.0.1",
		},
	}

	// macOS does not support host networking.
	// See https://github.com/docker/for-mac/issues/2716
	if runtime.GOOS == "darwin" {
		portSpec := fmt.Sprintf("127.0.0.1:%v:%v/tcp", r.port, "8443")
		hostConfig.NetworkMode = ""
		exposed, bindings, err := nat.ParsePortSpecs([]string{portSpec})
		if err != nil {
			return nil, xerrors.Errorf("failed to parse port spec: %w", err)
		}
		containerConfig.ExposedPorts = exposed
		hostConfig.PortBindings = bindings
	}

	return hostConfig, nil
}

// environment sets any environment variables that may need to be set inside
// the container.
func (r *runner) environment(envs []string) []string {
	sshAuthSock, exists := os.LookupEnv("SSH_AUTH_SOCK")
	if exists {
		s := fmt.Sprintf("SSH_AUTH_SOCK=%s", sshAuthSock)
		envs = append(envs, s)
	}

	if runtime.GOOS == "linux" {
		// When on linux and the display variable exists we forward it so
		// that GUI applications can run.
		if os.Getenv("DISPLAY") != "" {
			envs = append(envs, "DISPLAY="+os.Getenv("DISPLAY"))
		}

		if os.Getenv("XAUTHORITY") != "" {
			envs = append(envs, "XAUTHORITY="+filepath.Join(containerHome, ".Xauthority"))
		}
	}

	return envs
}

// addHatMount mounts the hat into the user's container if they've specified one.
func (r *runner) addHatMount(mounts []mount.Mount, labels map[string]string) []mount.Mount {
	hatPath, ok := labels[hatLabel]
	if !ok {
		return mounts
	}

	return append(mounts, mount.Mount{
		Type:   "bind",
		Source: hatPath,
		Target: "~/.hat",
	})
}

const hostExtensionsDir = "~/.vscode/host-extensions"

func (r *runner) mounts(mounts []mount.Mount, image string) ([]mount.Mount, error) {
	// Mount in VS Code configs.
	mounts = append(mounts, mount.Mount{
		Type:   "bind",
		Source: vscodeConfigDir(),
		Target: "~/.config/Code",
	})
	mounts = append(mounts, mount.Mount{
		Type:   "bind",
		Source: vscodeExtensionsDir(),
		Target: hostExtensionsDir,
	})

	mounts = mountGUI(mounts)

	// 'SSH_AUTH_SOCK' is provided by a running ssh-agent. Passing in the
	// socket to the container allows for using the user's existing setup for
	// ssh authentication instead of having to create a new keys or explicity
	// pass them in.
	if sshAuthSock, exists := os.LookupEnv("SSH_AUTH_SOCK"); exists {
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

// mountGUI mounts in any x11 sockets so that they can be used
// inside the container.
func mountGUI(mounts []mount.Mount) []mount.Mount {
	if runtime.GOOS == "linux" {
		// Only mount in the x11 socket if the DISPLAY env exists.
		if os.Getenv("DISPLAY") == "" {
			return mounts
		}

		const xsock = "/tmp/.X11-unix"
		mounts = append(mounts, mount.Mount{
			Type:   "bind",
			Source: xsock,
			Target: xsock,
		})

		// We also mount the xauthority file in so any xsessions can store
		// session cookies.
		if os.Getenv("XAUTHORITY") != "" {
			mounts = append(mounts, mount.Mount{
				Type:   "bind",
				Source: os.Getenv("XAUTHORITY"),
				Target: "~/.Xauthority",
			})
		}
	}

	return mounts
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

	proot, ok := img.Config.Labels[projectRootLabel]
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
		proxyURL:        cnt.Config.Labels[proxyURLLabel],
	}, nil
}

// runOnStart runs the image's `on_start` label in the container in the project directory.
func (r *runner) runOnStart(image string) error {
	cli := dockerClient()
	defer cli.Close()

	// Get project directory.
	projectDir, err := r.projectDir(image)
	if err != nil {
		return err
	}
	projectDir = resolvePath(containerHome, projectDir)

	// Get on_start label from image.
	img, _, err := cli.ImageInspectWithRaw(context.Background(), image)
	if err != nil {
		return xerrors.Errorf("failed to inspect image: %w", err)
	}
	onStartCmd, ok := img.Config.Labels[onStartLabel]
	if !ok {
		// No on_start label, so we quit early.
		return nil
	}

	// Execute the command detached in the container.
	cmd := dockutil.DetachedExecDir(r.cntName, projectDir, "/bin/bash", "-c", onStartCmd)
	return cmd.Run()
}

func (r *runner) forkProxy() error {
	var err error
	r.proxyURL, err = forkProxy(r.cntName)
	return err
}

func forkProxy(cntName string) (proxyURL string, _ error) {
	sailProxy := exec.Command(os.Args[0], "proxy", cntName)
	stdout, err := sailProxy.StdoutPipe()
	if err != nil {
		return "", xerrors.Errorf("failed to create stdout pipe: %v", err)
	}
	defer stdout.Close()

	f, err := ioutil.TempFile("", "sailproxy_"+cntName)
	if err != nil {
		return "", xerrors.Errorf("failed to open /tmp: %w", err)
	}
	defer f.Close()

	flog.Info("writing sail proxy logs to %v", f.Name())

	sailProxy.Stderr = f

	sailProxy.SysProcAttr = &syscall.SysProcAttr{
		// See https://grokbase.com/t/gg/golang-nuts/147jmc4h0k/go-nuts-starting-detached-child-process#201407185ia7a7ldk3veno3linjktq4dve
		Setpgid: true,
	}
	err = sailProxy.Start()

	_, err = fmt.Fscan(stdout, &proxyURL)
	if err != nil {
		return "", xerrors.Errorf("proxy failed to output URL: %v", err)
	}

	_, err = url.Parse(proxyURL)
	if err != nil {
		return "", xerrors.Errorf("failed to parse proxy url from proxy: %v", err)
	}

	return proxyURL, nil
}

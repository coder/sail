package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.coder.com/sail/internal/dockutil"

	"go.coder.com/sail/internal/browserapp"
	"go.coder.com/sail/internal/codeserver"

	"github.com/docker/docker/api/types"
	"go.coder.com/flog"
	"go.coder.com/sail/internal/xexec"
	"golang.org/x/xerrors"
)

type projectStatus string

const (
	on  projectStatus = "on"
	off projectStatus = "off"
)

// project represents a sail project.
type project struct {
	conf config
	repo repo
}

func (p *project) pathName() string {
	return strings.TrimSuffix(p.repo.Path, ".git")
}

func (p *project) localDir() string {
	hostHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	path := strings.TrimSuffix(p.repo.Path, ".git")
	projectDir := filepath.Join(p.conf.ProjectRoot, path)

	projectDir = resolvePath(hostHomeDir, projectDir)
	return projectDir
}

func (p *project) dockerfilePath() string {
	return filepath.Join(p.localDir(), ".sail", "Dockerfile")
}

// clone clones a git repository to dir.
func clone(repo repo, dir string) error {
	uri := repo.CloneURI()
	cmd := xexec.Fmt("git clone %v %v", uri, dir)
	xexec.Attach(cmd)

	err := cmd.Run()
	if err != nil {
		return xerrors.Errorf("failed to clone '%s' to '%s': %w", uri, dir, err)
	}
	return nil
}

func isContainerNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "No such container")
}

func (p *project) cntExists() (bool, error) {
	cli := dockerClient()
	defer cli.Close()

	_, err := cli.ContainerInspect(context.Background(), p.cntName())
	if err != nil {
		if isContainerNotFoundError(err) {
			return false, nil
		}
		return false, xerrors.Errorf("failed to inspect %v: %w", p.cntName(), err)
	}
	return true, nil
}

func (p *project) running() (bool, error) {
	cli := dockerClient()
	defer cli.Close()

	cnt, err := cli.ContainerInspect(context.Background(), p.cntName())
	if err != nil {
		return false, xerrors.Errorf("failed to get container %v: %v", p.cntName(), err)
	}
	return cnt.State.Running, nil
}

func (p *project) requireRunning() {
	running, err := p.running()
	if err != nil {
		flog.Fatal("%v", err)
	}
	if !running {
		flog.Fatal("container %v is not running", p.cntName())
	}
}

// ensureDir ensures that a project directory exists or creates
// one if it doesn't exist.
func (p *project) ensureDir() error {
	err := os.MkdirAll(p.localDir(), 0750)
	if err != nil {
		return xerrors.Errorf("failed to make project dir %v: %w", p.localDir(), err)
	}

	// If the git directory exists, don't bother re-downloading the project.
	gitDir := filepath.Join(p.localDir(), ".git")
	_, err = os.Stat(gitDir)
	if err == nil {
		return nil
	}

	return clone(p.repo, p.localDir())
}

// buildImage finds the `.sail/Dockerfile` in the project directory
// and builds it. It sets the sail base image label on the image
// so the runner can use it when creating the container.
func (p *project) buildImage() (string, bool, error) {
	const relPath = ".sail/Dockerfile"
	path := filepath.Join(p.localDir(), relPath)

	_, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", false, xerrors.Errorf("failed to stat %v: %w", path, err)
		}

		err = p.ensureDefaultImage()
		if err != nil {
			return "", false, xerrors.Errorf("failed to pull default image %v: %w", p.conf.DefaultImage, err)
		}
		return "", false, nil
	}

	imageID := p.repo.DockerName()

	cmdStr := fmt.Sprintf("docker build --network=host -t %v -f %v %v --label %v=%v",
		imageID, path, p.localDir(), baseImageLabel, imageID,
	)
	flog.Info("running %v", cmdStr)
	cmd := xexec.Fmt(cmdStr)
	xexec.Attach(cmd)
	err = cmd.Run()
	if err != nil {
		return "", false, xerrors.Errorf("failed to build: %w", err)
	}
	return imageID, true, nil
}

func (p *project) ensureDefaultImage() error {
	flog.Info("ensuring default image %v exists", p.conf.DefaultImage)

	cmd := xexec.Fmt("docker pull %s", p.conf.DefaultImage)
	xexec.Attach(cmd)

	return cmd.Run()
}

func (p *project) cntName() string {
	return p.repo.DockerName()
}

// containerDir returns the directory of which the project is mounted within the container.
func (p *project) containerDir() (string, error) {
	client := dockerClient()
	defer client.Close()

	cnt, err := client.ContainerInspect(context.Background(), p.cntName())
	if err != nil {
		return "", err
	}
	dir, ok := cnt.Config.Labels[projectDirLabel]
	if !ok {
		return "", xerrors.Errorf("no %v label", projectDirLabel)
	}
	return dir, nil
}

func (p *project) readCodeServerLog() ([]byte, error) {
	cmd := xexec.Fmt("docker logs %v", p.cntName())

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, xerrors.Errorf("failed to cat %v: %w", containerLogPath, err)
	}

	return out, nil
}

// waitOnline waits until code-server has bound to it's port.
func (p *project) waitOnline() error {
	cli := dockerClient()
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	for ctx.Err() == nil {
		cnt, err := cli.ContainerInspect(ctx, p.cntName())
		if err != nil {
			return err
		}
		if !cnt.State.Running {
			return xerrors.Errorf("container %v not running", p.cntName())
		}

		_, err = codeserver.PID(p.cntName())
		if err == nil {
			return nil
		}

		time.Sleep(time.Millisecond * 100)
	}

	return ctx.Err()
}

func (p *project) open() error {
	cli := dockerClient()
	defer cli.Close()

	err := cli.ContainerStart(context.Background(), p.cntName(), types.ContainerStartOptions{})
	if err != nil {
		return xerrors.Errorf("failed to start container: %w", err)
	}

	port, err := codeServerPort(p.cntName())
	if err != nil {
		return err
	}

	u := "http://localhost:" + port

	flog.Info("opening %v", u)
	return browserapp.Open(u)
}

func (p *project) delete() error {
	cli := dockerClient()
	defer cli.Close()

	return dockutil.StopRemove(context.Background(), cli, p.cntName())
}

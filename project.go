package main

import (
	"context"
	"fmt"
	"github.com/skratchdot/open-golang/open"
	"go.coder.com/flog"
	"go.coder.com/narwhal/internal/xexec"
	"go.coder.com/narwhal/internal/xnet"
	"golang.org/x/xerrors"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type projectStatus string

const (
	on  projectStatus = "on"
	off projectStatus = "off"
)

// project represents a narwhal project.
type project struct {
	gf   *globalFlags
	conf config
	repo repo
}

func (p *project) name() string {
	return strings.TrimSuffix(p.repo.Path, ".git")
}

func (p *project) localDir() string {
	path := strings.TrimSuffix(p.repo.Path, ".git")
	projectDir := filepath.Join(p.conf.ProjectRoot, path)
	return cleanPath(projectDir)
}

func (p *project) dockerfilePath() string {
	return filepath.Join(p.localDir(), ".narwhal", "Dockerfile")
}

// clone clones a git repository on h.
// It returns a path to the repository.
func clone(repo repo, dir string) {
	cmd := xexec.Fmt("git clone %v %v", repo.CloneURI(), dir)
	xexec.Attach(cmd)

	err := cmd.Run()
	if err != nil {
		flog.Fatal("failed to clone project: %v", err)
	}
}

// pull pulls the latest changes for the repo.
func pull(repo repo, dir string) {
	cmd := xexec.Fmt("git pull --all")
	xexec.Attach(cmd)
	cmd.Dir = dir

	err := cmd.Run()
	if err != nil {
		flog.Fatal("failed to pull project: %v", err)
	}
}

func isContainerNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "No such container")
}

func (p *project) cntExists() bool {
	cli := dockerClient()
	_, err := cli.ContainerInspect(context.Background(), p.cntName())
	if err != nil {
		if isContainerNotFoundError(err) {
			return false
		}
		flog.Fatal("failed to inspect %v: %v", p.cntName(), err)
	}
	return true
}

func (p *project) running() bool {
	cli := dockerClient()
	cnt, err := cli.ContainerInspect(context.Background(), p.cntName())
	if err != nil {
		flog.Fatal("failed to get container %v: %v", p.cntName(), err)
	}
	return cnt.State.Running
}

func (p *project) requireRunning() {
	if !p.running() {
		flog.Fatal("container %v is not running", p.cntName())
	}
}

// ensureDir ensures that a
// project directory exists or creates one if it doesn't exist.
func (p *project) ensureDir() {
	err := os.MkdirAll(p.localDir(), 0750)
	if err != nil {
		flog.Fatal("failed to make project dir %v: %v", p.localDir(), err)
	}

	// If the git directory exists, don't bother re-downloading the project.
	gitDir := filepath.Join(p.localDir(), ".git")
	_, err = os.Stat(gitDir)
	if err == nil {
		pull(p.repo, p.localDir())
		return
	}

	clone(p.repo, p.localDir())
}

// buildImage finds the `.narwhal/Dockerfile` in the project directory
// and builds it.
func (p *project) buildImage() (string, bool, error) {
	const relPath = ".narwhal/Dockerfile"
	path := filepath.Join(p.localDir(), relPath)

	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, xerrors.Errorf("failed to stat %v: %w", path, err)
	}

	imageID := p.repo.DockerName()

	cmdStr := fmt.Sprintf("docker build -t %v -f %v %v", imageID, path, p.localDir())
	flog.Info("running %v", cmdStr)
	cmd := xexec.Fmt(cmdStr)
	xexec.Attach(cmd)
	err = cmd.Run()
	if err != nil {
		return "", false, xerrors.Errorf("failed to build: %w", err)
	}
	return imageID, true, nil
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

func (p *project) Exec(cmd string, args ...string) *exec.Cmd {
	args = append([]string{"exec", "-i", p.cntName(), cmd}, args...)
	return exec.Command("docker", args...)
}

func (p *project) ExecTTY(cmd string, args ...string) *exec.Cmd {
	args = append([]string{"exec", "-w", "/root", "-it", p.cntName(), cmd}, args...)
	return exec.Command("docker", args...)
}

func (p *project) FmtExec(cmdFmt string, args ...interface{}) *exec.Cmd {
	return p.Exec("bash", "-c", fmt.Sprintf(cmdFmt, args...))
}

func (p *project) DetachedExec(cmd string, args ...string) *exec.Cmd {
	args = append([]string{"exec", "-d", p.cntName(), cmd}, args...)
	return exec.Command("docker", args...)
}

func (p *project) ExecEnv(envs []string, cmd string, args ...string) *exec.Cmd {
	args = append([]string{"exec", "-e", strings.Join(envs, ","), "-i", p.cntName(), cmd}, args...)
	return exec.Command("docker", args...)
}

// CodeServerPort gets the port of the running code-server binary.
func (p *project) CodeServerPort() (string, error) {
	cli := dockerClient()
	defer cli.Close()

	cnt, err := cli.ContainerInspect(context.Background(), p.cntName())
	if err != nil {
		return "", err
	}

	port, ok := cnt.Config.Labels[portLabel]
	if !ok {
		return "", xerrors.Errorf("no %v label found", portLabel)
	}
	return port, nil
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

		port, ok := cnt.Config.Labels[portLabel]
		if !ok {
			return xerrors.Errorf("no %v label found", portLabel)
		}

		if !xnet.PortFree(port) {
			return nil
		}

		time.Sleep(time.Millisecond * 100)
	}

	return ctx.Err()
}

func (p *project) open() error {
	port, err := p.CodeServerPort()
	if err != nil {
		return err
	}

	u := "http://" + net.JoinHostPort("127.0.0.1", port)
	flog.Info("opening browser serving %v", u)
	return open.Run(u)
}

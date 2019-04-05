package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/skratchdot/open-golang/open"
	"go.coder.com/flog"
	"go.coder.com/narwhal/internal/xexec"
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

func (p *project) dir() string {
	path := strings.TrimSuffix(p.repo.Path, ".git")
	projectDir := filepath.Join(p.conf.ProjectRoot, path)
	return cleanPath(projectDir)
}

func (p *project) dockerfilePath() string {
	return filepath.Join(p.dir(), ".narwhal", "Dockerfile")
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

func (p *project) cntExists() bool {
	cli := dockerClient()
	_, err := cli.ContainerInspect(context.Background(), p.cntName())
	if err != nil {
		if strings.Contains(err.Error(), "No such container") {
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
	err := os.MkdirAll(p.dir(), 0750)
	if err != nil {
		flog.Fatal("failed to make project dir %v: %v", p.dir(), err)
	}

	// If the git directory exists, don't bother re-downloading the project.
	gitDir := filepath.Join(p.dir(), ".git")
	_, err = os.Stat(gitDir)
	if err == nil {
		pull(p.repo, p.dir())
		return
	}

	clone(p.repo, p.dir())
}

// buildImage finds the `.narwhal/Dockerfile` in the project directory
// and builds it.
func (p *project) buildImage() (string, bool, error) {
	const relPath = ".narwhal/Dockerfile"
	path := filepath.Join(p.dir(), relPath)

	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, xerrors.Errorf("failed to stat %v: %w", path, err)
	}

	imageID := p.repo.DockerName()

	cmdStr := fmt.Sprintf("docker build -t %v -f %v %v", imageID, path, p.dir())
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
func (p *project) containerDir() string {
	return filepath.Join(p.conf.ContainerProjectRoot, p.repo.BaseName())
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

func (p *project) CodeServerRunning() bool {
	cmd := p.FmtExec("pgrep code-server")
	return cmd.Run() == nil
}

// CodeServerPort gets the port of the running code-server binary.
func (p *project) CodeServerPort() (string, error) {
	// netstat and similar may not work correctly in the container due to procfs limitations.
	// So, we rely on this janky ps regex.
	cmd := p.FmtExec(`ps aux | grep -Po "(?<=code-server --port )([0-9]{1,5})" | head -n 1`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.New("no processes found")
	}

	return strings.TrimSpace(string(out)), nil
}

func (p *project) ReadCodeServerLog() ([]byte, error) {
	cmd := p.Exec("cat", CodeServerLogLocation)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, xerrors.Errorf("failed to cat %v: %w", CodeServerLogLocation, err)
	}

	return out, nil
}

// StartCodeServer starts code-server and binds it to the given port.
func (p *project) StartCodeServer(dir string, port string) error {
	if p.CodeServerRunning() {
		return errCodeServerRunning
	}

	cmd := p.DetachedExec(
		"bash", "-c",
		// Port must be first parameter for janky port detection to work.
		"cd "+dir+"; code-server --port "+port+" --data-dir ~/.config/Code --extensions-dir ~/.vscode/extensions --allow-http --no-auth 2>&1 > "+CodeServerLogLocation,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return xerrors.Errorf("failed to start code-server: %w\n%s", err, out)
	}

	expires := time.Now().Add(time.Second * 10)

	for time.Now().Before(expires) {
		if !p.CodeServerRunning() {
			return errCodeServerFailed
		}
		if checkPort(port) {
			return nil
		}

		time.Sleep(time.Millisecond * 100)
	}

	return errCodeServerTimedOut
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

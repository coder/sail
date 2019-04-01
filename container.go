package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"golang.org/x/xerrors"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// container describes a remote or local container.
type container struct {
	Name string
	Host Host
}

func (c container) Exec(cmd string, args ...string) *exec.Cmd {
	args = append([]string{"exec", "-i", c.Name, cmd}, args...)
	return c.Host.Command("docker", args...)
}

func (c container) FmtExec(cmdFmt string, args ...interface{}) *exec.Cmd {
	return c.Exec("bash", "-c", fmt.Sprintf(cmdFmt, args...))
}

func (c container) DetachedExec(cmd string, args ...string) *exec.Cmd {
	args = append([]string{"exec", "-d", c.Name, cmd}, args...)
	return c.Host.Command("docker", args...)
}

func (c container) ExecEnv(envs []string, cmd string, args ...string) *exec.Cmd {
	args = append([]string{"exec", "-e", strings.Join(envs, ","), "-i", c.Name, cmd}, args...)
	return c.Host.Command("docker", args...)
}

// RunContainer creates and runs a new container.
func RunContainer(log io.Writer, h Host, image string, name string) error {
	// We run sh to the container never terminated.
	cmd := h.Command("docker", "run", "--name", name, "-dt", image, "sh")
	cmd.Stdout = log
	cmd.Stderr = log
	return cmd.Run()
}

// ListContainers lists containers with the given prefix.
// Names are returned in descending order with respect to when it
// was created.
func ListContainers(h Host, prefix string) ([]string, error) {
	cmd := h.Command("docker", "ps",
		"--format", "{{ .Names }}", "--filter", "name=codercom-bigdur",
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, xerrors.Errorf("failed to get list of containers: %w", err)
	}

	var names []string
	for _, v := range bytes.Split(out, []byte("\n")) {
		names = append(names, string(v))
	}
	return names, nil
}

// DownloadCodeServer downloads code-server to /usr/bin/code-server.
func (c container) DownloadCodeServer() error {
	url, err := CodeServerDownloadURL(context.Background(), c.Host.OS())
	if err != nil {
		return xerrors.Errorf("failed to get download url: %w", err)
	}

	cmd := c.ExecEnv([]string{
		"URL=" + url,
	}, "bash", "-c", CodeServerExtractScript, url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return xerrors.Errorf("failed to run extract script: %w\n%s", err, CodeServerExtractScript)
	}

	return nil
}

// CodeServerLogLocation is the location of the code-server log.
const CodeServerLogLocation = "/tmp/code-server.log"

func (c container) CodeServerRunning() bool {
	cmd := c.FmtExec("pgrep code-server")
	return cmd.Run() == nil
}

// CheckPort returns true if the port is bound.
func (c container) CheckPort(port string) bool {
	cmd := c.FmtExec("lsof -i:%v", port)
	return cmd.Run() == nil
}

var (
	errCodeServerRunning  = errors.New("code-server is already running")
	errCodeServerTimedOut = errors.New("code-server took too long to start")
	errCodeServerFailed   = errors.New("code-server failed to start")
)

func (c container) ReadCodeServerLog() ([]byte, error) {
	cmd := c.Exec("cat", CodeServerLogLocation)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, xerrors.Errorf("failed to cat %v: %w", CodeServerLogLocation, err)
	}

	return out, nil
}

// StartCodeServer starts code-server and binds it to the given port.
func (c container) StartCodeServer(port string) error {
	if c.CodeServerRunning() {
		return errCodeServerRunning
	}

	cmd := c.DetachedExec(
		"bash", "-c",
		"code-server --port "+port+" 2>&1 > "+CodeServerLogLocation,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return xerrors.Errorf("failed to start code-server: %w\n%s", err, out)
	}

	expires := time.Now().Add(time.Second * 10)

	for time.Now().Before(expires) {
		if !c.CodeServerRunning() {
			return errCodeServerFailed
		}
		if c.CheckPort(port) {
			return nil
		}

		time.Sleep(time.Millisecond * 100)
	}

	return errCodeServerTimedOut
}

// FindAvailablePort finds an available port in the provided range.
func (c container) FindAvailablePort(start uint16, end uint16) (string, error) {

}

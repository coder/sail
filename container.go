package main

import (
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/xerrors"
	"io"
	"math/rand"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// container describes a local container.
type container struct {
	Name string
}

func (c container) Exec(cmd string, args ...string) *exec.Cmd {
	args = append([]string{"exec", "-i", c.Name, cmd}, args...)
	return exec.Command("docker", args...)
}

func (c container) ExecTTY(cmd string, args ...string) *exec.Cmd {
	args = append([]string{"exec", "-w", "/root", "-it", c.Name, cmd}, args...)
	return exec.Command("docker", args...)
}

func (c container) FmtExec(cmdFmt string, args ...interface{}) *exec.Cmd {
	return c.Exec("bash", "-c", fmt.Sprintf(cmdFmt, args...))
}

func (c container) DetachedExec(cmd string, args ...string) *exec.Cmd {
	args = append([]string{"exec", "-d", c.Name, cmd}, args...)
	return exec.Command("docker", args...)
}

func (c container) ExecEnv(envs []string, cmd string, args ...string) *exec.Cmd {
	args = append([]string{"exec", "-e", strings.Join(envs, ","), "-i", c.Name, cmd}, args...)
	return exec.Command("docker", args...)
}

const narwhalLabel = "narwhal"

type containerConfig struct {
	image    string
	name     string
	hostname string
	shares   map[string]string
}

// runContainer creates and runs a new container.
func runContainer(log io.Writer, c containerConfig) error {
	var volumeFlag strings.Builder
	for k, v := range c.shares {
		fmt.Fprintf(&volumeFlag, "-v %v:%v ", k, v)
	}

	// Use host network to remove the need for export.
	// We run sleep so the container never terminates.
	// We don't run sh because sh doesn't kill on SIGTERM.
	cmd := fmtExec("docker run %s -h %v --network=host --label=%v --name %v -dt %v sleep 900000d",
		volumeFlag.String(), c.hostname,
		narwhalLabel, c.name, c.image,
	)
	// This only outputs the container name.
	//cmd.Stdout = log
	cmd.Stderr = log
	return cmd.Run()
}

// listContainers lists containers with the given prefix.
// Names are returned in descending order with respect to when it
// was created.
func listContainers(all bool, prefix string) ([]string, error) {
	var allFlag string
	if all {
		allFlag = "-a"
	}

	cmd := fmtExec("docker ps %v --format '{{ .Names }}' --filter name=%v --filter label=%v",
		allFlag, prefix, narwhalLabel,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, xerrors.Errorf("failed to get list of containers: %w", err)
	}

	var names []string
	for _, v := range bytes.Split(bytes.TrimSpace(out), []byte("\n")) {
		v = bytes.TrimSpace(v)
		if string(v) == "" {
			continue
		}
		names = append(names, string(v))
	}
	return names, nil
}

func containerExists(name string) (bool, error) {
	out, err := fmtExec("docker inspect %v", name).CombinedOutput()
	if err != nil {
		if bytes.Contains(out, []byte("No such object")) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// CodeServerLogLocation is the location of the code-server log.
const CodeServerLogLocation = "/tmp/code-server.log"

func (c container) CodeServerRunning() bool {
	cmd := c.FmtExec("pgrep code-server")
	return cmd.Run() == nil
}

// CodeServerPort gets the port of the running code-server binary.
func (c container) CodeServerPort() (string, error) {
	// netstat and similar may not work correctly in the container due to procfs limitations.
	// So, we rely on this janky ps regex.
	cmd := c.FmtExec(`ps aux | grep -Po "(?<=code-server --port )([0-9]{1,5})" | head -n 1`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.New("no processes found")
	}

	return strings.TrimSpace(string(out)), nil
}

// checkPort returns true if the port is bound.
// We want to run this on the host and not in the container
func checkPort(port string) bool {
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return false
	}
	_ = l.Close()
	return true
}

func findAvailablePort(min, max uint16) (string, error) {
	for _, tryPort := range rand.Perm(int(max - min)) {
		tryPort += int(min)

		strport := strconv.Itoa(tryPort)
		if checkPort(strport) {
			return strport, nil
		}
	}
	return "", errors.New("no availabe ports")
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
func (c container) StartCodeServer(dir string, port string) error {
	if c.CodeServerRunning() {
		return errCodeServerRunning
	}

	cmd := c.DetachedExec(
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
		if !c.CodeServerRunning() {
			return errCodeServerFailed
		}
		if checkPort(port) {
			return nil
		}

		time.Sleep(time.Millisecond * 100)
	}

	return errCodeServerTimedOut
}

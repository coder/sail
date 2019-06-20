package dockutil

import (
	"fmt"
	"os/exec"
	"strings"
)

func Exec(cntName, cmd string, args ...string) *exec.Cmd {
	args = append([]string{"exec", "-i", cntName, cmd}, args...)
	return exec.Command("docker", args...)
}

func ExecTTY(cntName, dir, cmd string, args ...string) *exec.Cmd {
	args = append([]string{"exec", "-w", dir, "-it", cntName, cmd}, args...)
	return exec.Command("docker", args...)
}

func FmtExec(cntName, cmdFmt string, args ...interface{}) *exec.Cmd {
	return Exec(cntName, "bash", "-c", fmt.Sprintf(cmdFmt, args...))
}

func DetachedExec(cntName, cmd string, args ...string) *exec.Cmd {
	args = append([]string{"exec", "-d", cntName, cmd}, args...)
	return exec.Command("docker", args...)
}

func DetachedExecDir(cntName, dir, cmd string, args ...string) *exec.Cmd {
	args = append([]string{"exec", "-dw", dir, cntName, cmd}, args...)
	return exec.Command("docker", args...)
}

func ExecEnv(cntName string, envs []string, cmd string, args ...string) *exec.Cmd {
	args = append([]string{"exec", "-e", strings.Join(envs, ","), "-i", cntName, cmd}, args...)
	return exec.Command("docker", args...)
}

package main

import (
	"fmt"
	"go.coder.com/flog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type projectStatus string

const (
	on  projectStatus = "on"
	off projectStatus = "off"
)

// project represents a narwhal project.
type project struct {
	flg  flags
	conf config
	repo repo
}

func (p *project) dir() string {
	path := strings.TrimSuffix(p.repo.Path, ".git")
	projectDir := filepath.Join(p.conf.ProjectRoot, path)
	return projectDir
}

// clone clones a git repository on h.
// It returns a path to the repository.
func clone(repo repo, dir string) {
	cmd := fmtExec("git clone %v %v", repo.CloneURI(), dir)
	attach(cmd)

	err := cmd.Run()
	if err != nil {
		flog.Fatal("failed to clone project: %v", err)
	}
}

// pull pulls the latest changes for the repo.
func pull(repo repo, dir string) {
	cmd := fmtExec("git pull --all")
	attach(cmd)
	cmd.Dir = dir

	err := cmd.Run()
	if err != nil {
		flog.Fatal("failed to pull project: %v", err)
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
	fmt.Printf("stat %v %v", gitDir, err)

	clone(p.repo, p.dir())
}

// ensureImage finds the `.narwhal/Dockerfile` in the project directory
// and builds it.
func (p *project) ensureImage() (string, bool) {
	const relPath = ".narwhal/Dockerfile"
	path := filepath.Join(p.dir(), relPath)

	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false
		}
		flog.Fatal("failed to stat %v: %v", path, err)
	}

	imageID := p.repo.DockerName()

	cmdStr := fmt.Sprintf("docker build -t %v -f %v %v", imageID, path, p.dir())
	flog.Info("running %v", cmdStr)
	cmd := fmtExec(cmdStr)
	attach(cmd)
	err = cmd.Run()
	if err != nil {
		flog.Fatal("failed to docker build: %v", err)
	}
	return imageID, true
}

// ensureDockerDaemon verifies that Docker is running.
func (flg *flags) ensureDockerDaemon() {
	out, err := exec.Command("docker", "info").CombinedOutput()
	if err != nil {
		flog.Fatal("failed to run `docker info`: %v\n%s", err, out)
	}
	flg.debug("verified Docker is running")
}


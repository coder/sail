package environment

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"golang.org/x/xerrors"
)

type Environment struct {
	repo *Repo
	name string
	cnt  types.ContainerJSON
}

var ErrMissingContainer = xerrors.Errorf("missing container")

// FindEnvironment tries to find a container for the given repo, or
// ErrMissingContainer if not found.
func FindEnvironment(ctx context.Context, r *Repo) (*Environment, error) {
	cli := dockerClient()
	defer cli.Close()

	name := r.DockerName()

	cnt, err := cli.ContainerInspect(ctx, name)
	if isContainerNotFoundError(err) {
		return nil, ErrMissingContainer
	}
	if err != nil {
		return nil, xerrors.Errorf("failed to inspect container: %w", err)
	}

	env := &Environment{
		repo: r,
		name: name,
		cnt:  cnt,
	}

	// Start it up.
	err = Start(ctx, env)
	if err != nil {
		return nil, err
	}

	// // TODO: If container needed to be started, we should give code-server
	// // some time to start up.
	// port, err := env.processPort(ctx, "code-server")
	// if err != nil {
	// 	return nil, err
	// }
	// env.port = port

	return env, nil
}

func Start(ctx context.Context, env *Environment) error {
	cli := dockerClient()
	defer cli.Close()

	err := cli.ContainerStart(ctx, env.name, types.ContainerStartOptions{})
	if err != nil {
		return xerrors.Errorf("failed to start container: %w", err)
	}

	return nil
}

func Stop(ctx context.Context, env *Environment) error {
	cli := dockerClient()
	defer cli.Close()

	err := cli.ContainerStop(ctx, env.name, nil)
	if err != nil {
		return xerrors.Errorf("failed to stop container: %w", err)
	}

	return nil
}

func Remove(ctx context.Context, env *Environment) error {
	cli := dockerClient()
	defer cli.Close()

	err := cli.ContainerRemove(ctx, env.name, types.ContainerRemoveOptions{})
	if err != nil {
		return xerrors.Errorf("failed to remove container: %w", err)
	}

	return nil
}

func Purge(ctx context.Context, env *Environment) error {
	err := Stop(ctx, env)
	if err != nil {
		return err
	}
	err = Remove(ctx, env)
	if err != nil {
		return err
	}

	cli := dockerClient()
	defer cli.Close()

	err = cli.VolumeRemove(ctx, formatRepoVolumeName(env.repo), true)
	if err != nil {
		return xerrors.Errorf("failed to remove volume: %w", err)
	}

	return nil
}

// clone clones the repo into the project directory.
func (e *Environment) clone(ctx context.Context, dir string) error {
	out, err := e.exec(ctx, "sudo", "chown", "-R", "user", dir).CombinedOutput()
	if err != nil {
		return xerrors.Errorf("failed to chown: %s, %w", out, err)
	}

	cloneStr := fmt.Sprintf("cd %s; git clone %s .", dir, e.repo.URL.String())
	out, err = e.exec(ctx, "bash", []string{"-c", cloneStr}...).CombinedOutput()
	if err != nil {
		return xerrors.Errorf("failed to clone: %s, %w", out, err)
	}

	return nil
}

func (e *Environment) exec(ctx context.Context, cmd string, args ...string) *exec.Cmd {
	// TODO: Use docker api.
	args = append([]string{"exec", "-i", e.name, cmd}, args...)
	return exec.CommandContext(ctx, "docker", args...)
}

var errNoSuchFile = xerrors.Errorf("no such file")

// readPath reads a path inside the environment. The returned reader is suitable
// for use with a tar reader.
//
// The root of the tar archive will be '.'
// E.g. if path is '/tmp/somedir', a file exists at '/tmp/somedir/file', the tar
// header name will be 'file'.
func (e *Environment) readPath(ctx context.Context, path string) (io.Reader, error) {
	cli := dockerClient()
	defer cli.Close()

	rdr, _, err := cli.CopyFromContainer(ctx, e.name, path)
	if isPathNotFound(err) {
		return nil, errNoSuchFile
	}
	if err != nil {
		return nil, xerrors.Errorf("failed to get reader for path '%s': %w", path, err)
	}
	defer rdr.Close()

	var (
		buf bytes.Buffer

		base = filepath.Base(path)

		tr = tar.NewReader(rdr)
		tw = tar.NewWriter(&buf)
	)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, xerrors.Errorf("failed to read from tar reader: %w", err)
		}

		hdr.Name = strings.TrimLeft(hdr.Name, base+"/")
		err = tw.WriteHeader(hdr)
		if err != nil {
			return nil, xerrors.Errorf("failed to write header: %w", err)
		}

		_, err = io.Copy(tw, tr)
		if err != nil {
			return nil, xerrors.Errorf("failed to copy: %w", err)
		}
	}
	err = tw.Close()
	if err != nil {
		return nil, xerrors.Errorf("failed to close tar writer: %w", err)
	}

	return &buf, nil
}

func isPathNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "No such container:path")
}

func isContainerNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "No such container")
}

package environment

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/mount"
	"go.coder.com/flog"
	"go.coder.com/sail/internal/randstr"
	"go.coder.com/sail/internal/sshforward"
	"golang.org/x/xerrors"
)

// Bootstrap will attempt to create an environment using the docker file
// located at '.sail/Dockerfile' within the repo.
//
// An initial default environment and volume will be created, then the repo
// will be cloned into the volume. If the sail docker file does exists, an
// image will be created and the environment will be rebuilt using that new
// image.
//
// The default environment will be returned if no sail docker file exists.
//
// TODO: Come up with a cleaner solution to needing to pass in the remote host.
// TODO: Leaks ssh forward commands. No cleanup is being done.
func Bootstrap(ctx context.Context, cfg *BuildConfig, repo *Repo, remote string) (*Environment, error) {
	// TODO: Should this always try to create?
	lv, err := ensureVolumeForRepo(ctx, repo)
	if err != nil {
		return nil, err
	}

	cfg.Image = defaultRepoImage(repo, cfg.Image)

	projectPath := defaultDirForRepo(repo)
	projectMount := mount.Mount{
		Type:   mount.TypeVolume,
		Source: lv.vol.Name,
		Target: projectPath,
	}
	cfg.Mounts = append(cfg.Mounts, projectMount)

	// Forward ssh auth sock to container if available. Allows for doing git
	// operations over ssh inside the container.
	localSock, authSockExists := os.LookupEnv("SSH_AUTH_SOCK")
	// Default to the container socket being the same as the local
	// socket. If the container is remote, the socket will need to be
	// unique and forwarded over ssh.
	forwardedSock := localSock
	if authSockExists {
		if remote != "" {
			forwardedSock = fmt.Sprintf("/tmp/sail-agent-%s.sock", randstr.Make(5))
			flog.Info("forwarding ssh auth agent through %s", forwardedSock)
			agentForwarder := sshforward.NewRemoteSocketForwarder(forwardedSock, localSock, remote)
			err := agentForwarder.Forward()
			if err != nil {
				flog.Fatal("failed to forward ssh auth sock: %v", err)
			}
			defer agentForwarder.Close()
		}

		authEnv := fmt.Sprintf("SSH_AUTH_SOCK=%s", forwardedSock)
		authMount := mount.Mount{
			Type:   mount.TypeBind,
			Source: forwardedSock,
			Target: forwardedSock,
		}

		cfg.Envs = append(cfg.Envs, authEnv)
		cfg.Mounts = append(cfg.Mounts, authMount)
	}

	env, err := Build(ctx, cfg)
	if err != nil {
		return nil, err
	}

	if authSockExists {
		// TODO: Don't do 0777. Right now it's set to 0777 to allow for
		// subsequent containers to use the same socket without having to chown
		// it every time.
		out, err := env.Exec(ctx, "sudo", "chmod", "-R", "0777", forwardedSock).CombinedOutput()
		if err != nil {
			return nil, xerrors.Errorf("failed to chmod auth sock: %s: %w", out, err)
		}
	}

	err = cloneInto(ctx, env, repo, projectPath)
	if err != nil {
		return nil, err
	}

	sailPath := filepath.Join(projectPath, ".sail")
	_, err = env.readPath(ctx, sailPath)
	if xerrors.Is(err, errNoSuchFile) {
		flog.Info("no '.sail/Dockerfile' for repo")
		return env, nil
	}
	if err != nil {
		return nil, err
	}

	prov := &EnvPathProvider{
		Env:  env,
		Path: sailPath,
	}
	env, err = Modify(ctx, prov, env)
	if err != nil {
		return nil, xerrors.Errorf("failed to modify: %w", err)
	}

	return env, nil
}

// ensureVolumeForRepo ensures that there's a volume dedicated to storing the
// repo.
func ensureVolumeForRepo(ctx context.Context, r *Repo) (*localVolume, error) {
	name := r.DockerName()
	lv, err := findLocalVolume(ctx, name)
	if xerrors.Is(err, errMissingVolume) {
		lv, err = createLocalVolume(ctx, name)
	}
	if err != nil {
		return nil, err
	}

	return lv, nil
}

// cloneInto clones the repo into the environment.
//
// The enviroment should have its auth set up in such a way that would allow the
// user to clone a private repo.
func cloneInto(ctx context.Context, env *Environment, repo *Repo, path string) error {
	out, err := env.Exec(ctx, "sudo", "mkdir", "-p", path).CombinedOutput()
	if err != nil {
		return xerrors.Errorf("failed to create dir: %s: %w", out, err)
	}

	out, err = env.Exec(ctx, "sudo", "chown", "-R", "user", path).CombinedOutput()
	if err != nil {
		return xerrors.Errorf("failed to chown: %s: %w", out, err)
	}

	uri := repo.CloneURI()
	flog.Info("cloning from %s", uri)
	cloneStr := fmt.Sprintf("cd %s; GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=no' git clone %s .", path, uri)
	// TODO: Ensure this works with ssh passphrases.
	cmd := env.Exec(ctx, "bash", []string{"-c", cloneStr}...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return xerrors.Errorf("failed to clone: %w", err)
	}

	return nil
}

var errMissingDockerfile = xerrors.Errorf("missing dockerfile")

// buildImage builds an image from the given directory. The dir should contain a
// valid dockerfile at its root.
func buildImage(ctx context.Context, buildContext io.Reader, name string) error {
	cli := dockerClient()
	defer cli.Close()

	flog.Info("building image: %s", name)

	opts := types.ImageBuildOptions{
		Tags: []string{name},
	}
	resp, err := cli.ImageBuild(ctx, buildContext, opts)
	if err != nil {
		return xerrors.Errorf("failed to build image: %w", err)
	}
	defer resp.Body.Close()

	// Mostly for debugging.
	io.Copy(ioutil.Discard, resp.Body)

	return nil
}

// defaultRepoImage returns a base image suitable for development with the
// repo's language. If the repo language isn't able to be determined, this
// returns the provided default image.
func defaultRepoImage(repo *Repo, def string) string {
	lang := (*repo).Language()

	fmtImage := func(s string) string {
		return fmt.Sprintf("codercom/ubuntu-dev-%s:latest", s)
	}

	switch strings.ToLower(lang) {
	case "go":
		return fmtImage("go")
	case "javascript", "typescript":
		return fmtImage("node12")
	case "python":
		return fmtImage("python3.7")
	case "c", "c++":
		return fmtImage("gcc8")
	case "java":
		return fmtImage("openjdk12")
	case "ruby":
		return fmtImage("ruby2.6")
	default:
		return def
	}
}

func defaultDirForRepo(r *Repo) string {
	return filepath.Join("/home/user/Projects", r.BaseName())
}

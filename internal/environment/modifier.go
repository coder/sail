package environment

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"go.coder.com/sail/internal/randstr"
	"golang.org/x/xerrors"
)

// BuildContextProvider is able to provide a build context for creating an
// image.
type BuildContextProvider interface {
	BuildContext(ctx context.Context) (io.Reader, error)
}

// Modify applies a modification to an environment by taking a build context and
// applying it the environment's existing image. The new environment will
// inherit env vars and mounts.
//
// The provided environment will be removed, allowing for the new environment to
// use the same name.
//
// The build context's "FROM" will be discarded an replaced with the reference
// to the environment's current image.
func Modify(ctx context.Context, prov BuildContextProvider, env *Environment) (*Environment, error) {
	rdr, err := prov.BuildContext(ctx)
	if err != nil {
		return nil, err
	}

	rdr, err = replaceFrom(ctx, rdr, env.cnt.Image)
	if err != nil {
		return nil, err
	}

	imgName := env.name + "_" + randstr.MakeCharset(randstr.Lower, 5)
	err = buildImage(ctx, rdr, imgName)
	if err != nil {
		return nil, err
	}

	err = Stop(ctx, env)
	if err != nil {
		return nil, err
	}
	err = Remove(ctx, env)
	if err != nil {
		return nil, err
	}

	cfg := &BuildConfig{
		Name:   env.name,
		Image:  imgName,
		Envs:   env.cnt.Config.Env,
		Mounts: env.cnt.HostConfig.Mounts,
	}

	env, err = Build(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return env, nil
}

// LocalProvider provides a local build context. This may be a path pointing
// directly to a dockerfile, or its containing directory.
type LocalProvider string

var _ BuildContextProvider = new(LocalProvider)

func (p LocalProvider) BuildContext(ctx context.Context) (io.Reader, error) {
	dir := string(p)
	_, err := ioutil.ReadDir(dir)
	if err != nil && strings.Contains(err.Error(), "not a directory") {
		dir = filepath.Dir(string(p))
		_, err = ioutil.ReadDir(dir)
	}
	if err != nil {
		return nil, xerrors.Errorf("failed to read dir '%s': %w", string(p), err)
	}

	var (
		buf bytes.Buffer
		tw  = tar.NewWriter(&buf)
	)
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Should all be relative to the root of the dir.
		link := strings.TrimPrefix(path, filepath.Clean(dir)+"/")
		if link == "" {
			// Skip root.
			return nil
		}

		hdr, err := tar.FileInfoHeader(info, link)
		if err != nil {
			return xerrors.Errorf("failed to get header from file info: %w", err)
		}
		// By default, FileInfoHeader only puts the basename on the info. We
		// want the complete relative path (relative to the root of tar).
		hdr.Name = link

		if info.IsDir() {
			err = tw.WriteHeader(hdr)
			if err != nil {
				return xerrors.Errorf("failed to write header for dir: %w", err)
			}
			return nil
		}

		bs, err := ioutil.ReadFile(path)
		if err != nil {
			return xerrors.Errorf("failed to read file '%s': %w", info.Name(), err)
		}

		hdr.Size = int64(len(bs))
		err = tw.WriteHeader(hdr)
		if err != nil {
			return xerrors.Errorf("failed to write header '%s': %w", info.Name, err)
		}

		_, err = tw.Write(bs)
		if err != nil {
			return xerrors.Errorf("failed to write to tar writer: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	err = tw.Close()
	if err != nil {
		return nil, xerrors.Errorf("failed to close tar writer: %w", err)
	}

	return &buf, nil
}

// ExternalGitProvider is able to provide a build context from a git repo.
//
// The provided environment will be used to clone the repo. When ssh agent
// forwarding works, this will allow for private repos to be used.
//
// TODO: Maybe git clone outside of the the environment, then use the
// LocalProvider to provide the dockerfile.
type ExternalGitProvider struct {
	WorkingEnv *Environment
	URI        string
}

var _ BuildContextProvider = new(ExternalGitProvider)

func (p *ExternalGitProvider) BuildContext(ctx context.Context) (io.Reader, error) {
	workingDir := "/tmp/hat-working-" + randstr.Make(5)

	cloneStr := fmt.Sprintf("mkdir -p %s; cd %s; git clone %s .", workingDir, workingDir, p.URI)
	out, err := p.WorkingEnv.exec(ctx, "bash", []string{"-c", cloneStr}...).CombinedOutput()
	if err != nil {
		return nil, xerrors.Errorf("failed to clone: %s, %w", out, err)
	}

	rdr, err := p.WorkingEnv.readPath(ctx, workingDir)
	if err != nil {
		return nil, err
	}

	return rdr, nil
}

// ClonedRepoProvider provides a build context from the given path inside the
// environment.
type EnvPathProvider struct {
	Env  *Environment
	Path string
}

var _ BuildContextProvider = new(EnvPathProvider)

func (p *EnvPathProvider) BuildContext(ctx context.Context) (io.Reader, error) {
	return p.Env.readPath(ctx, p.Path)
}

type RawDockerfileProvider []byte

func (p RawDockerfileProvider) BuildContext(_ context.Context) (io.Reader, error) {
	var (
		buf bytes.Buffer
		tw  = tar.NewWriter(&buf)
	)

	err := tw.WriteHeader(&tar.Header{
		Name: "Dockerfile",
		Size: int64(len([]byte(p))),
	})
	if err != nil {
		return nil, xerrors.Errorf("failed to write header: %w", err)
	}
	_, err = tw.Write([]byte(p))
	if err != nil {
		return nil, xerrors.Errorf("failed to write dockerfile: %w", err)
	}
	err = tw.Close()
	if err != nil {
		return nil, xerrors.Errorf("failed to close tar writer: %w", err)
	}

	return &buf, err
}

func replaceFrom(ctx context.Context, rdr io.Reader, base string) (io.Reader, error) {
	var (
		buf bytes.Buffer
		tw  = tar.NewWriter(&buf)
		tr  = tar.NewReader(rdr)
	)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, xerrors.Errorf("failed to read from tar reader: %w", err)
		}

		bs, err := ioutil.ReadAll(tr)
		if err != nil {
			return nil, xerrors.Errorf("failed to read next file from tar reader: %w", err)
		}

		if strings.Contains(hdr.Name, "Dockerfile") {
			fiBuf := bytes.NewBuffer(make([]byte, 0, len(bs)))
			sc := bufio.NewScanner(bytes.NewReader(bs))

			for sc.Scan() {
				byt := sc.Bytes()
				if bytes.HasPrefix(byt, []byte("FROM")) {
					byt = []byte("FROM " + base)
				}
				fiBuf.Write(byt)
				fiBuf.WriteByte('\n')
			}

			bs = bytes.TrimSpace(fiBuf.Bytes())
			hdr.Size = int64(len(bs))
		}

		err = tw.WriteHeader(hdr)
		if err != nil {
			return nil, xerrors.Errorf("failed to write header: %w", err)
		}
		_, err = tw.Write(bs)
		if err != nil {
			return nil, xerrors.Errorf("failed to write file bytes: %w", err)
		}
	}

	return &buf, nil
}

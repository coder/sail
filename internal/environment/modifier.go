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

	"github.com/codercom/crand"
	"github.com/docker/docker/api/types"
	"go.coder.com/flog"
	"golang.org/x/xerrors"
)

// Modifier allows for modifying environments.
type Modifier interface {
	Modify(ctx context.Context, env *Environment) (*Environment, error)
}

// BuildContextProvider is able to provide a build context for creating an
// image.
type BuildContextProvider interface {
	BuildContext(ctx context.Context) (io.Reader, error)
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
		// want the complete relative path.
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

// GitRepoProvider is able to provide a build context from a git repo.
//
// The provided environment will be used to clone the repo. When ssh agent
// forwarding works, this will allow for private repos to be used.
type GitRepoProvider struct {
	WorkingEnv *Environment
	URI        string
}

var _ BuildContextProvider = new(GitRepoProvider)

func (p *GitRepoProvider) BuildContext(ctx context.Context) (io.Reader, error) {
	workingDir := "/tmp/hat-working-" + crand.MustString(5)

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

type modifier struct {
	provider BuildContextProvider
}

var _ Modifier = new(modifier)

func NewModifier(p BuildContextProvider) Modifier {
	return &modifier{
		provider: p,
	}
}

func (m *modifier) Modify(ctx context.Context, env *Environment) (*Environment, error) {
	rdr, err := m.provider.BuildContext(ctx)
	if err != nil {
		return nil, err
	}

	rdr, err = replaceFrom(ctx, rdr, env.cnt.Image)
	if err != nil {
		return nil, err
	}

	imgName := env.name + "_" + crand.MustStringCharset(crand.Lower, 5)
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

	b := &Builder{
		image: imgName,
		repo:  env.repo,
		// skipClone: true,
	}

	env, err = b.Build(ctx)
	if err != nil {
		return nil, err
	}

	return env, nil
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
	io.Copy(os.Stdout, resp.Body)

	return nil
}

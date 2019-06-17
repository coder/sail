package main

import (
	"context"
	"math/rand"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.coder.com/sail/internal/dockutil"
	"go.coder.com/sail/internal/xnet"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type params struct {
	rb     *rollback
	proj   *project
	bldr   *hatBuilder
	runner *runner
	port   string
}

// run sets up our params test suite. It ensures that the project image
// and hat image are applied if one is specified, and the container is
// started. Once all of the setup is done, then any of the provided test
// functions can be run.
func run(t *testing.T, name, repo, hatPath string, fns ...func(t *testing.T, p *params)) {
	t.Run(name, func(t *testing.T) {
		t.Parallel()

		p := &params{
			rb: newRollback(),
		}

		defer p.rb.run()

		conf := mustReadConfig(filepath.Join(metaRoot(), ".sail.toml"))

		repo, err := parseRepo("ssh", "github.com", "", repo)
		require.NoError(t, err)

		p.proj = &project{
			conf: conf,
			repo: repo,
		}

		// Ensure our project repo is cloned to the local machine.
		err = p.proj.ensureDir()
		require.NoError(t, err)
		p.rb.add(func() {
			// TODO: Do we want to remove this? I accidentally deleted
			// my own sail path that I was developing in...
			// err := os.RemoveAll(p.proj.localDir())
			// require.NoError(t, err)
		})

		// Use the project's custom sail image if one is built.
		baseImage, isCustom, err := p.proj.buildImage()
		require.NoError(t, err)
		if !isCustom {
			baseImage = p.proj.conf.DefaultImage
		} else {
			p.rb.add(func() {
				requireImageRemove(t, baseImage)
			})
		}

		image := baseImage

		// Create the hat builder and apply the hat if one
		// is specified.
		p.bldr = &hatBuilder{
			hatPath:   hatPath,
			baseImage: baseImage,
		}

		if hatPath != "" {
			image, err = p.bldr.applyHat()
			require.NoError(t, err)
			p.rb.add(func() {
				requireImageRemove(t, image)
			})
		}

		// Construct our container runner and run
		// the container.
		p.port, err = xnet.FindAvailablePort()
		require.NoError(t, err)

		p.runner = &runner{
			projectName:     p.proj.repo.BaseName(),
			projectLocalDir: p.proj.localDir(),
			cntName:         p.proj.cntName(),
			hostname:        p.proj.repo.BaseName(),
			port:            p.port,
		}

		err = p.runner.runContainer(image)
		require.NoError(t, err)
		p.rb.add(func() {
			requireContainerRemove(t, p.proj.cntName())
		})

		// Iterate through all the provided testing functions.
		for _, fn := range fns {
			fn(t, p)
		}
	})
}

func requireProjectsNotRunning(t *testing.T, projects ...string) {
	runningProjects, err := listProjects()
	require.NoError(t, err)

	for _, proj := range projects {
		for _, runningProj := range runningProjects {
			require.NotEqual(t,
				proj, runningProj.name,
				"Unable to run tests, %s currently running and needed for tests", proj,
			)
		}
	}
}

func requireGetImageLabels(t *testing.T, image string) map[string]string {
	return requireImageInspect(t, image).ContainerConfig.Labels
}

func requireImageInspect(t *testing.T, image string) types.ImageInspect {
	cli := dockerClient()
	defer cli.Close()

	insp, _, err := cli.ImageInspectWithRaw(context.Background(), image)
	require.NoError(t, err)

	return insp
}

func requireGetContainerLabels(t *testing.T, cntName string) map[string]string {
	return requireContainerInspect(t, cntName).Config.Labels
}

func requireContainerInspect(t *testing.T, cntName string) types.ContainerJSON {
	cli := dockerClient()
	defer cli.Close()

	insp, err := cli.ContainerInspect(context.Background(), cntName)
	require.NoError(t, err)

	return insp
}

func assertLabel(t *testing.T, labels map[string]string, key, val string) {
	assert.Contains(t, labels, key)
	assert.Equal(t, val, labels[key])
}

func requireImageRemove(t *testing.T, image string) {
	cli := dockerClient()
	defer cli.Close()

	_, err := cli.ImageRemove(
		context.Background(),
		image,
		types.ImageRemoveOptions{
			Force:         false,
			PruneChildren: true,
		},
	)
	require.NoError(t, err)
}

func requireContainerRemove(t *testing.T, cntName string) {
	cli := dockerClient()
	defer cli.Close()

	err := dockutil.StopRemove(context.Background(), cli, cntName)
	require.NoError(t, err)
}

func requireUbuntuDevImage(t *testing.T) {
	require.NoError(t, ensureImage("codercom/ubuntu-dev"))
}

type rollback struct {
	fns []func()
}

func newRollback() *rollback {
	return &rollback{
		fns: make([]func(), 0),
	}
}

func (r *rollback) run() {
	for i := len(r.fns) - 1; i >= 0; i-- {
		r.fns[i]()
	}
}

func (r *rollback) add(fn func()) {
	r.fns = append(r.fns, fn)
}

package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.coder.com/sail/internal/dockutil"
)

func Test_runner(t *testing.T) {
	// Ensure that the testing environment won't conflict with any running sail projects.
	requireProjectsNotRunning(t, "cdr/nbin", "cdr/flog", "cdr/bigdur", "cdr/sshcode", "cdr/cli")
	requireUbuntuDevImage(t)

	// labelChecker asserts that all of the correct labels
	// are present on the image and container.
	labelChecker := func(t *testing.T, p *params) {
		t.Run("Labels", func(t *testing.T) {
			insp := requireContainerInspect(t, p.proj.cntName())

			imgLabels := requireGetImageLabels(t, insp.Image)

			assertLabel(t, imgLabels, baseImageLabel, p.bldr.baseImage)
			if p.bldr.hatPath != "" {
				assertLabel(t, imgLabels, hatLabel, p.bldr.hatPath)
			}

			labels := insp.Config.Labels
			require.NotNil(t, labels)

			assertLabel(t, labels, baseImageLabel, p.bldr.baseImage)
			if p.bldr.hatPath != "" {
				assertLabel(t, labels, hatLabel, p.bldr.hatPath)
			}
			assertLabel(t, labels, projectLocalDirLabel, p.proj.localDir())

			cntDir, err := p.proj.containerDir()
			require.NoError(t, err)
			assertLabel(t, labels, projectDirLabel, cntDir)
			assertLabel(t, labels, projectNameLabel, p.proj.repo.BaseName())
		})
	}

	// codeServerStarts ensures that the code server process
	// starts up inside the container.
	codeServerStarts := func(t *testing.T, p *params) {
		t.Run("CodeServerStarts", func(t *testing.T) {
			err := p.proj.waitOnline()
			require.NoError(t, err)
		})
	}

	// loadFromContainer ensures that our state is properly stored
	// on the container and can rebuild our in memory structures
	// correctly.
	loadFromContainer := func(t *testing.T, p *params) {
		t.Run("FromContainer", func(t *testing.T) {
			bldr, err := hatBuilderFromContainer(p.proj.cntName())
			require.NoError(t, err)

			assert.Equal(t, p.bldr.hatPath, bldr.hatPath)
			assert.Equal(t, p.bldr.baseImage, bldr.baseImage)

			runner, err := runnerFromContainer(p.proj.cntName())
			require.NoError(t, err)

			assert.Equal(t, p.runner.cntName, runner.cntName)
			assert.Equal(t, p.runner.hostname, runner.hostname)
			assert.Equal(t, p.runner.port, runner.port)
			assert.Equal(t, p.runner.projectLocalDir, runner.projectLocalDir)
			assert.Equal(t, p.runner.projectName, runner.projectName)
			assert.Equal(t, p.runner.testCmd, runner.testCmd)
		})
	}

	// containsFile ensures that a container contains a file.
	// This is used for testing the on_start label.
	containsFile := func(name, path string) func(*testing.T, *params) {
		return func(t *testing.T, p *params) {
			t.Run(name, func(t *testing.T) {
				cntDir, err := p.proj.containerDir()
				require.NoError(t, err)
				cntDir = resolvePath(containerHome, cntDir)

				// Run the file existence check using /bin/sh.
				cmdStr := fmt.Sprintf(`[ -f "%s" ]`, path)
				err = dockutil.ExecDir(p.proj.cntName(), cntDir, "/bin/sh", "-c", cmdStr).Run()
				require.NoError(t, err)
			})
		}
	}

	run(t, "BaseImageNoHat", "https://github.com/cdr/nbin", "",
		labelChecker,
		codeServerStarts,
		loadFromContainer,
	)

	run(t, "BaseImageHat", "https://github.com/cdr/flog", "./hat-examples/fish",
		labelChecker,
		codeServerStarts,
		loadFromContainer,
	)

	run(t, "ProjImageNoHat", "https://github.com/cdr/bigdur", "",
		labelChecker,
		codeServerStarts,
		loadFromContainer,
	)

	run(t, "ProjImageHat", "https://github.com/cdr/sshcode", "./hat-examples/net",
		labelChecker,
		codeServerStarts,
		loadFromContainer,
	)

	run(t, "ProjImageOnStartHat", "https://github.com/cdr/cli", "./hat-examples/on_start",
		labelChecker,
		codeServerStarts,
		loadFromContainer,

		// ./hat-examples/on_start should create `did_on_start` in the project directory.
		containsFile("ContainsOnStartFile", "did_on_start"),
	)
}

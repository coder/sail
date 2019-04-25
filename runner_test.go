package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_runner(t *testing.T) {
	requireNoRunningSailContainers(t)

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
			assert.Equal(t, p.runner.hostUser+":user", runner.hostUser)
			assert.Equal(t, p.runner.hostname, runner.hostname)
			assert.Equal(t, p.runner.port, runner.port)
			assert.Equal(t, p.runner.projectLocalDir, runner.projectLocalDir)
			assert.Equal(t, p.runner.projectName, runner.projectName)
			assert.Equal(t, p.runner.testCmd, runner.testCmd)
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

	run(t, "BaseImageNoHat", "codercom/retry", "",
		labelChecker,
		loadFromContainer,
		codeServerStarts,
	)

	run(t, "BaseImageHat", "codercom/docs", "./hat-examples/fish",
		labelChecker,
		loadFromContainer,
		codeServerStarts,
	)

	run(t, "ProjImageNoHat", "codercom/bigdur", "",
		labelChecker,
		loadFromContainer,
		codeServerStarts,
	)

	run(t, "ProjImageHat", "codercom/extip", "./hat-examples/net",
		labelChecker,
		loadFromContainer,
		codeServerStarts,
	)
}

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_project(t *testing.T) {
	t.Parallel()

	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	conf := mustReadConfig(filepath.Join(metaRoot(), ".sail.toml"))

	var tests = []struct {
		schema          string
		name            string
		repo            string
		expCntName      string
		expEnsureDirErr bool
		expCustomBldImg bool
	}{
		{
			"ssh",
			"OK",
			"codercom/bigdur",
			"codercom_bigdur",
			false,
			true,
		},
		{
			"ssh",
			"RepoNotExist",
			"codercom/do-not-exist",
			"codercom_do-not-exist",
			true,
			false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			rb := newRollback()
			defer rb.run()

			repo, err := parseRepo(test.schema, "github.com", "", test.repo)
			require.NoError(t, err)

			p := &project{
				conf: conf,
				repo: repo,
			}

			name := p.cntName()
			require.Equal(t, test.expCntName, name)

			expLocalDir := filepath.Join(homeDir, "Projects", test.repo)

			require.Equal(t, expLocalDir, p.localDir())
			require.Equal(t,
				filepath.Join(expLocalDir, ".sail", "Dockerfile"),
				p.dockerfilePath(),
			)

			t.Run("EnsureDir", func(t *testing.T) {
				err = p.ensureDir()
				if test.expEnsureDirErr {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)
				rb.add(func() {
					err := os.RemoveAll(p.localDir())
					require.NoError(t, err)
				})
			})

			t.Run("BuildImage", func(t *testing.T) {
				image, isCustom, err := p.buildImage()
				require.NoError(t, err)
				if !test.expCustomBldImg {
					assert.False(t, isCustom)
					assert.Empty(t, image)
					return
				}

				assert.True(t, isCustom)

				labels := requireGetImageLabels(t, image)
				assertLabel(t, labels, baseImageLabel, p.repo.DockerName())

				rb.add(func() {
					requireImageRemove(t, p.repo.DockerName())
				})
			})
		})
	}
}

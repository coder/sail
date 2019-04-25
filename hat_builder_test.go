package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_hatBuilder(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		name      string
		baseImage string
		hatPath   string
		expectErr bool
	}{
		{
			"BaseImageNotExist",
			"codercom/do-not-exist:my-tag",
			"./hat-examples/fish",
			true,
		},
		{
			"HatNotExist",
			"codercom/ubuntu-dev",
			"./hat-examples/no-hat",
			true,
		},
		{
			"GithubHatNotExist",
			"codercom/ubuntu-dev",
			"github:codercom/no-hat",
			true,
		},
		{
			"OK",
			"codercom/ubuntu-dev",
			"./hat-examples/fish",
			false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			bldr := &hatBuilder{
				baseImage: test.baseImage,
				hatPath:   test.hatPath,
			}

			image, err := bldr.applyHat()
			if test.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			t.Run("ImageLabels", func(t *testing.T) {
				labels := requireGetImageLabels(t, image)

				assertLabel(t, labels, hatLabel, test.hatPath)
				assertLabel(t, labels, baseImageLabel, test.baseImage)
			})

			t.Run("RemoveImage", func(t *testing.T) {
				requireImageRemove(t, image)
			})
		})
	}
}

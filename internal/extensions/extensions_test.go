package extensions

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/stretchr/testify/require"
)

func TestParseExtensionList(t *testing.T) {
	l, err := ParseExtensionList(os.Getenv("HOME") + "/.vscode-oss/extensions")
	assert.NoError(t, err)

	fmt.Println(l)

	d, _ := DockerfileSetExtensions(nil, l)
	fmt.Println(string(d))
}

var testhat = `FROM codercom/ubuntu-dev

# DO NOT EDIT. EXTENSIONS MANAGED BY SAIL.
RUN code-server --install-extension AlanWalk.markdown-toc && \
	code-server --install-extension akamud.vscode-theme-onedark && \
	code-server --install-extension anseki.vscode-color && \
	code-server --install-extension azemoh.one-monokai && \
	code-server --install-extension bierner.markdown-preview-github-styles && \
	code-server --install-extension bungcip.better-toml && \
	code-server --install-extension DavidAnson.vscode-markdownlint && \
	code-server --install-extension eamodio.gitlens && \
	code-server --install-extension eserozvataf.one-dark-pro-monokai-darker && \
	code-server --install-extension GitHub.vscode-pull-request-github && \
	code-server --install-extension HookyQR.beautify && \
	code-server --install-extension mauve.terraform && \
	code-server --install-extension ms-vscode.Theme-MaterialKit && \
	code-server --install-extension patrys.vscode-code-outline && \
	code-server --install-extension PeterJausovec.vscode-docker && \
	code-server --install-extension PKief.material-icon-theme && \
	code-server --install-extension ms-vscode.Go && \
	code-server --install-extension vscodevim.vim && \
	code-server --install-extension yzhang.markdown-all-in-one && \
	code-server --install-extension zhuangtongfa.Material-theme && \
	code-server --install-extension zxh404.vscode-proto3
# DO NOT EDIT. EXTENSIONS MANAGED BY SAIL.
`

var expectedExts = []string{
	"AlanWalk.markdown-toc",
	"akamud.vscode-theme-onedark",
	"anseki.vscode-color",
	"azemoh.one-monokai",
	"bierner.markdown-preview-github-styles",
	"bungcip.better-toml",
	"DavidAnson.vscode-markdownlint",
	"eamodio.gitlens",
	"eserozvataf.one-dark-pro-monokai-darker",
	"GitHub.vscode-pull-request-github",
	"HookyQR.beautify",
	"mauve.terraform",
	"ms-vscode.Theme-MaterialKit",
	"patrys.vscode-code-outline",
	"PeterJausovec.vscode-docker",
	"PKief.material-icon-theme",
	"ms-vscode.Go",
	"vscodevim.vim",
	"yzhang.markdown-all-in-one",
	"zhuangtongfa.Material-theme",
	"zxh404.vscode-proto3",
}

func Test_extensionBounds(t *testing.T) {
	var (
		expected = []int{2, 24}
		got      = make([]int, 2)
		err      error
	)

	got[0], got[1], err = extensionBounds(bytes.Split([]byte(testhat), []byte{10}))
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func Test_dockerfileAddExtensions(t *testing.T) {
	ext := []string{"ayy.lmao"}

	q, err := DockerfileAddExtensions([]byte(testhat), ext)
	require.NoError(t, err)

	got := extensionsFromDockerfile(bytes.Split([]byte(q), []byte{10}))
	assert.Equal(t, append(expectedExts, ext[0]), got)
}

func Test_dockerfileRemoveExtensions(t *testing.T) {
	ext := []string{"ayy.lmao"}

	q, err := DockerfileAddExtensions([]byte(testhat), ext)
	require.NoError(t, err)

	got := extensionsFromDockerfile(bytes.Split([]byte(q), []byte{10}))
	assert.Equal(t, append(expectedExts, ext[0]), got)
}

func Test_extensionsFromDockerfile(t *testing.T) {
	got := extensionsFromDockerfile(bytes.Split([]byte(testhat), []byte{10}))
	assert.Equal(t, expectedExts, got)
}

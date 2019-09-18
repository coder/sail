package codeserver

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDownloadURL_FetchAndRun(t *testing.T) {
	t.Parallel()

	url := DownloadURL(CodeServerVersion)

	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	if testing.Short() {
		t.Skipf("downloading code-server can take a while")
	}

	tmpfi, err := ioutil.TempFile("", "codeserver")
	require.NoError(t, err)
	defer os.Remove(tmpfi.Name())

	_, err = io.Copy(tmpfi, resp.Body)
	require.NoError(t, err)

	err = os.Chmod(tmpfi.Name(), 0750)
	require.NoError(t, err)

	err = tmpfi.Close()
	require.NoError(t, err)

	_, err = exec.Command(tmpfi.Name(), "--help").CombinedOutput()
	require.NoError(t, err)
}

package codeserver

import (
	"context"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestDownloadURL_Extract(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	url, err := DownloadURL(ctx)
	require.NoError(t, err)

	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	binRd, err := Extract(ctx, resp.Body)
	require.NoError(t, err)

	if testing.Short() {
		t.Skipf("downloading code-server can take a while")
	}

	tmpfi, err := ioutil.TempFile("", "codeserver")
	require.NoError(t, err)
	defer os.Remove(tmpfi.Name())

	_, err = io.Copy(tmpfi, binRd)
	require.NoError(t, err)

	err = os.Chmod(tmpfi.Name(), 0750)
	require.NoError(t, err)

	err = tmpfi.Close()
	require.NoError(t, err)

	_, err = exec.Command(tmpfi.Name(), "-h").CombinedOutput()
	require.NoError(t, err)
}

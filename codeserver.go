package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"golang.org/x/xerrors"

	"go.coder.com/flog"
	"go.coder.com/sail/internal/codeserver"
)

// ensureCodeServerCachePath determines the applicable cache directory path for
// this system, adds our suffix, does MkdirAll on it and returns it with
// /code-server appended to the end.
//
// Example output:
//     /home/dean/.cache/sail-code-server-cache/2.1485-vsc1.38.1/code-server.
func ensureCodeServerCachePath(codeServerVersion string) (string, error) {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", xerrors.Errorf("failed to determine user cache directory: %w", err)
	}

	cachePath := filepath.Join(userCacheDir, "sail-code-server-cache", codeServerVersion, "code-server")
	err = os.MkdirAll(filepath.Dir(cachePath), 0750)
	if err != nil {
		return "", err
	}

	return cachePath, nil
}

// loadCodeServer produces a path containing the code-server binary.
// It will attempt to cache the binary.
func loadCodeServer(ctx context.Context) (string, error) {
	const codeServerVersion = codeserver.CodeServerVersion

	cachePath, err := ensureCodeServerCachePath(codeServerVersion)
	if err != nil {
		return "", xerrors.Errorf("failed to ensure code-server cache path: %w", err)
	}

	// Only download the file if it doesn't exist in the cache directory.
	_, err = os.Stat(cachePath)
	if err == nil {
		return cachePath, nil
	}
	if err != nil && !xerrors.Is(err, os.ErrNotExist) {
		return "", xerrors.Errorf("failed to check if code-server is cached: %w", err)
	}

	start := time.Now()
	flog.Info("downloading and caching code-server %v", codeServerVersion)
	downloadURL := codeserver.DownloadURL(codeServerVersion)

	// We can't just overwrite the binary, as that would cause a `text file
	// busy` error if code-server is running. We write to a temporary path
	// first, and then atomically swap in this new file.
	tmpCachePath := cachePath + strconv.FormatInt(time.Now().UnixNano(), 10)
	defer os.Remove(tmpCachePath)

	cachedBinFi, err := os.OpenFile(tmpCachePath, os.O_CREATE|os.O_RDWR, 0750)
	if err != nil {
		return "", xerrors.Errorf("failed to open temporary file %v for writing: %w", err)
	}
	defer cachedBinFi.Close()

	binRd, err := http.Get(downloadURL)
	if err != nil {
		return "", xerrors.Errorf(`failed to download code-server from "%v": %w`, downloadURL, err)
	}
	defer binRd.Body.Close()

	_, err = io.Copy(cachedBinFi, binRd.Body)
	if err != nil {
		return "", xerrors.Errorf("failed to copy binary into %v: %w", tmpCachePath, err)
	}

	err = cachedBinFi.Close()
	if err != nil {
		return "", xerrors.Errorf("failed to close temporary file %v: %w", cachedBinFi.Name(), err)
	}

	// TODO: make this actually atomic.
	_ = os.Remove(cachePath)
	err = os.Rename(tmpCachePath, cachePath)
	if err != nil {
		return "", xerrors.Errorf("failed to rename %v to %v: %w", tmpCachePath, cachePath, err)
	}

	flog.Info("downloaded and cached code-server %v in %v", codeServerVersion, time.Since(start))

	return cachePath, nil
}

// codeServerPort gets the port of the running code-server binary.
//
// It will retry for 5 seconds if we fail to find the port in case
// the code-server binary is still starting up.
func codeServerPort(cntName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	var (
		port string
		err  error
	)

	for ctx.Err() == nil {
		if runtime.GOOS == "darwin" {
			// macOS uses port forwarding instead of host networking so netstat stuff below will not work
			// as it will find the port inside the container, which we already know is 8080.
			cmd := exec.CommandContext(ctx, "docker", "port", cntName, "8080")
			var out []byte
			out, err = cmd.CombinedOutput()
			if err != nil {
				continue
			}

			addr := strings.TrimSpace(string(out))
			_, port, err = net.SplitHostPort(addr)
			if err != nil {
				return "", xerrors.Errorf("invalid address from docker port: %q", string(out))
			}
		} else {
			port, err = codeserver.Port(cntName)
			if xerrors.Is(err, codeserver.PortNotFoundError) {
				continue
			}
			if err != nil {
				return "", err
			}
		}

		var resp *http.Response
		resp, err = http.Get("http://localhost:" + port)
		if err == nil {
			resp.Body.Close()
			return port, nil
		}

		time.Sleep(time.Millisecond * 100)
	}

	return "", xerrors.Errorf("failed while trying to find code-server port: %w", err)
}

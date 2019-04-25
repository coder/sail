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
	"strings"
	"time"

	"go.coder.com/flog"
	"go.coder.com/sail/internal/codeserver"
	"golang.org/x/xerrors"
)

// loadCodeServer produces a path containing the code-server binary.
// It will attempt to cache the binary.
func loadCodeServer(ctx context.Context) (string, error) {
	start := time.Now()

	const cachePath = "/tmp/sail-code-server-cache/code-server"

	// Only check for a new codeserver if it's over an hour old.
	info, err := os.Stat(cachePath)
	if err == nil {
		if info.ModTime().Add(time.Hour).After(time.Now()) {
			return cachePath, nil
		}
	}

	u, err := codeserver.DownloadURL(ctx)
	if err != nil {
		return "", err
	}

	err = os.MkdirAll(filepath.Dir(cachePath), 0750)
	if err != nil {
		return "", err
	}

	fi, err := os.OpenFile(cachePath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0750)
	if err != nil {
		if os.IsExist(err) {
			return cachePath, nil
		}
		return "", err
	}
	defer fi.Close()

	tarFi, err := http.Get(u)
	if err != nil {
		os.Remove(cachePath)
		return "", xerrors.Errorf("failed to get %v: %w", u, err)
	}
	defer tarFi.Body.Close()

	binRd, err := codeserver.Extract(ctx, tarFi.Body)
	if err != nil {
		os.Remove(cachePath)
		return "", xerrors.Errorf("failed to untar %v: %w", u, err)
	}

	_, err = io.Copy(fi, binRd)
	if err != nil {
		os.Remove(cachePath)
		return "", xerrors.Errorf("failed to copy binary into %v: %w", cachePath, err)
	}

	flog.Info("loaded code-server in %v", time.Since(start))

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
			// as it will find the port inside the container, which we already know is 8443.
			cmd := exec.CommandContext(ctx, "docker", "port", cntName, "8443")
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

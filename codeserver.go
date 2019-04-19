package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

	u, err := codeserver.DownloadURL(ctx)
	if err != nil {
		return "", err
	}

	cachePath := filepath.Join(
		"/tmp/sail-code-server-cache",
		u[strings.LastIndex(u, "/"):],
		"code-server",
	)

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
		port, err = codeserver.Port(cntName)
		if err == nil {
			return port, nil
		}

		if !xerrors.Is(err, codeserver.PortNotFoundError) {
			return "", err
		}

		time.Sleep(time.Millisecond * 100)
	}

	return "", xerrors.Errorf("failed while trying to find code-server port: %w", err)
}

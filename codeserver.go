package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"github.com/google/go-github/v24/github"
	"go.coder.com/flog"
	"golang.org/x/xerrors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// codeServerExtractScript is a shell script that extracts code-server release file (first argument).
// to /usr/bin/code-server.
const codeServerExtractScript = `
set -euf

dir=$(mktemp -d)

wget -O - $URL 2>/dev/null | tar -xzf - -C $dir
fullpath=$(find $dir -name code-server | head -n 1)

binpath=/usr/bin/code-server

if [ -a $binpath ]; then
	rm $binpath
fi

ln $fullpath $binpath
`

// codeServerDownloadURL gets a URL for the latest version of code-server.
func codeServerDownloadURL(ctx context.Context) (string, error) {
	client := github.NewClient(nil)
	rel, _, err := client.Repositories.GetLatestRelease(ctx, "codercom", "code-server")
	if err != nil {
		return "", xerrors.Errorf("failed to get latest code-server release: %w", err)
	}
	for _, v := range rel.Assets {
		// TODO: fix this jank.
		if strings.Index(*v.Name, "linux") < 0 {
			continue
		}
		return *v.BrowserDownloadURL, nil
	}
	return "", errors.New("no released found for platform")
}

// extractCodeServer takes a code-server release tar and writes out the main binary to bin.
func extractCodeServer(ctx context.Context, tarFi io.Reader) (io.Reader, error) {
	grd, err := gzip.NewReader(tarFi)
	if err != nil {
		return nil, xerrors.Errorf("failed to create gzip decoder: %w", err)
	}
	defer grd.Close()

	rd := tar.NewReader(grd)
	for {
		hdr, err := rd.Next()
		if err != nil {
			if err == io.EOF {
				return nil, errors.New("code-server not found")
			}
			return nil, err
		}
		if filepath.Base(hdr.Name) == "code-server" {
			return rd, nil
		}
	}
}

// loadCodeServer produces a path containing the code-server binary.
// It will attempt to cache the binary.
func loadCodeServer(ctx context.Context) (string, error) {
	start := time.Now()

	u, err := codeServerDownloadURL(ctx)
	if err != nil {
		return "", err
	}

	cachePath := filepath.Join(
		"/tmp/narwhal-code-server-cache",
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

	binRd, err := extractCodeServer(ctx, tarFi.Body)
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

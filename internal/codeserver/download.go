package codeserver

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v24/github"
	"golang.org/x/xerrors"
)

// DownloadURL gets a URL for the latest version of code-server.
func DownloadURL(ctx context.Context) (string, error) {
	client := github.NewClient(nil)
	rel, _, err := client.Repositories.GetLatestRelease(ctx, "cdr", "code-server")
	if err != nil {
		return "", xerrors.Errorf("failed to get latest code-server release: %w", err)
	}
	for _, v := range rel.Assets {
		// TODO: fix this jank, detect container architecture instead of hardcoding to x86_64
		if strings.Index(*v.Name, "linux-x86_64") < 0 {
			continue
		}
		return *v.BrowserDownloadURL, nil
	}
	return "", xerrors.New("no released found for platform")
}

// Extract takes a code-server release tar and writes out the main binary to bin.
func Extract(ctx context.Context, tarFi io.Reader) (io.Reader, error) {
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
				return nil, xerrors.New("code-server not found")
			}
			return nil, err
		}
		if filepath.Base(hdr.Name) == "code-server" {
			return rd, nil
		}
	}
}

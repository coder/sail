package main

import (
	"context"
	"errors"
	"github.com/google/go-github/v24/github"
	"golang.org/x/xerrors"
	"strings"
)

// CodeServerExtractScript is a shell script that extracts code-server release file (first argument).
// to /usr/bin/code-server.
const CodeServerExtractScript = `
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

// CodeServerDownloadURL gets a URL for the latest version of code-server.
func CodeServerDownloadURL(ctx context.Context, os string) (string, error) {
	client := github.NewClient(nil)
	rel, _, err := client.Repositories.GetLatestRelease(ctx, "codercom", "code-server")
	if err != nil {
		return "", xerrors.Errorf("failed to get latest code-server release: %w", err)
	}
	for _, v := range rel.Assets {
		// TODO: fix this jank.
		if strings.Index(*v.Name, os) < 0 {
			continue
		}
		return *v.BrowserDownloadURL, nil
	}
	return "", errors.New("no released found for platform")
}

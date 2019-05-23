#!/bin/bash

set -euxo pipefail

mkdir -p /tmp/sail-code-server-cache
curl  > /tmp/sail-code-server-cache/code-server https://codesrv-ci.cdr.sh/latest-linux
chmod +x /tmp/sail-code-server-cache/code-server

go test -v ./...
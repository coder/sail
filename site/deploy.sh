#!/bin/bash

set -euxo pipefail

hugo --minify
gsutil -m cp -R ./public/* "gs://sail.dev"
gsutil -m setmeta -h "Content-Type:text/html" \
  -h "Cache-Control:private, max-age=0, no-transform" "gs://sail.dev/**.html"
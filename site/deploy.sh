#!/bin/bash

set -euxo pipefail

# Remove previously generated bullshit.
rm -rf public/
rm -rf resources/_gen

hugo --minify
gsutil -m cp -R ./public/* "gs://sail.dev"
gsutil -m setmeta -h "Content-Type:text/html" \
  -h "Cache-Control:private, max-age=0, no-transform" "gs://sail.dev/**.html"

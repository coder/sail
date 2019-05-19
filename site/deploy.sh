#!/bin/bash

set -euxo pipefail

hugo --minify
gsutil -m cp -R ./public/* "gs://sail.dev"
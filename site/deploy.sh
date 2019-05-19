#!/bin/bash

set -euxo pipefail

hugo
gsutil -m cp -R ./public/* "gs://sail.dev"
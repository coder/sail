#!/bin/bash
set -eu

BASE_IMAGE=ubuntu-dev

# Build the base for all images.
pushd base
    docker build -t sail-base --label com.coder.sail.base_image=sail-base .
popd

# Build our base ubuntu-dev image for non language specific environments.
pushd $BASE_IMAGE
    ./buildpush.sh
popd
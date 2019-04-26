#!/bin/bash
set -eu

BASE_IMAGE=ubuntu-dev

LANG_IMAGES=(
    ubuntu-dev-go
    ubuntu-dev-python2.7
    ubuntu-dev-python3.7
    ubuntu-dev-ruby2.6
    ubuntu-dev-gcc8
    ubuntu-dev-node12
    ubuntu-dev-openjdk12
)

# Build our base image for non language specific environments.
pushd $BASE_IMAGE
    ./buildpush.sh
popd

# Build all our language specific environments.
for lang in "${LANG_IMAGES[@]}"; do
    ./buildpush.sh $lang
done

#!/bin/bash
set -eu

DIR=$1
IMG_NAME=codercom/$DIR

pushd $DIR

# Build our image.
../buildlang.sh $IMG_NAME

# Push to Docker Hub.
../push.sh $IMG_NAME

popd

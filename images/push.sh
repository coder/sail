#!/bin/bash
set -eu

IMAGE_NAME=$1

docker push $IMAGE_NAME

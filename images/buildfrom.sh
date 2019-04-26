#!/bin/bash
set -eu

DOCKERFILE=$1
FROM=$2
IMAGE_NAME=$3

# Change the FROM value to the passed in value, then build the image from stdin.
sed "s#%BASE#$FROM#g" $DOCKERFILE | docker build -t $IMAGE_NAME --label com.coder.sail.base_image=$IMAGE_NAME -f- .

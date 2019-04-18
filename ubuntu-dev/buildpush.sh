#!/bin/bash
set -eu

docker build --network=host -t $1 . --label com.coder.sail.base_image=$1
docker push $1

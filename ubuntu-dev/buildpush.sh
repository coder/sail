#!/bin/bash
set -eu

docker build --network=host -t $1 .
docker push $1

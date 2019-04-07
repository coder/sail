#!/bin/bash
set -e

docker build --network=host -t codercom/ubuntu-dev .
docker push codercom/ubuntu-dev

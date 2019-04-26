#!/bin/bash
set -eu 

DOCKERFILE=../Dockerfile
FROM=buildpack-deps:cosmic
IMAGE_NAME=codercom/ubuntu-dev


../buildfrom.sh $DOCKERFILE $FROM $IMAGE_NAME
../push.sh $IMAGE_NAME
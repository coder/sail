#!/bin/bash
set -eu 

DOCKERFILE=Dockerfile
FROM=sail-base
IMAGE_NAME=codercom/ubuntu-dev

../buildfrom.sh $DOCKERFILE $FROM $IMAGE_NAME
../push.sh $IMAGE_NAME
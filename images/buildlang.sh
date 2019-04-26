#!/bin/bash
set -eu

LANG_IMG_NAME=$1

COMM_DOCKERFILE=./Dockerfile.comm
COMM_FROM=buildpack-deps:cosmic
COMM_IMG_NAME=$LANG_IMG_NAME-comm

# Build our community base image.
#
# Since this script is called from the languages directory, we inherit that
# location, so we need to backout to access buildfrom.sh.
../buildfrom.sh $COMM_DOCKERFILE $COMM_FROM $COMM_IMG_NAME

BASE_DOCKERFILE=../Dockerfile
BASE_IMG_NAME=$LANG_IMG_NAME-base

# Add the ubuntu-dev base on top.
../buildfrom.sh $BASE_DOCKERFILE $COMM_IMG_NAME $BASE_IMG_NAME


LANG_DOCKERFILE=./Dockerfile.lang

# Now add any language specific extensions or tooling.
../buildfrom.sh $LANG_DOCKERFILE $BASE_IMG_NAME $LANG_IMG_NAME


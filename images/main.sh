#!/bin/bash
set -eu


LANG_IMAGES=(
    ubuntu-dev-gcc8
    ubuntu-dev-go
    ubuntu-dev-llvm8
    ubuntu-dev-node12
    ubuntu-dev-openjdk12
    ubuntu-dev-python2.7
    ubuntu-dev-python3.7
    ubuntu-dev-ruby2.6
)

./buildbase.sh


# Build all our language specific environments.
for lang in "${LANG_IMAGES[@]}"; do
    ./buildpush.sh $lang
done

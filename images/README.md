# Images

Sail's base image is the Dockerfile in this directory with the FROM clause changed to
buildpack-deps:cosmic.

The Sail base images are all based off of the buildpack-deps:cosmic image.

In order for the Sail images to have lower maintenance and minimal language 
specific knowledge, each language image is based on the open source community image
and expanded upon to add any Sail dependencies.

## Extending Community Images

Most community language images have an image based on buildpack-deps:stretch. Since we
want to use a more up to date ubuntu base, the Dockerfile from these community images
should be copied into their corresponding `ubuntu-dev-<lang>` directory and modified
to properly build with buildpack-deps:cosmic.

This modified community Dockerfile should be placed into the languages directory as
`Dockerfile.comm`. The source of the community image and any modifications to the image
should be explicitly stated in a comment at the top of this Dockerfile.

After building this community Dockerfile image, we can then use that image as the FROM clause
to build the Sail base which installs all the Sail dependencies on top of the language
configuration.

If a language has vscode extensions or other developer tooling that should be installed,
it should be placed into a dockerfile named `Dockerfile.lang` in the correct language directory.

We can then use the Sail base image that was just created on top of the language base as the
FROM clause in this language extensions and tooling image.

After building this image, we should have everything we need for a proper Sail language
environment and can push this image up to the codercom Docker Hub.


## Build

A quick overview of the different scripts and what they do:

`main.sh` - Iterates through all of the images, builds them, and pushes them to the codercom docker hub.

`push.sh` - Pushes a built image to the codercom docker hub.

`buildfrom.sh` - This changes the `FROM %BASE` in a Dockerfile to the argument that's provided. It uses
the provided name to tag the image and add the sail base_image label to the image.

`buildlang.sh` - This should be called from within a language directory. It first calls `buildfrom.sh` on
`Dockerfile.lang` to create the first intermediate image. It then calls `buildfrom.sh` using the
community language image as the base to build the sail base image. It calls `buildfrom.sh` a third and final
time on `Dockerfile.lang` using the language extended sail base image to install any language specific vscode
extensions or tooling.

`buildpush.sh` - This script takes an image name, i.e. `ubuntu-dev-go1.12`, changes into the specified directory,
and calls `buildlang.sh` and `push.sh` to build the language image and push the finalized image to the codercom
docker hub.
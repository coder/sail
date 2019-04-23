+++
type="doc"
title="Immutability"
browser_title="Sail - Docs - Immutability"
draft=true
+++

Sail is about explicitely describing your development environment in a way
that's easy to share and iterate on.

If you want to make a change to the environment, modify your hat or the project's
Dockerfile.

### Project-defined environment

Projects can define their own environment by specifying a `.sail/Dockerfile` file.

The dev container must have `codercom/ubuntu-dev` as a parent.

The build will occur in the repo's root directory.
+++
type="docs"
title="Docker"
browser_title="Sail - Docs - Docker"
section_order=4
+++

Sail can be thought of as a wrapper around the Docker toolchain, focused
on managing development environments.

It stores most of it's state in the Docker daemon
in the form of metadata and [labels](/docs/concepts/labels/).


## Immutability

Sail tries to encourage a workflow of explicitly describing your development environment in a way
that's easy to share and iterate on.

If you want to make a change to the environment, you should modify your [hat](/docs/concepts/hats/) or the [project's](/docs/concepts/projects/)
Dockerfile.

## Project Defined Environment

Projects can define their own environment by specifying a `.sail/Dockerfile` file.
If a `.sail/Dockerfile` is not present in a repository, the `codercom/ubuntu-dev`
image will be used as the base environment.

When specifying a custom project environment, the dev container must have
be an ancestor of `codercom/ubuntu-dev` in order to have the proper dependencies
setup.

When building the project's image, the build will be rooted in the project's root,
essentially calling this docker build command:

```bash
 docker build -f $project_root/<org>/<repo>/.sail/Dockerfile $project_root/<org>/<repo>
```

## Container Permissions

The current user on the host is mapped to the user named `user` within
the development environment.  That means files permissioned to `user` within
the container, will appear to have the same set of permissions for your user on the host.


Sail uses `user` and not `root` within the container because:

- A lot of tools will complain about being root.
- Most developers are used to being non-root and the `sudo` workflow.

## Container Naming

Containers are named `<org>_<project>` in Docker, but `<org>/project` in Sail.

## Networking

To keep workflow as close to local development as possible, sail uses docker
host networking when possible. That means if your webserver within Sail binds
to `:8080`, it will be accessible from `127.0.0.1:8080` in your browser.

Docker for Mac doesn't support host networking, so this won't work when running
Sail on a Mac host. A workaround is planned for a future release of Sail.

## Dockerfile Best Practices

[Dockerfile best practices](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/) 
for a reference on how to properly structure and write your [project](/docs/concepts/projects/) and [hat](/docs/concepts/hats/)
Dockerfiles.

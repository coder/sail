+++
type="docs"
title="Projects"
browser_title="Sail - Docs - Projects"
section_order=0
+++

One of Sail's core concepts is the project. A project can be thought of as a source code
repository that is contained in an environment with all of its dependencies and required
configuration.

## Dependency and Configuration

A source code repository can specify what dependencies are required by creating a `.sail/Dockerfile`
file in the root of the repo. Sail will build a docker image from this Dockerfile and run the project
inside a container running that environment.

If a project doesn't have a `.sail/Dockerfile` file, then the default
[codercom/ubuntu-dev](https://hub.docker.com/r/codercom/ubuntu-dev) image will be used to run the
project's environment.


## Persistence

Since container filesystems are ephemeral by default, Sail clones the project's repository onto
the host at Sail's `$project_root` and bind mounts it into the container. 

Since the projects are bind mounted into the container, deleting a container does not delete project files
and you can seamlessly interact with project files outside of the container.

### Host View of the Project

The `$project_root` is an environment variable that can be set in Sail's global configuration 
file, but by default it's located at `~/Projects`. Projects are cloned into the `$project_root`
in a structure like so:
```
$project_root/<org>/<repo>
```

For example, if you were to start a new sail environment to work on sail:
```bash
sail run codercom/sail
```
It would be cloned to `$project_root/codercom/sail`.

### Container View of the Project

By default, the project is bind mounted inside of the container to `~/<repo>`

To enable some special-case languages such as Go, the bind mount target location 
can be configured via the project_root label in your project's `.sail/Dockerfile`. 

For example:

```Dockerfile
LABEL project_root "~/go/src/"
```

Will bind mount the host directory `$project_root/<org>/<repo>` to `~/go/src/<repo>` in the container.

## Configuration

Configuring a project is done through common [Dockerfile](https://docs.docker.com/engine/reference/builder/) commands.

For example, if your project has autotools as a dependency, you could install that into your environment through the
project's `.sail/Dockerfile` like so:

```Dockerfile
FROM codercom/ubuntu-dev

RUN apt-get update && apt-get install -y \
    autoconf \
    automake \
    libtool
```

For specifying things like bind mounts and where a project should be bind mounted to, Sail artificially
extends the Dockerfile syntax through labels as seen above in the [Container View of the Project](#container-view-of-the-project).

Sail labels will be described further in [Labels](/docs/concepts/labels).

### Developer Configuration

As a developer, you'll want to bring your own configurations and tooling when working on a project. You can easily 
extend any project's environment through the use of a [hat](/docs/concepts/hats). A hat allows you to install your own 
configurations and tooling on top of each project's environment through a hat Dockerfile so that you don't have
to leave the workflow you're used to behind.


## Supported Version Control Systems

Currently Sail only supports git.


+++
type="doc"
title="Docker"
browser_title="Sail - Docs - Docker"
draft=true
+++

Sail can be thought of as a wrapper around the Docker toolchain, focused
on managing development environments.

It stores most of it's state in the Docker daemon
in the form of metadata and labels.

Informational Sail Docker labels begin with `com.coder.sail`.

## Command labels

Sail artificially extends the Dockerfile syntax through labels.

For example:

```Dockerfile
LABEL project_root "~/go/src"
```

Will instruct Sail to use `~/go/src` as the project root.

We call these kinds of labels _command labels_. Each command label is introduced
throughout these docs.

## Container Naming

Containers are named `<org>-<project>` in Docker, but `<org>/project` in Sail.

## Networking

### MacOS

On MacOS hosts, the Docker container is given a bridge network.

### Linux

To keep workflow as close to local development as possible, we give
the container the host network. That means if your webserver within Sail binds
to `:8080`, `127.0.0.1:8080` on the host will work just fine.


## Dockerfile Best Practices

[Official Dockerfile best practices.](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)

---

Maximize layer caching by seperating command as much as possible:

**Bad**
```
RUN sudo apt install -y htop dstat dog
```

**Good**
```
RUN sudo apt install -y htop
RUN sudo apt install -y dstat
RUN sudo apt install -y dog
```

As you become aware of what installations the project _always_ needs, we're
better off merging the commands and moving them to the beginning of the Dockerfile.

**Tip**: whenever changing or adding a command, always do it at the bottom of
the Dockerfile. The Dockerfile will approach it's layer-optimized version.


---
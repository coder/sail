# Goals

The primary goal of this branch is to reduce the reliance on the local machine
as much as possible. By doing so, `sail` provides an abstraction about where the
underlying container is being ran, and where all the data actually lives.

By having this abstraction, `sail` can truly be editor agnostic. All that would
be required would be the editor's ability to interface with docker (e.g with
Emacs' TRAMP or VSCode's Docker extension). `code-server` will remain an option.

# Required changes

## Mounting

All bind mounts will need to be removed, since we will not be able to rely on
the where the docker container is running.

Mounting should be done with docker volumes since volumes provide an abstraction
as to where the data is located.

### Volume drivers

For data that's expected to be local to one machine, but should persist during
container rebuilds, the `local` (default) volume driver should be used.

For data that will be shared between many containers, some other yet to be
determined driver should be used (potentially something similar to
[flocker](https://github.com/ClusterHQ/flocker)).


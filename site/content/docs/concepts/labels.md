+++
type="docs"
title="Labels"
browser_title="Sail - Docs - Labels"
section_order=3
+++

Sail makes extensive use of Docker labels to maintain state and to allow users to fully configure
their project environments.

## Configuration Labels

### Project Root Label

As described in [projects](/docs/concepts/projects/), the bind mount target of the project's root can be specified
using the `project_root` label. By default the project root is bind mounted to `~/<repo>` inside of
the container.

For example:

```Dockerfile
LABEL project_root "~/go/src/"
```

Will bind mount the host directory `$project_root/<org>/<repo>` to `~/go/src/<repo>` in the container.

### Run on Open Labels

You can run a command in your sail container after it starts by specifying
the `on_open` label. If you'd like to run multiple commands on launch, we
recommend using a `.sh` file as your `on_open` label, as you cannot provide
multiple `on_open` statements.

The `on_open` label is run detached inside of `/bin/sh` as soon as the
container is started, with the work directory set to your `project_root`
(see the section above).

For example:
```Dockerfile
LABEL on_open "npm install"
```
```Dockerfile
LABEL on_open "go get"
```
```Dockerfile
LABEL on_open "./.sail/on_open.sh"
```

Make sure any scripts you make are executable, otherwise sail will fail to
launch.

### Share Labels

A sail share is a directory on the host that you want shared with your
sail container. Any shares that are specified within the Project or hat
Dockerfiles will be bind mounted to the proper location inside of the
container.

Projects and hats can specify shares using command labels of the form:

`share.<share_name>="host_path:guest_path"`.

For example, if you wanted to share your go mod cache with your container
you would add this to your project or hat Dockerfile:

```Dockerfile
LABEL share.go_mod="~/go/pkg/mod:~/go/pkg/mod"
```

---

Shares are recommended for

- Filesystem-level caches
    - Go mod cache
    - Yarn cache
- User-specific configuration
    - VS Code configuration (auto)
    - SSH keys
    - gitconfig
- Working data
    - Project files (auto)
    - Data analysis results

It's important to keep in mind that shares can easily undermine the
reproducibility and consistency of your environments. Be careful with blanket shares
such as `~:~` which introduce variance.

## State Labels

Sail uses Docker labels that begin with `com.coder.sail` to manage any state
that the CLI may need. These labels are only required by the Sail CLI and aren't
useful for user configuration.

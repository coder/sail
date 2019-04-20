+++
type="doc"
title="Projects"
browser_title="Sail - Docs - Projects"
+++

sail enforces that projects are stored on the host filesystem at `$project_root/<org>/<repo>`.

`$project_root` is a configuration variable.

Projects are stored in the container at `~/<repo>`.

To enable some special-case languages such as Go, the project root can be configured
via the project_root label. For example:

```Dockerfile
LABEL project_root "~/go/src/"
```

Projects are bind-mounted into the container, so deleting a container does not delete project files
and you can seamlessly interact with project files outside of the container.

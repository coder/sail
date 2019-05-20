+++
type="docs"
title="Project Extensions"
browser_title="Sail - Docs - Project Extensions"
section_order=5
+++

Installing VS Code extensions through your Sail Dockerfile is dead-simple if
you're image is based from `ubuntu-dev`.

In your Dockerfile, call `install_ext <extension ID>`.

For example:

```Dockerfile
FROM ubuntu-dev
RUN install_ext vscodevim.vim
```

_Tip: You can find the extension ID at the extension's page._

![Extension ID in VS Code](/extension-id.png)

## Under The Hood

`code-server` is started with two extension directories:

- `~/.vscode/extensions` contains extensions for the specific environment.
- `~/.vscode/host-extensions` is bind-mounted in from `~/.vscode/extensions` on
the host.

This ensures that

1. Projects can specify their extensions.
1. Users continue using the extensions that they installed through native
VS Code.


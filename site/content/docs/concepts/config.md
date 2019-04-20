+++
type="doc"
title="Configuration"
browser_title="Sail - Docs - Configuration"
+++

Sail is about moving configuration into controlled projects and hats, so naturally
it strives to keep it's configuration minimal. Sail accepts a variety of flags
through it's commands, but there is a little bit global configuration at

`~/.config/sail/sail.toml`.

The self-documenting default configuration for convenience:

```toml
# default_image is the default Docker image to use if the repository provides none.
default_image = "codercom/ubuntu-dev"

# project_root is the base from which projects are mounted.
# projects are stored in directories with form "<root>/<org>/<repo."
project_root = "~/Projects"

# default hat lets you configure a hat that's applied automatically by default.
# default_hat = ""
```
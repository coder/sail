+++
type="docs"
title="Configuration"
browser_title="Sail - Docs - Configuration"
section_order=5
+++

Sail is about moving configuration into controlled projects and hats, so naturally
it strives to keep it's configuration minimal. Sail accepts a variety of flags
through it's commands, but there is a little bit of global configuration located at:

`~/.config/sail/sail.toml`.

Here is the default self-documenting configuration for reference:

```toml
# default_image is the default Docker image to use if the repository provides none.
default_image = "codercom/ubuntu-dev"

# project_root is the base from which projects are mounted.
# projects are stored in directories with form "<root>/<org>/<repo."
project_root = "~/Projects"

# default hat lets you configure a hat that's applied automatically by default.
# default_hat = ""
```

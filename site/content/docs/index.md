+++
type="doc"
title="Getting Started"
browser_title="Sail - Docs - Getting Started"
+++
Welcome to the the Sail docs.

## Install Stable Release

Binary releases can be found on our [GitHub.](https://github.com/codercom/sail/releases)

## Install From Source

To install `sail` the latest version of sail, run:

```bash
go install go.coder.com/sail
```

_`go install` will install to ~/go/bin_

## First steps

Spin up a secure editor for `codercom/sail`.

```bash
sail run codercom/sail
# Creates a Docker container called `codercom-sail`,
# installs code-server in it, and creates a browser.
```

## Updating

To gracefully update `sail`, simply overwrite the binary with the binary 
in the new release.

If you installed via `go install`, just run the same command again.
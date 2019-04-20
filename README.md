# sail

`sail` is a CLI to efficiently manage Dockerized [`code-server`](https://github.com/codercom/code-server) development environments.
.

## Features

- Projects can specify their own development environment via `.sail/Dockerfile`.
- Shares VS Code settings between environments.
  - Syncs with local VS Code as well.
- Supports Linux and MacOS.
- Native GitHub support.

## Install

```bash
go get go.coder.com/sail
```

## Basic usage

Spin up a secure editor for `codercom/sail`.

Or:

```bash
sail run codercom/sail
# Creates a Docker container called `codercom-sail`,
# installs code-server in it, and creates a browser.
```

## Documentation

Documentation is available in markdown form at [site/content/docs.](site/content/docs)

Or, you can find it at [https://sail.dev/docs](https://sail.dev/docs)

## Future

These features are planned for future releases:

- Windows support
- Synchronizing code-server extensions, settings, and themes.
- Remote Host support.
  - Cloud integration so `my-compute-instance` can be auto-resolved into an AWS/GCP/Azure instance.

+++
type="docs"
title="Installation"
browser_title="Sail - Docs - Installation"
section_order=1
+++

## Platform Support

Currently Sail supports both Linux and MacOS. Windows support is planned for a future release.

## Host Dependencies

Before using Sail, there are several dependencies that must be installed on the host system:

- [Docker](https://docs.docker.com/install/)
- [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
- [Chrome](https://www.google.com/chrome/) or [Chromium](https://www.chromium.org/getting-involved/download-chromium) - not required, but strongly recommended for best [code-server](https://github.com/cdr/code-server) support.


## Installation

### Stable Releases

It's recommended that user's install the sail binary from the stable releases.

Binary releases can be downloaded from our [GitHub.](https://github.com/cdr/sail/releases)

### From Source

To install the latest version of `sail`, you'll need [go](https://golang.org/) installed and configured on your system. 

Sail uses go modules to build the project, so the easiest way to install it to your system is to clone it in a directory
outside of your `GOPATH`.

```
mkdir $HOME/src
cd $HOME/src
git clone https://github.com/cdr/sail.git
cd sail
go install
```


### Verifying the Installation

To verify Sail is properly installed, run `sail --help` on your system. If everything is installed
properly, you should see Sail's help text.

```bash
sail --help
```

## Browser Extension

In order to have an optimal experience while using Sail, we recommend [installing the browser extension](/docs/browser-extension/).


## Updating

To gracefully update `sail`, simply overwrite the binary with the binary 
in the new release.

If you installed via `go install`, just run the same command again.

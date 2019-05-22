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
- [Chrome](https://www.google.com/chrome/) or [Chromium](https://www.chromium.org/getting-involved/download-chromium) - not required, but strongly recommended for best [code-server](https://github.com/cdr/code-server) support. If chrome is not installed, the default browser will be used.

## Installation

For simple, secure and fast installation, the following command will install the latest version
of sail for your OS and architecture into `/usr/local/bin`. You will need to have `/usr/local/bin`
in your [$PATH](https://superuser.com/questions/284342/what-are-path-and-other-environment-variables-and-how-can-i-set-or-use-them) in order to use it.

```bash
curl https://sail.dev/install.sh | bash
```

### Stable Releases

You can also manually install from the [github releases](https://github.com/cdr/sail/releases) and
place the binary wherever you want.

### From Source

For more **advanced users** who want to install the latest version from master, you can install Sail from source.

You'll need the [go programming language](https://golang.org/) installed and configured on your machine, and `$GOPATH/bin`
added to your [PATH](https://superuser.com/questions/284342/what-are-path-and-other-environment-variables-and-how-can-i-set-or-use-them) for
the following to work correctly.

Sail uses go modules to build the project, so the easiest way to install it to your system is to clone it in a directory
outside of your `GOPATH`.

```
mkdir $HOME/src
cd $HOME/src
git clone https://github.com/cdr/sail.git
cd sail
go install
```


## Verifying the Installation

To verify Sail is properly installed, run `sail --help` on your system. If everything is installed
properly, you should see Sail's help text.

```bash
sail --help
```

## Browser Extension

In order to have an optimal experience while using Sail, we recommend [installing the browser extension](/docs/browser-extension/).


## Updating

Just reinstall with whatever method you installed with.

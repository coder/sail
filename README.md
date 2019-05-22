# sail

[!["Open Issues"](https://img.shields.io/github/issues-raw/cdr/sail.svg)](https://github.com/cdr/sail/issues)
[![MIT license](https://img.shields.io/badge/license-MIT-green.svg)](https://github.com/cdr/sail/blob/master/LICENSE)
[![Discord](https://img.shields.io/discord/463752820026376202.svg?label=&logo=discord&logoColor=ffffff&color=7389D8&labelColor=6A7EC2)](https://discord.gg/zxSwN8Z)

`sail` is a universal workflow for reproducible, project-defined development environments.

Basically, it lets you open a repo in a VS Code window with a Docker-based backend.

With the browser extension, you can open a repo right from GitHub or GitLab, or
you can do

```
sail run cdr/sshcode
```

to open a project right from the command line.

**[Browser extension](https://sail.dev/docs/concepts/browser-extension/) demo:**

![Demo](/site/demo.gif)

## Features

- **No more "It works on my machine"**, everyone working on the same project is working in the same environment.
- **Stop duplicating effort**, source-control and collaborate on the environment.
- **Instant set-up**, open an IDE for a project straight from GitHub or GitLab.

## Documentation

Documentation is available at [https://sail.dev/docs](https://sail.dev/docs/introduction/). 

Or, you can read it in it's markdown form at [site/content/docs.](site/content/docs)

## Quick Start

### Requirements

**Currently Sail supports both Linux and MacOS. Windows support is planned for a future release.**

Before using Sail, there are several dependencies that must be installed on the host system:

- [Docker](https://docs.docker.com/install/)
- [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
- [Chrome](https://www.google.com/chrome/) or [Chromium](https://www.chromium.org/getting-involved/download-chromium) - not required, but strongly recommended for best [code-server](https://github.com/cdr/code-server) support.
If chrome is not installed, the default browser will be used.


### Install

For simple, secure and fast installation, the following command will install the latest version
of sail for your OS and architecture into `/usr/local/bin`. You will need to have `/usr/local/bin`
in your [$PATH](https://superuser.com/questions/284342/what-are-path-and-other-environment-variables-and-how-can-i-set-or-use-them) in order to use it.

```
curl https://sail.dev/install.sh | bash
```

### Verify the Installation

To verify Sail is properly installed, run `sail --help` on your system. If everything is installed correctly, you should see Sail's help text.

### Run

You should now be able to run `sail run cdr/sail` from your terminal to start an environment designed for working
on the Sail repo.

### Browser Extension

To open GitHub or GitLab projects in a Sail environment with a single click, see the [browser extension install instructions](https://sail.dev/docs/browser-extension/).

### Learn More

Additional docs covering concepts and configuration can be found at [https://sail.dev/docs](https://sail.dev/docs/introduction/).

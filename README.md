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

## Install

```bash
curl https://sail.dev/install.sh | bash
```

## Features

- **No more "It works on my machine"**, everyone working on the same project is working in the same environment.
- **Stop duplicating effort**, source-control and collaborate on the environment.
- **Instant set-up**, open an IDE for a project straight from GitHub or GitLab.

## Documentation

Documentation is available at [https://sail.dev/docs](https://sail.dev/docs/introduction/). 

Or, you can read it in it's markdown form at [site/content/docs.](site/content/docs)

# sail

`sail` is a CLI to efficiently manage Dockerized [`code-server`](https://github.com/codercom/code-server) development environments.
.

# Features

- Projects can specify their own development environment via `.sail/Dockerfile`.
- Shares VS Code settings between environments.
	- Syncs with local VS Code as well.
- Supports Linux and MacOS.
- Native GitHub support.

# Install

```
go install go.coder.com/sail
```

# Basic usage

Spin up a secure editor for `codercom/sail`.

Or:

```bash
sail run codercom/sail
# Creates a Docker container called `codercom-sail`,
# installs code-server in it, and creates a browser.
```

# Projects

sail enforces that projects are stored on the host filesystem at `$project_root/<org>/<repo>`.

`$project_root` is a configuration variable.

Projects are stored in the container at `~/<repo>`.

To enable some special-case languages such as Go, the project root can be configured
via the project_root label. For example:

```
LABEL project_root "~/go/src/"
```

Projects are bind-mounted into the container, so deleting a container does not delete project files
and you can seamlessly interact with project files outside of the container.

# Environments

## Project-defined environment 

Projects can define their own environment by specifying a `.sail/Dockerfile` file.

The dev container must have `codercom/ubuntu-dev` as a parent.

The build will occur in the repo's root directory.

## Permissions

The user that creates the container has their Uid mapped to the `user` user within the container.

This ensures that newly created project files have the correct permissions on 
the host.

## Live modification

The sail workflow promotes a unique environment for each project, with common
configurations explicitely declared.


The workflow for modifying an environment goes like:

1) Have code-server open in some window.
1) Have a terminal open.
1) Call `sail edit someorg/project`
	1) Optionally, call `sail edit -hat someorg/project` to just modify the hat.
1) Edit the file in the editor that pops up.
1) Save
1) code-server window reloads with changed environment.

sail will listen for changes to the file being edited, and will magically
recreate the environment as changes are made (assuming those changes make
sense).

To make iterations seamless:

1) Docker caching is heavily employed.
1) `code-server` automatically reloads the page when the new environment is
created.
1) As usual, the project folder is persisted so no changes are lost.
1) UI state is persisted so the exact layout of your tabs in undisturbed.

## Hats

A _hat_ is a build directory with a Dockerfile which has it's `FROM` clause 
replaced with the repository-provided image. Essentially, hats let you
personalize your development environment.

For example:

```
FROM ubuntu-dev
RUN sudo apt install fish
RUN chsh user -s $(which fish)
```

is a hat that would install fish, and configure it as the default
shell regardless of which shell the repository-provided image uses.

The `FROM ubuntu-dev` will be replaced with `FROM <repo_image>` when sail
assembles your dev container.

---

`sail` promotes the use of Ubuntu/apt-based dev containers so that hats are 
reliable.

You can only wear a single hat at a time.

## Shares

Projects and hats can specify shares using Docker labels of form 
`share.<share_name>="host_path:guest_path"`.

For example, 

```
LABEL share.go_mod="~/go/pkg/mod:/root/go/pkg/mod"
```

Shares the host's Go mod cache with the guest.


### GitHub

To enable expirementation, hats can be used from github like so:

`-hat github:ammario/dotfiles`

---

Hats are supposed to enable global personalization, so GitHub hats should just be used for expirementation.

# Configuration

A self-documenting configuration is stored  at `~/.config/sail/sail.toml`

Checkout `defaultconfig.go` for the default config.

# Future

These features are planned for future releases:

- Windows support
- Synchronizing code-server extensions, settings, and themes.
- Remote Host support.
	- Cloud integration so `my-compute-instance` can be auto-resolved into an AWS/GCP/Azure instance.

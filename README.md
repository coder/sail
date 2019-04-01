# narwhal

`narwhal` is an expiremental CLI to efficiently Docker-based, remote development environments via
[`code-server`](https://github.com/codercom/code-server).

# Features

- Automatically connects to remote servers via SSH.
- Securely forwards remote editor over SSH.
- ssh-agent forwarding to the remote server for efficient `git` cloning.
- Projects can specify their own development environment via `.narwhal/Dockerfile`.
- Supports Linux and MacOS.

# Basic usage

Spin up a secure editor for `codercom/narwhal` on the host `135.23.23.1`.

```bash
narwhal 135.23.23.1 codercom/narwhal
# (Opens editor in browser)
```

Or:

```bash
narwhal codercom/narwhal
# (Uses local host)

```

which will use the local narwhal client.

# Environment re-use

If `narwhal` detects the same host being used with the same project, it will just open up the existing
environment.

# Configuration

A self-documenting configuration is stored  at `~/.config/narwhal/narwhal.toml`

Checkout `defaultconfig.go` for the default config.

# Local usage

The special host `localhost` will tell `narwhal` to host environments locally.

# Aliasing

Because `narwhal` uses the default `ssh` utility, one can use [SSH aliases](https://collectiveidea.com/blog/archives/2011/02/04/how-to-ssh-aliases)
for more manageable host names.

# Future

These features are planned for future releases:

- Windows support
- Synchronizing code-server extensions, settings, and themes.
- Cloud integration so `my-compute-instance` can be auto-resolved into an AWS/GCP/Azure instance.

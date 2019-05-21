+++
type="docs"
title="run"
browser_title="Sail - Commands - run"
section_order=0
+++

```
Usage: sail run [flags] <repo>

Runs a project container.
If a project is not yet created or running with the name,
one will be created and a new editor will be opened.
If a project is already up and running, this won't
start a new container, but instead will reuse the
already running container and open a new editor.

If a schema and host are not provided, sail will use github over SSH.
There are multiple ways to modify this behavior.

1. Specify a host. See examples section
2. Specify a schema and host. See examples section
3. Edit the config to provide your preferred defaults.

Examples:
	Use default host and schema (github.com over SSH, editable in config)
	- sail run cdr/code-server

	Force SSH on a Github repo (user git is assumed by default)
	- sail run ssh://github.com/cdr/sshcode
	- sail run --ssh github.com/cdr/sshcode

	Specify a custom SSH user
	- sail run ssh://colin@git.colin.com/super/secret-repo
	- sail run --ssh colin@git.colin.com/super/secret-repo

	Force HTTPS on a Gitlab repo
	- sail run https://gitlab.com/inkscape/inkscape
	- sail run --https gitlab.com/inkscape/inkscape
	
Note:
If you use ssh://, http://, or https://, you must specify a host.

This won't work:
	- sail run ssh://cdr/code-server

Instead, use flags to avoid providing a host.

This will work:
	- sail run --ssh cdr/code-server

sail run flags:
	--hat	Custom hat to use.
	--http	Clone repo over HTTP	(false)
	--https	Clone repo over HTTPS	(false)
	--image	Custom docker image to use.
	--keep	Keep container when it fails to build.	(false)
	--no-open	Don't open an editor session	(false)
	--rm	Delete existing container	(false)
	--ssh	Clone repo over SSH	(false)
	--test-cmd	A command to use in-place of starting code-server for testing purposes.
```

The `run` command starts up a container, and opens a browser window pointing to
the project's running code-server.

## Browser

Chrome is always used if it is available, because sail can open it in `--app` mode,
which makes the code-server interface feel exactly like native VS Code.

If Chrome isn't available, sail opens the URL in the OS's default browser.

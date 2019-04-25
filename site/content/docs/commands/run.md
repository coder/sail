+++
type="docs"
title="run"
browser_title="Sail - Commands - run"
section_order=0
+++

```
NAME:
	sail run - Runs a project container.

USAGE:
	sail run [flags] <project>

DESCRIPTION:
	This command is used for opening and running a project.
	If a project is not yet created or running with the name,
	one will be created and a new editor will be opened.
	If a project is already up and running, this won't
	start a new container, but instead will reuse the
	already running container and open a new editor.

Flags:
	-hat	Custom hat to use.
	-image	Custom docker image to use.
	-keep	Keep container when it fails to build.	(false)
	-test-cmd	A command to use in-place of starting code-server for testing purposes.
```

The `run` command starts up a container, and opens a browser window pointing to
the project's running code-server.

## Browser

Chrome is always used if it is available, because sail can open it in `--app` mode,
which makes the code-server interface feel exactly like native VS Code.

If Chrome isn't available, sail opens the URL in the OS's default browser.

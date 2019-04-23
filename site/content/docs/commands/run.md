+++
type="doc"
title="run"
browser_title="Sail - Commands - run"
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

The `run` command starts up a container, and plops you into code-server.

## Browser

Chrome is always used if it is available, because we can open it in `--app` mode,
which makes the code-server interface feel exactly live native VS Code.

If Chrome isn't available, we open the URL in the OS's default browser.
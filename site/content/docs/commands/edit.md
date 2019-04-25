+++
type="docs"
title="edit"
browser_title="Sail - Commands - edit"
+++

```
NAME:
    sail edit - edit your environment in real-time.

USAGE:
    sail edit [flags] <repo>

DESCRIPTION:
    This command allows you to edit your project's environment while it's running.
    Depending on what flags are set, the Dockerfile you want to change will be opened in your default
    editor which can be set using the "EDITOR" environment variable. Once your changes are complete
    and the editor is closed, the environment will be rebuilt and rerun with minimal downtime.

    If no flags are set, this will open your project's Dockerfile. If the -hat flag is set, this
    will open the hat Dockerfile associated with your running project in the editor. If the -new-hat
    flag is set, the project will be adjusted to use the new hat.

Flags:
    -hat	Edit the hat associated with this project.	(false)
    -new-hat	Path to new hat.
```

The `edit` command lets you edit your environment.

## Workflow

The workflow for modifying an environment goes like:

1. Have Sail open in some window.
1. Have a host terminal open.
1. Call `sail edit someorg/project`
  1. Optionally, call `sail edit -hat someorg/project` to just modify the hat.
1. Edit the file in the editor that pops up.
1. Save
1. code-server window reloads with changed environment.

Iteration is seamless because

1. Docker caches most of the commands.
1. `code-server` automatically reconnects the page when the new environment is
created.
1. The project folder is always persisted so WIP changes are safe.
1. UI state is persisted so the exact layout of your editor in undisturbed.
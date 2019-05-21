+++
type="docs"
title="edit"
browser_title="Sail - Commands - edit"
section_order=1
+++

```
Usage: sail edit [flags] <repo>

This command allows you to edit your project's environment while it's running.
Depending on what flags are set, the Dockerfile you want to change will be opened in your default
editor which can be set using the "EDITOR" environment variable. Once your changes are complete
and the editor is closed, the environment will be rebuilt and rerun with minimal downtime.

If no flags are set, this will open your project's Dockerfile. If the -hat flag is set, this
will open the hat Dockerfile associated with your running project in the editor. If the -new-hat
flag is set, the project will be adjusted to use the new hat.

VS Code users can edit their environment by editing their .sail/Dockerfile within the editor. VS Code
will rebuild the container when they click on the 'rebuild' button.

sail edit flags:
	--hat	Edit the hat associated with this project.	(false)
	--new-hat	Path to new hat.
```

The `edit` command lets you edit your environment.

**VS Code users should use [integrated editing](/docs/concepts/environment-editing/) instead.**

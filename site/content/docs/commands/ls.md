+++
type="docs"
title="ls"
browser_title="Sail - Commands - ls"
section_order=2
+++

```
Usage: sail ls

Lists all containers with the com.coder.sail label.

sail ls flags:
	--all	Show stopped container.	(false)
```

The `ls` command lists all containers with Sail Docker labels.

Example output:

```
name                 hat   url                     status
cdr/sail                   http://127.0.0.1:8828   Up About an hour
cdr/sshcode                http://127.0.0.1:8130   Up About an hour
cdr/m                      http://127.0.0.1:8754   Up About an hour
cdr/code-server            http://127.0.0.1:8828   Up About an hour
cdr/sail-tmp-kEG58         http://127.0.0.1:8130   Up About an hour
```

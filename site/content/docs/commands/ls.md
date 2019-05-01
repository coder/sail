+++
type="docs"
title="ls"
browser_title="Sail - Commands - ls"
section_order=2
+++

```
NAME:
    sail ls - Lists all sail containers.

USAGE:
    sail ls

DESCRIPTION:
    Queries docker for all containers with the com.coder.sail label.

Flags:
    -all	Show stopped container.	(false)
```

The `ls` command lists all containers with Sail Docker labels.

Example output:

```
name                 hat   url                     status
cdr/sail                   http://127.0.0.1:8828
cdr/sshcode                http://127.0.0.1:8130
cdr/m                      http://127.0.0.1:8754
cdr/code-server            http://127.0.0.1:8828
cdr/sail-tmp-kEG58         http://127.0.0.1:8130
```

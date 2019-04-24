+++
type="docs"
title="ls"
browser_title="Sail - Commands - ls"
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
name                      hat   url                     status
codercom/sail                   http://127.0.0.1:8828
codercom/sshcode                http://127.0.0.1:8130
codercom/m                      http://127.0.0.1:8754
codercom/code-server            http://127.0.0.1:8828
codercom/sail-tmp-kEG58         http://127.0.0.1:8130
```

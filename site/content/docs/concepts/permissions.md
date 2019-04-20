+++
type="doc"
title="Permissions"
browser_title="Sail - Docs - Permissions"
+++

The current user on the host is mapped to the user named `user` within
the development environment.  That means files permissioned to `user` within
the container, will appear to have the same set of permissions for your user on the host.


We use `user` and not `root` within the container because 

- A lot of tools will complain about being root.
- Most developers are used to being non-root and the `sudo` workflow.




+++
type="docs"
title="Running GUI Applications"
browser_title="Sail - Docs - Running GUI Applications"
section_order=1
+++


Sail supports running GUI applications when running on a Linux host with x11 support.

If the Linux machine running Sail has the `$DISPLAY` environment variable set, then the
x11 socket and xauthority file will be mounted in, and the correct environment variables
will be set in the container. This allows the user running inside of the container to run
any GUI applications they want from within their Sail environment.

For example, to start firefox from within a Sail environment:
```bash
# Ensure firefox is installed, for quick prototyping just install firefox from the terminal, but
# if this becomes a project dependency, it should be installed from the project's `.sail/Dockerfile`.
$ sudo apt-get install -y firefox

# This should open up firefox in a new window.
$ /usr/bin/firefox
```

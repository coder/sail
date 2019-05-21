+++
type="docs"
title="Add Sail to Your Project"
browser_title="Sail - Docs - Add Sail to Your Project"
section_order=0
+++

Adding sail support to your project can be done by adding a single Dockerfile. The
Dockerfile must be located at `.sail/Dockerfile` in the root of your project.

Once this file is created, you can modify it to change the `FROM` clause to be any
Sail supported image. Supported images are any of the images hosted in the [codercom 
docker hub](https://hub.docker.com/u/codercom) with the naming convention of `codercom/ubuntu-dev*`.

## Choosing a Base Image

Currently, the images are based off of ubuntu 18.10, and some of them contain preinstalled
and configured language environments that you can base your project's environment off of.

For instance, if you have a python project, you would want to change the `FROM` clause
to be `FROM codercom/ubuntu-dev-python3.7` or `FROM codercom/ubuntu-dev-python2.7` depending
on your project's version of python. This will ensure that the correct python and pip versions
are installed and configured, and that any common python vscode plugins are installed.

## Customizing Your Project's Environment

Once the base environment is chosen, any additional project dependencies and configuration can
be added using normal [Dockerfile syntax](https://docs.docker.com/engine/reference/builder/) and [Sail labels](/docs/concepts/labels).

For example:

```Dockerfile
# Use a predefined language base.
FROM codercom/ubuntu-dev-python3.7:latest

# Install some developer tooling to help out with system 
# and program monitoring.
RUN sudo apt-get update -y && sudo apt-get install -y \
    dstat \
    wireshark

# Install setuptools to use with your python project.
RUN pip install -U setuptools

# Add any environment vars you could want.
ENV PATH $PATH:/my/additional/bins

# Add a shared dir for project data.
LABEL share.app_cache "~/app/cache:~/app/cache"
```

Sail will build your project's environment from this Dockerfile, allowing you to explicitly state
your project's dependencies and configuration so that all developers are working in the same environment.

+++
type="docs"
title="Accessing Docker Within Sail"
browser_title="Sail - Docs - Accessing Docker Within Sail"
section_order=2
+++


Accessing docker from within Sail can be done by installing the docker toolchain
and bind mounting the host's docker socket into the Sail environment using a Sail
share.

Sail provides an extra environment variable `$OUTER_ROOT` which is the path to 
the current Sail environment's project root on the host.

In order to setup a project with docker support, your project's `.sail/Dockerfile`
should look similar to this:

```Dockerfile
FROM codercom/ubuntu-dev

# Share the host's docker socket with the Sail project so that you can
# access it using the docker client.
LABEL share.docker_sock "/var/run/docker.sock:/var/run/docker.sock"

# Follow the instructions for installing docker on ubuntu here:
# https://docs.docker.com/install/linux/docker-ce/ubuntu/
RUN sudo apt-get update && sudo apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg-agent \
    software-properties-common

RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -

RUN sudo apt-key fingerprint 0EBFCD88

RUN sudo add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) \
   stable"

# Only install the client since we're using the docker daemon system running on the host.
RUN sudo apt-get install -y docker-ce-cli
```

This will allow the Sail environment to access the docker daemon system on the host.

Note: This allows the environment to see all containers on the host, even itself,
so it's important not to destroy the Sail container running your project.

#!/bin/bash

set -euxo pipefail

# Add a runner user so we can run docker as 1000:1000.
sudo adduser -u 1000 -g 1000 -M --disabled-password --gecos "" runner

# Add the user to the docker group.
sudo usermod -aG docker runner

sudo chown runner:runner -R .

sudo su - runner
#!/bin/bash

set -euxo pipefail

# Add a runner user so we can run docker as 1000:1000.
sudo useradd -u 1000 -g 1000 -G docker -M  runner

sudo chown runner:runner -R .

sudo su - runner
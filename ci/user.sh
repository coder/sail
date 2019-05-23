#!/bin/bash

set -euxo pipefail

# Add a runner user so we can run docker as 1000:1000.
sudo groupadd -g 1000 runner
sudo useradd -u 1000 -g runner -M -G docker -s /bin/bash runner

sudo chown runner:runner -R .

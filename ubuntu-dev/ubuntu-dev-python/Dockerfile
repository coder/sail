FROM codercom/ubuntu-dev

# This technically has no effect until #35 is resolved.
RUN code-server --install-extension ms-python.python

# Install pip.
RUN sudo apt-get update && sudo apt-get install -y \
    python-pip \
    python3-pip

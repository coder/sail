FROM codercom/ubuntu-dev-go:latest
SHELL ["/bin/bash", "-c"]
RUN sudo apt-get update && \
  sudo apt-get install -y htop
RUN curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.34.0/install.sh | bash && \
  . ~/.nvm/nvm.sh \
  && nvm install node

LABEL project_root "~/go/src/go.coder.com"

# Modules break much of Go's tooling.
ENV GO111MODULE=off

# Install the latest version of Hugo.
RUN wget -O /tmp/hugo.deb https://github.com/gohugoio/hugo/releases/download/v0.55.4/hugo_extended_0.55.4_Linux-64bit.deb && \
  sudo dpkg -i /tmp/hugo.deb && \
  rm -f /tmp/hugo.deb

RUN installext peterjausovec.vscode-docker

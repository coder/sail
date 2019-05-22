FROM codercom/ubuntu-dev-go:latest

RUN sudo apt-get update
RUN sudo apt-get install -y htop
LABEL project_root "~/go/src/go.coder.com"

# Modules break much of Go's tooling.
ENV GO111MODULE=off

# Install the latest version of Hugo.
RUN wget -O /tmp/hugo.deb https://github.com/gohugoio/hugo/releases/download/v0.55.4/hugo_extended_0.55.4_Linux-64bit.deb && \
  sudo dpkg -i /tmp/hugo.deb && \
  rm -f /tmp/hugo.deb

RUN installext peterjausovec.vscode-docker

RUN sudo apt-get install curl

RUN sudo apt-get install software-properties-common && \
  curl -sL https://deb.nodesource.com/setup_10.x | sudo -E bash - && \
  sudo apt-get install nodejs
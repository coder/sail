FROM ubuntu:18.10

RUN apt-get update && \
    apt-get install -y vim neovim git curl wget lsof inetutils-ping gcc sudo htop locales

RUN locale-gen en_US.UTF-8
ENV LC_ALL=en_US.UTF-8

RUN apt-get install -y ripgrep

# netstat
RUN apt-get install -y net-tools

LABEL share.ssh="~/.ssh:~/.ssh"
LABEL share.gitconfig="~/.gitconfig:~/.gitconfig"

# Download in code-server into path. sail will typically override the binary
# anyways, but it's nice to have this during the build pipepline so we can
# install extensions.
RUN wget -O /usr/bin/code-server https://codesrv-ci.cdr.sh/latest-linux && \
    chmod +x /usr/bin/code-server

RUN adduser --gecos '' --disabled-password user && \
    echo "user ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers.d/nopasswd
USER user
ENV HOME /home/user

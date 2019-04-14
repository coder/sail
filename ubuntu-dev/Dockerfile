FROM ubuntu:18.10

RUN apt-get update && \
    apt-get install -y vim neovim git curl wget lsof inetutils-ping gcc sudo htop

RUN apt-get install -y ripgrep

# netstat
RUN apt-get install -y net-tools

LABEL share.ssh="~/.ssh:~/.ssh"
LABEL share.gitconfig="~/.gitconfig:~/.gitconfig"

RUN adduser --gecos '' --disabled-password user && \
    echo "user ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers.d/nopasswd
USER user

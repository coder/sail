FROM buildpack-deps:20.04

RUN apt-get update && apt-get install -y \
    vim \
    neovim \
    git \
    curl \
    wget \
    lsof \
    inetutils-ping \
    sudo \
    htop \
    man \
    ripgrep \
    net-tools \
    locales

RUN locale-gen en_US.UTF-8
# We unfortunately cannot use update-locale because docker will not use the env variables
# configured in /etc/default/locale so we need to set it manually.
ENV LC_ALL=en_US.UTF-8

# Download in code-server into path. sail will typically override the binary
# anyways, but it's nice to have this during the build pipepline so we can
# install extensions.

RUN wget -O code-server.tgz "https://codesrv-ci.cdr.sh/releases/3.0.1/linux-x86_64.tar.gz" && \
    tar -C /usr/lib -xzf code-server.tgz && \
    rm code-server.tgz && \
    ln -s /usr/lib/code-server-3.0.1-linux-x86_64/code-server /usr/bin/code-server && \
    chmod +x /usr/lib/code-server-3.0.1-linux-x86_64/code-server && \
    chmod +x /usr/bin/code-server

ADD installext /usr/bin/installext
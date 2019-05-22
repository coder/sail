FROM buildpack-deps:cosmic

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
RUN wget -O /usr/bin/code-server https://codesrv-ci.cdr.sh/latest-linux && \
    chmod +x /usr/bin/code-server


ADD installext /usr/bin/installext
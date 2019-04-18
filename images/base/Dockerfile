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

# Use fixuid so that we can dynamically change the uid:gid to match what is being used by the host user.
# Note: If the project is large, the `chown -R` call from fixuid can take a long time.
# TODO: optimize the chown for larger projects.
RUN addgroup --gid 1000 user && \
    adduser --uid 1000 --ingroup user --home /home/user --shell /bin/sh --disabled-password --gecos "" user && \
    echo "user ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers.d/nopasswd

RUN USER=user && \
    GROUP=user && \
    curl -SsL https://github.com/boxboat/fixuid/releases/download/v0.4/fixuid-0.4-linux-amd64.tar.gz | tar -C /usr/local/bin -xzf - && \
    chown root:root /usr/local/bin/fixuid && \
    chmod 4755 /usr/local/bin/fixuid && \
    mkdir -p /etc/fixuid && \
    printf "user: $USER\ngroup: $GROUP\n" > /etc/fixuid/config.yml

USER user:user

ENTRYPOINT ["fixuid"]

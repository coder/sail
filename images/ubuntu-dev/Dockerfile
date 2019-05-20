FROM %BASE

LABEL share.ssh="~/.ssh:~/.ssh"

RUN adduser --gecos '' --disabled-password user && \
    echo "user ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers.d/nopasswd
USER user

RUN mkdir -p ~/.vscode/extensions

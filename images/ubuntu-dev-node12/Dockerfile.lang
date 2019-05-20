FROM %BASE

# Configure node to install global packages to the user's home directory.
ENV NPM_PACKAGES /home/user/.npm-packages
RUN mkdir $NPM_PACKAGES
RUN sh -c 'echo "prefix=${HOME}/.npm-packages" >> $HOME/.npmrc'
ENV PATH $NPM_PACKAGES/bin:$PATH

# Add n version manager.
RUN npm install -g n

# Install typescript.
RUN npm install -g typescript

# Share the host's yarn cache.
LABEL share.go_mod "~/.cache/yarn/v4:~/.cache/yarn/v4"

# This technically has no effect until #35 is resolved.
RUN installext dbaeumer.vscode-eslint
RUN installext ms-vscode.vscode-typescript-tslint-plugin


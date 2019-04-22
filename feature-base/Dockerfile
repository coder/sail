FROM codercom/code-server

RUN curl -sL https://deb.nodesource.com/setup_11.x | sudo -E bash -

RUN sudo apt-get update && sudo apt-get install -y \
    htop \
    neofetch \
    apt-utils \
    python3 \
    python3-pip \
    nodejs \
    gcc \
    g++ \
    make \
    cmake

# install pylint for python extension
RUN python3 -m pip install -U pylint --user

# install typescript and yarn
RUN sudo npm i -g typescript yarn

# install golang
RUN wget -qO- https://dl.google.com/go/go1.12.4.linux-amd64.tar.gz | sudo tar -xzf - -C /usr/local
ENV PATH="$PATH:/usr/local/go/bin:/home/coder/go/bin"

# add go extension tooling
RUN go get -u -v github.com/ramya-rao-a/go-outline && \
    go get -u -v github.com/acroca/go-symbols && \
    go get -u -v github.com/mdempsky/gocode && \
    go get -u -v github.com/rogpeppe/godef && \
    go get -u -v golang.org/x/tools/cmd/godoc && \
    go get -u -v github.com/zmb3/gogetdoc && \
    go get -u -v golang.org/x/lint/golint && \
    go get -u -v github.com/fatih/gomodifytags && \
    go get -u -v golang.org/x/tools/cmd/gorename && \
    go get -u -v sourcegraph.com/sqs/goreturns && \
    go get -u -v golang.org/x/tools/cmd/goimports && \
    go get -u -v golang.org/x/tools/cmd/gopls && \
    go get -u -v github.com/cweill/gotests/... && \
    go get -u -v golang.org/x/tools/cmd/guru && \
    go get -u -v github.com/josharian/impl && \
    go get -u -v github.com/haya14busa/goplay/cmd/goplay && \
    go get -u -v github.com/uudashr/gopkgs/cmd/gopkgs && \
    go get -u -v github.com/davidrjenni/reftools/cmd/fillstruct && \
    go get -u -v github.com/alecthomas/gometalinter && \
    gometalinter --install

# install rust
RUN curl -sSf https://sh.rustup.rs | bash -s -- -y
ENV PATH="$PATH:/home/coder/.cargo/bin"

# install rust tooling
RUN rustup update && rustup component add rls rust-analysis rust-src

WORKDIR /home/coder

# fix perms
RUN mkdir -p ~/.config && sudo chown -R coder:coder ~/.config && \
    mkdir -p ~/.npm && sudo chown -R coder:coder ~/.npm

# clone testing repos
RUN git clone https://github.com/BurntSushi/ripgrep && \
    git clone https://github.com/CanonicalLtd/netplan && \
    git clone https://github.com/codercom/bigdur && \
    git clone https://github.com/aksakalli/gtop

# add extensions
RUN code-server --install-extension eserozvataf.one-dark-pro-monokai-darker && \
    code-server --install-extension rust-lang.rust && \
    code-server --install-extension ms-python.python && \
    code-server --install-extension ms-vscode.go && \
    code-server --install-extension vscodevim.vim 

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

# github.com/mdempsky/gocode
# github.com/uudashr/gopkgs/cmd/gopkgs
# github.com/ramya-rao-a/go-outline
# github.com/acroca/go-symbols
# golang.org/x/tools/cmd/guru
# golang.org/x/tools/cmd/gorename
# github.com/go-delve/delve/cmd/dlv
# github.com/stamblerre/gocode
# github.com/rogpeppe/godef
# github.com/sqs/goreturns
# golang.org/x/lint/golint
# golang.org/x/tools/cmd/gopls
# github.com/cweill/gotests/...
# github.com/fatih/gomodifytags
# github.com/josharian/impl
# github.com/davidrjenni/reftools/cmd/fillstruct
# github.com/haya14busa/goplay/cmd/goplay
# github.com/godoctor/godoctor
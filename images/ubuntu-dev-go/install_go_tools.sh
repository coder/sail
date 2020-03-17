#!/bin/bash

# Taken from https://github.com/Microsoft/vscode-go/wiki/Go-tools-that-the-Go-extension-depends-on
go get -u -v github.com/ramya-rao-a/go-outline
go get -u -v github.com/acroca/go-symbols
go get -u -v github.com/mdempsky/gocode
go get -u -v github.com/rogpeppe/godef
go get -u -v golang.org/x/tools/cmd/godoc
go get -u -v github.com/zmb3/gogetdoc
go get -u -v golang.org/x/lint/golint
go get -u -v github.com/fatih/gomodifytags
go get -u -v golang.org/x/tools/cmd/gorename
go get -u -v sourcegraph.com/sqs/goreturns
go get -u -v golang.org/x/tools/cmd/goimports
go get -u -v github.com/cweill/gotests/...
go get -u -v golang.org/x/tools/cmd/guru
go get -u -v github.com/josharian/impl
go get -u -v github.com/haya14busa/goplay/cmd/goplay
go get -u -v github.com/uudashr/gopkgs/cmd/gopkgs
go get -u -v github.com/davidrjenni/reftools/cmd/fillstruct
go get -u -v github.com/alecthomas/gometalinter

go get -u -v github.com/go-delve/delve/cmd/dlv

# gocode-gomod needs to be built manually as the binary is renamed.
go get -u -v -d github.com/stamblerre/gocode
go build -o $GOPATH/bin/gocode-gomod github.com/stamblerre/gocode

# Install linters for gometalinter.
$GOPATH/bin/gometalinter --install

# gopls is generally recommended over community tools.
# It's much faster and more reliable than the other options.
# FIX: https://github.com/golang/go/issues/36442 by running as described here https://github.com/golang/tools/blob/master/gopls/doc/user.md#installation
GO111MODULE=on go get golang.org/x/tools/gopls@latest


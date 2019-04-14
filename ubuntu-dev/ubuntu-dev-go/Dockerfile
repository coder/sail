FROM codercom/ubuntu-dev

ENV GO_VERSION "go1.12"
RUN curl https://storage.googleapis.com/golang/$GO_VERSION.linux-amd64.tar.gz | sudo tar -C /usr/local -xzf -
RUN sudo ln -s /usr/local/go/bin/go /usr/local/bin/go

ADD install_go_tools.sh /tmp/
RUN bash /tmp/install_go_tools.sh

LABEL share.go_mod "~/go/pkg/mod:~/go/pkg/mod"
LABEL project_root "~/go/src/"

ENV PATH "$PATH:~/go/bin"

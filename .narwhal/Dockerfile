FROM codercom/ubuntu-dev

ENV VERSION "go1.12"
RUN curl https://storage.googleapis.com/golang/$VERSION.linux-amd64.tar.gz | sudo tar -C /usr/local -xzf -
RUN sudo ln -s /usr/local/go/bin/go /usr/local/bin/go

ADD .narwhal/install_go_tools.sh /tmp/
RUN bash /tmp/install_go_tools.sh
# Put everything in $PATH.
RUN sudo cp ~/go/bin/* /usr/bin

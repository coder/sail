FROM %BASE

ADD install_go_tools.sh /tmp/
RUN bash /tmp/install_go_tools.sh

# This technically has no effect until #35 is resolved.
RUN installext ms-vscode.go

LABEL share.go_mod "~/go/pkg/mod:~/go/pkg/mod"
LABEL project_root "~/go/src/"


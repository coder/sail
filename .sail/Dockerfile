FROM codercom/ubuntu-dev-go
RUN sudo apt-get install -y htop hugo
LABEL project_root "~/go/src/go.coder.com"

# Modules break much of Go's tooling.
ENV GO111MODULE=off

FROM codercom/ubuntu-dev-go:latest
SHELL ["/bin/bash", "-c"]

# Downgrade to ubuntu 18.04 LTS (codercom/ubuntu-dev-go:latest appears to be using a non-LTS version that is end of life support)
RUN sudo sed -i 's/cosmic/bionic/g' /etc/apt/sources.list
RUN sudo echo $'Package: * \n\
Pin: release a=bionic \n\
Pin-Priority: 1001\n' | sudo tee -a /etc/apt/preferences > /dev/null
RUN sudo apt-get update -y --allow-downgrades && \
    sudo apt-get upgrade -y --allow-downgrades && \
    sudo apt dist-upgrade -y --allow-downgrades

RUN sudo apt-get install -y htop

RUN curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.34.0/install.sh | bash && \
  . ~/.nvm/nvm.sh \
  && nvm install node

LABEL project_root "~/go/src/go.coder.com"

# Install the latest version of Hugo.
RUN wget -O /tmp/hugo.deb https://github.com/gohugoio/hugo/releases/download/v0.55.4/hugo_extended_0.55.4_Linux-64bit.deb && \
  sudo dpkg -i /tmp/hugo.deb && \
  rm -f /tmp/hugo.deb

RUN installext peterjausovec.vscode-docker

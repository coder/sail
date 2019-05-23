FROM codercom/ubuntu-dev

RUN sudo apt-get update && sudo apt-get -y install fish
RUN sudo chsh user -s $(which fish)

LABEL share.fish="~/.config/fish:~/.config/fish"

FROM codercom/ubuntu-dev-go
RUN sudo apt-get install -y htop hugo
LABEL project_root "~/go/src/go.coder.com"

# Install node
ENV HOME /home/user
RUN mkdir ~/node
RUN cd ~/node && curl https://nodejs.org/dist/v10.15.3/node-v10.15.3-linux-x64.tar.xz | tar xJ --strip-components=1 --
RUN sudo ln -s ~/node/bin/node /usr/bin/node
ENV PATH "$PATH:$HOME/node/bin"
RUN npm install -g yarn

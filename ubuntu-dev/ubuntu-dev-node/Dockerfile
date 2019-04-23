FROM codercom/ubuntu-dev

USER user
RUN mkdir ~/node
RUN cd ~/node && curl https://nodejs.org/dist/v10.15.3/node-v10.15.3-linux-x64.tar.xz | tar xJ --strip-components=1 --
RUN sudo ln -s ~/node/bin/node /usr/bin/node
ENV PATH "$PATH:$HOME/node/bin"

RUN npm install -g yarn
USER root
RUN apt-get install make g++ -y
USER user

RUN mkdir -p ~/.local/share/code-server/globalStorage
RUN mkdir -p ~/.cache/yarn
LABEL share.yarn_cache "~/.cache/yarn:~/.cache/yarn"

ENV NODE_HOME ~/node

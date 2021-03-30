FROM ubuntu:18.04 as base

RUN  echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections && \
     apt-get update && \
     apt-get install -y git build-essential kernel-package fakeroot \
                        libncurses5-dev libssl-dev ccache bison flex \
                        libelf-dev rsync
CMD ["bash"]
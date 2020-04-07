# An image to build and package the Bitmask Lite (RiseupVPN and other branded builds)
# (c) LEAP Encryption Access Project 2018-2019

# we start with 18.04 (bionic) to match the core18 used in the snap
# but we will have to install a couple of extras on top...
FROM ubuntu:18.04

MAINTAINER LEAP Encryption Access Project <info@leap.se>
LABEL Description="An image to build Bitmask Lite" Vendor="LEAP" Version="1.1"
ENV OSXSDK_SHA256="631b4144c6bf75bf7a4d480d685a9b5bda10ee8d03dbf0db829391e2ef858789" \
    PATH="$PATH:/osxcross/target/bin"

RUN apt-get update && apt-get upgrade -y && \
    apt-get install -y --no-install-recommends \
            build-essential \
            make cmake \
            git curl wget \
            libappindicator3-dev libgtk-3-dev \
            mingw-w64 upx-ucl python snapcraft \
            unzip \
            devscripts fakeroot debhelper \
            clang llvm-dev libxml2-dev uuid-dev \
            libssl-dev bash patch tar \
            xz-utils bzip2 gzip sed cpio libbz2-dev \
            software-properties-common && \
    add-apt-repository -y ppa:longsleep/golang-backports && \
    apt-get update && \
    apt-get install -y --no-install-recommends golang-1.14-go golang-golang-x-tools-dev && \
    apt-get install -y dh-golang

# version in 18.04 (bionic) is too old
RUN wget http://archive.ubuntu.com/ubuntu/pool/universe/g/golang-x-text/golang-golang-x-text-dev_0.3.2-1_all.deb && \
    dpkg -i golang-golang-x-text-dev_0.3.2-1_all.deb

# we also need a newer nsis version
RUN echo "deb http://de.archive.ubuntu.com/ubuntu disco main universe" >> /etc/apt/sources.list && \
    apt-get update && \
    apt-get install -y --no-install-recommends nsis && \
    sed -i '$ d' /etc/apt/sources.list && apt-get update

# osx cross compiling
RUN git clone https://github.com/tpoechtrager/osxcross && \
    cd osxcross/tarballs && \
    wget https://s3.dockerproject.org/darwin/v2/MacOSX10.10.sdk.tar.xz && \
    echo "${OSXSDK_SHA256} *MacOSX10.10.sdk.tar.xz" | sha256sum -c - && \
    cd .. && UNATTENDED=1 ./build.sh && \
    ln -s /osxcross/target/SDK/MacOSX10.10.sdk/usr/include/objc/NSObjCRuntime.h /osxcross/target/SDK/MacOSX10.10.sdk/usr/include/objc/NSObjcRuntime.h

# bomutils (for osx packaging)
RUN git clone https://github.com/hogliux/bomutils && \
    cd bomutils && make && sudo make install

# xar (for osx packaging)
RUN git clone https://github.com/VantaInc/xar && \
    cd xar/xar && \
    ./autogen.sh && ./configure && \
    make && sudo make install

# An image to build and package the Bitmask Lite (RiseupVPN and other branded builds)
# (c) LEAP Encryption Access Project 2018-2019
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
            devscripts fakeroot debhelper nsis \
            clang llvm-dev libxml2-dev uuid-dev \
            libssl-dev bash patch tar \
            xz-utils bzip2 gzip sed cpio libbz2-dev && \
    apt-get install -y software-properties-common && \
    add-apt-repository -y ppa:longsleep/golang-backports && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
            dh-golang golang golang-golang-x-text-dev

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

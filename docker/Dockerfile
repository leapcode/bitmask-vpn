# An image to build and package the Bitmask Lite (RiseupVPN and other branded builds)
# (c) LEAP Encryption Access Project 2018-2019

# we start with 18.04 (bionic) to match the core18 used in the snap
# but we will have to install a couple of extras on top...
FROM ubuntu:18.04 as builder

MAINTAINER LEAP Encryption Access Project <info@leap.se>
LABEL Description="An image to build Bitmask Lite" Vendor="LEAP" Version="1.2"
ENV OSXSDK_SHA256="631b4144c6bf75bf7a4d480d685a9b5bda10ee8d03dbf0db829391e2ef858789" \
    PATH="$PATH:/osxcross/target/bin:/usr/lib/go-1.14/bin"

RUN apt-get update && apt-get upgrade --yes && \
    apt-get install --yes --no-install-recommends \
            build-essential \
            make cmake \
            git curl wget \
            libappindicator3-dev libgtk-3-dev \
            webkit2gtk-4.0 \
            mingw-w64 upx-ucl python snapcraft snapd \
            unzip \
            devscripts fakeroot debhelper \
            clang llvm-dev libxml2-dev uuid-dev \
            libssl-dev bash patch tar \
            xz-utils bzip2 gzip sed cpio libbz2-dev \
            software-properties-common dh-golang \
            jq \
            squashfs-tools && \
    add-apt-repository -y ppa:longsleep/golang-backports && \
    apt-get update && \
    apt-get install -y --no-install-recommends golang-1.14-go golang-go golang-golang-x-tools-dev && \
    rm -r /var/lib/apt/lists/*

# version in 18.04 (bionic) is too old
RUN wget http://archive.ubuntu.com/ubuntu/pool/universe/g/golang-x-text/golang-golang-x-text-dev_0.3.2-1_all.deb && \
    dpkg -i golang-golang-x-text-dev_0.3.2-1_all.deb

# we also need a newer nsis version
RUN echo "deb http://de.archive.ubuntu.com/ubuntu eoan main universe" >> /etc/apt/sources.list && \
    apt-get update && \
    apt-get install -y --no-install-recommends nsis && \
    sed -i '$ d' /etc/apt/sources.list && apt-get update && \
    rm -r /var/lib/apt/lists/*

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

# Grab the core18 snap (which snapcraft uses as a base) from the stable channel
# and unpack it in the proper place, to speed up snapcraft builds in the containers.
RUN curl -L $(curl -H 'X-Ubuntu-Series: 16' 'https://api.snapcraft.io/api/v1/snaps/details/core18' | jq '.download_url' -r) --output core18.snap
RUN mkdir -p /snap/core18
RUN unsquashfs -d /snap/core18/current core18.snap

# cache go modules
RUN rm -rf /gomods && mkdir -p /gomods/packages
WORKDIR /gomods
COPY mods/go.* /gomods/
COPY mods/packages/ /gomods/packages/
RUN go mod download

COPY builder.sh /

# Generate locale.
RUN apt-get update && apt-get dist-upgrade --yes && apt-get install --yes sudo locales && locale-gen en_US.UTF-8

# Set the proper environment.
ENV LANG="en_US.UTF-8"
ENV LANGUAGE="en_US:en"
ENV LC_ALL="en_US.UTF-8"

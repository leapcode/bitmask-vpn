# An image to build and package the BitmaskVPN (RiseupVPN and other branded builds)
# (c) LEAP Encryption Access Project 2018-2021

FROM ubuntu:24.04 as builder

MAINTAINER LEAP Encryption Access Project <info@leap.se>
ARG GO_VERSION=1.22
LABEL Description="An image to build Bitmask Lite" Vendor="LEAP" Version="1.2"
ENV OSXSDK_SHA256="631b4144c6bf75bf7a4d480d685a9b5bda10ee8d03dbf0db829391e2ef858789" \
    PATH="$PATH:/osxcross/target/bin:/usr/lib/go-${GO_VERSION}/bin" \
    QMAKE=qmake6

ARG DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get upgrade --yes && \
    apt-get install --yes --no-install-recommends \
            build-essential \
            make cmake \
            git curl wget \
            libappindicator3-dev libgtk-3-dev \
            webkit2gtk-4.0 \
            mingw-w64 upx-ucl python3 snapd \
            unzip sudo locales \
            devscripts fakeroot debhelper \
            clang llvm-dev libxml2-dev uuid-dev \
            libssl-dev bash patch tar \
            xz-utils bzip2 gzip sed cpio libbz2-dev \
            software-properties-common dh-golang \
            jq \
            squashfs-tools libgl-dev \
            qml-module-qtquick-controls2 libqt6qml6 libqt6svg6-dev qt6-l10n-tools \
            qt6-tools-dev qt6-tools-dev-tools qt6-base-dev qt6-base-dev-tools \
            qt6-declarative-dev qt6-declarative-dev-tools \
            qml6-module-qt5compat-graphicaleffects libqt6core5compat6 libqt6core5compat6-dev \
            golang golang-${GO_VERSION}-go golang-golang-x-tools-dev && \
    rm -r /var/lib/apt/lists/*

RUN ln -s $(qmake6 -query "QT_INSTALL_BINS")/lrelease /usr/local/bin/lrelease

# osx cross compiling
#RUN git clone https://github.com/tpoechtrager/osxcross && \
#    cd osxcross/tarballs && \
#    wget https://s3.dockerproject.org/darwin/v2/MacOSX10.10.sdk.tar.xz && \
#    echo "${OSXSDK_SHA256} *MacOSX10.10.sdk.tar.xz" | sha256sum -c - && \
#    cd .. && UNATTENDED=1 ./build.sh && \
#    ln -s /osxcross/target/SDK/MacOSX10.10.sdk/usr/include/objc/NSObjCRuntime.h /osxcross/target/SDK/MacOSX10.10.sdk/usr/include/objc/NSObjcRuntime.h

# bomutils (for osx packaging)
#RUN git clone https://github.com/hogliux/bomutils && \
#    cd bomutils && make && sudo make install

# xar (for osx packaging)
#RUN git clone https://github.com/VantaInc/xar && \
#    cd xar/xar && \
#    ./autogen.sh && ./configure && \
#    make && sudo make install

# Grab the core18 and core20 snap (which snapcraft uses as a base) from the stable channel
# and unpack it in the proper place, to speed up snapcraft builds in the containers.
# RUN curl -L $(curl -H 'X-Ubuntu-Series: 16' 'https://api.snapcraft.io/api/v1/snaps/details/core18' | jq '.download_url' -r) --output core18.snap
# RUN mkdir -p /snap/core18
# RUN unsquashfs -d /snap/core18/current core18.snap
# RUN curl -L $(curl -H 'X-Ubuntu-Series: 16' 'https://api.snapcraft.io/api/v1/snaps/details/core20' | jq '.download_url' -r) --output core20.snap
# RUN mkdir -p /snap/core20
# RUN unsquashfs -d /snap/core20/current core20.snap

# Grab the snapcraft snap from the stable channel and unpack it in the proper
# place.
# RUN curl -L $(curl -H 'X-Ubuntu-Series: 16' 'https://api.snapcraft.io/api/v1/snaps/details/snapcraft?channel=stable' | jq '.download_url' -r) --output snapcraft.snap
# RUN mkdir -p /snap/snapcraft
# RUN unsquashfs -d /snap/snapcraft/current snapcraft.snap


# Create a snapcraft runner (TODO: move version detection to the core of
# snapcraft).
# RUN mkdir -p /snap/bin
# RUN echo "#!/bin/sh" > /snap/bin/snapcraft
# RUN snap_version="$(awk '/^version:/{print $2}' /snap/snapcraft/current/meta/snap.yaml)" && echo "export SNAP_VERSION=\"$snap_version\"" >> /snap/bin/snapcraft
# RUN echo 'exec "$SNAP/usr/bin/python3" "$SNAP/bin/snapcraft" "$@"' >> /snap/bin/snapcraft
# RUN chmod +x /snap/bin/snapcraft
# RUN ln -s /snap/bin/snapcraft /bin/

COPY builder.sh /

# Generate locale.
RUN locale-gen en_US.UTF-8

# Set the proper environment.
ENV LANG="en_US.UTF-8"
ENV LANGUAGE="en_US:en"
ENV LC_ALL="en_US.UTF-8"
ENV SNAP="/snap/snapcraft/current"
ENV SNAP_NAME="snapcraft"
ENV SNAP_ARCH="amd64"

# An image to build and package the Bitmask Lite (RiseupVPN and other branded builds)
# (c) LEAP Encryption Access Project 2018
FROM ubuntu:disco

MAINTAINER LEAP Encryption Access Project <info@leap.se>
LABEL Description="An image to build Bitmask Lite" Vendor="LEAP" Version="1.1"

RUN apt update && apt upgrade -y

# Install all the deps for building and packaging bitmask-vpn

RUN apt install -y --no-install-recommends \
        build-essential \
        make \
        git curl wget \
        golang libappindicator3-dev libgtk-3-dev \
        mingw-w64 upx-ucl python snapcraft \
        unzip \
        devscripts fakeroot debhelper dh-golang golang-golang-x-text-dev 

# we need nsis >= 3.0
RUN echo 'deb http://archive.ubuntu.com/ubuntu cosmic universe' >> /etc/apt/sources.list
RUN apt update
RUN apt install -y --no-install-recommends -t cosmic nsis


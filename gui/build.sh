#!/bin/bash
set -e
set -x

XBUILD=${XBUILD-no}
WIN64="win64"
GO=`which go`

PROJECT=bitmask.pro
TARGET_GOLIB=lib/libgoshim.a
SOURCE_GOLIB=gui/backend.go

RELEASE=qtbuild/release

if [ "$TARGET" == "" ]
then
    TARGET=riseup-vpn
fi

if [ "$XBUILD" == "$WIN64" ]
then
    # TODO allow to override vars
    QMAKE="`pwd`/../../mxe/usr/x86_64-w64-mingw32.static/qt5/bin/qmake"
    PATH="`pwd`/../../mxe/usr/bin"/:$PATH
    CC=x86_64-w64-mingw32.static-gcc
else
    if [ "$QMAKE" == "" ]
    then
        QMAKE=`which qmake`
    fi
fi

PLATFORM=`uname -s`

function init {
    mkdir -p lib
}

function buildGoLib {
    echo "[+] Using go in" $GO "[`go version`]"
    $GO generate ./pkg/config/version/genver/gen.go
    if [ "$PLATFORM" == "Darwin" ]
    then
        OSX_TARGET=10.12
        GOOS=darwin
	CC=clang
	CGO_CFLAGS="-g -O2 -mmacosx-version-min=$OSX_TARGET"
	CGO_LDFLAGS="-g -O2 -mmacosx-version-min=$OSX_TARGET"
    fi
    if [ "$XBUILD" == "no" ]
    then
        echo "[+] Building Go library with standard Go compiler"
        CGO_ENABLED=1 GOOS=$GOOS CC=$CC CGO_CFLAGS=$CGO_CFLAGS CGO_LDFLAGS=$CGO_LDFLAGS go build -buildmode=c-archive -o $TARGET_GOLIB $SOURCE_GOLIB
    fi
    if [ "$XBUILD" == "$WIN64" ]
    then
        echo "[+] Building Go library with mxe"
        echo "[+] Using cc:" $CC
        CC=$CC CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -buildmode=c-archive -o $TARGET_GOLIB $SOURCE_GOLIB
    fi
}

function buildQmake {
    echo "[+] Now building Qml app with Qt qmake"
    echo "[+] Using qmake in:" $QMAKE
    mkdir -p qtbuild
    $QMAKE -o qtbuild/Makefile "CONFIG-=debug CONFIG+=release" $PROJECT
}

echo "[+] Building BitmaskVPN"

lrelease bitmask.pro
buildGoLib
buildQmake
make -C qtbuild clean
make -C qtbuild -j4 all

# i would expect that passing QMAKE_TARGET would produce the right output, but nope.
mv qtbuild/release/bitmask $RELEASE/$TARGET
strip $RELEASE/$TARGET
echo "[+] Binary is in" $RELEASE/$TARGET

#!/bin/bash
set -e


XBUILD=${XBUILD-no}
WIN64="win64"
GO=`which go`

PROJECT=bitmask.pro
TARGET_GOLIB=lib/libgoshim.a
SOURCE_GOLIB=gui/backend.go


if [ "$XBUILD" == "$WIN64" ]
then
    # TODO allow to override vars
    QMAKE="`pwd`/../mxe/usr/x86_64-w64-mingw32.static/qt5/bin/qmake"
    PATH="`pwd`/../mxe/usr/bin"/:$PATH
    CC=x86_64-w64-mingw32.static-gcc
else
    QMAKE=`which qmake`
fi


function init {
    mkdir -p lib
}

function buildGoLib {
    echo "[+] Using go in" $GO "[`go version`]"
    $GO generate ./pkg/config/version/gen.go
    if [ "$XBUILD" == "no" ]
    then
        echo "[+] Building Go library with standard Go compiler"
        CGO_ENABLED=1 go build -buildmode=c-archive -o $TARGET_GOLIB $SOURCE_GOLIB
    fi
    if [ "$XBUILD" == "$WIN64" ]
    then
        echo "[+] Building Go library with mxe"
        echo ">> using cc:" $CC
        CC=$CC CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -buildmode=c-archive -o $TARGET_GOLIB $SOURCE_GOLIB
    fi
}

function buildQmake {
    echo "[+] Now building Qml app with Qt qmake"
    echo ">> using qmake:" $QMAKE
    mkdir -p qtbuild
    $QMAKE -o qtbuild/Makefile "CONFIG-=debug CONFIG+=release" $PROJECT
}

echo "[+] Building BitmaskVPN"

buildGoLib
buildQmake
make -C qtbuild clean
make -C qtbuild -j4 all

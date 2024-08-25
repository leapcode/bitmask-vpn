#!/bin/bash
set -e

# DEBUG --------------
# set -x
# --------------------

XBUILD="${XBUILD:-no}"
LRELEASE="${LRELEASE:-lrelease}"
VENDOR_PATH="${VENDOR_PATH:-providers/riseup}"
APPNAME="${APPNAME:-Bitmask}"
LDFLAGS_VER="-X 0xacab.org/leap/bitmask-vpn/pkg/config/version.appVersion=${VERSION}"

OSX_TARGET=12
WIN64="win64"
GO=`which go`

PROJECT=bitmask.pro
TARGET_GOLIB=lib/libgoshim.a
SOURCE_GOLIB=gui/backend.go

MAKE=${MAKE:=make}
QTBUILD=build/qt
RELEASE_DIR=$QTBUILD/release
DEBUGP=$QTBUILD/debug

PLATFORM=$(uname -s)
LDFLAGS=""
BUILD_GOLIB="yes"

if [ "$TARGET" == "" ]
then
    TARGET=riseup-vpn
fi

# XXX for some reason, MAKEFLAGS is set to "w"
# by debhelper
if [ "$MAKEFLAGS" == "w" ]
then
    MAKEFLAGS=
fi

if [ "$CC" == "cc" ]
then
    CC="gcc"
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
        QMAKE="$(command -v qmake6 || command -v qmake)"
    fi
fi

PLATFORM=`uname -s`

function init {
    mkdir -p lib
}

function buildGoLib {
    if [ "$PLATFORM" == "Darwin" ]
    then
        GOOS=darwin
        CC=clang
        CGO_CFLAGS="-g -O2 -mmacosx-version-min=$OSX_TARGET"
        CGO_LDFLAGS="-g -O2 -mmacosx-version-min=$OSX_TARGET"
    fi

    if [ "$XBUILD" == "no" ]
    then
        echo "[+] Building Go library with standard Go compiler"
        CGO_ENABLED=1 GOOS=$GOOS CC=$CC CGO_CFLAGS=$CGO_CFLAGS CGO_LDFLAGS=$CGO_LDFLAGS go build -buildmode=c-archive \
            -ldflags="${LDFLAGS_VER} -extar=$AR -extld=$LD -extldflags=$LDFLAGS" -o $TARGET_GOLIB $SOURCE_GOLIB
    fi
    if [ "$XBUILD" == "$WIN64" ]
    then
        echo "[+] Building Go library with mxe"
        echo "[+] Using cc:" $CC
        CC=$CC CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -buildmode=c-archive -ldflags="${LDFLAGS_VER} \
            -extar=$AR -extld=$LD -extldflags=$LDFLAGS" -o $TARGET_GOLIB $SOURCE_GOLIB
    fi
}

function buildQmake {
    echo "[+] Now building Qml app with Qt qmake"
    echo "[+] Using qmake in:" $QMAKE
    mkdir -p $QTBUILD
    $QMAKE -early QMAKE_CC=$CC QMAKE_CXX=$CXX QMAKE_LINK=$CXX -o "$QTBUILD/Makefile" CONFIG+=release VENDOR_PATH="${VENDOR_PATH}" \
	    RELEASE=${RELEASE} TARGET=${TARGET} $PROJECT
    #CONFIG=+force_debug_info CONFIG+=debug CONFIG+=debug_and_release
}

function renameOutput {
    # i would expect that passing QMAKE_TARGET would produce the right output, but nope.
    if [ "$PLATFORM" == "Linux" ]
    then
        if [ "$DEBUG" == "1" ]
        then
          echo "[+] Selecting DEBUG build"
          mv $DEBUGP/bitmask $RELEASE_DIR/$TARGET
        else
          echo "[+] Selecting RELEASE build"
          mv $RELEASE_DIR/bitmask $RELEASE_DIR/$TARGET
          strip $RELEASE_DIR/$TARGET
        fi
        echo "[+] Binary is in" $RELEASE_DIR/$TARGET
    elif  [ "$PLATFORM" == "Darwin" ]
    then
        rm -rf $RELEASE_DIR/$TARGET.app
        mv $RELEASE_DIR/bitmask.app/ $RELEASE_DIR/$TARGET.app/
	mv $RELEASE_DIR/$TARGET.app/Contents/MacOS/bitmask $RELEASE_DIR/$TARGET.app/Contents/MacOS/$APPNAME
	# bsd sed
	sed -i '' "s/>bitmask/>${APPNAME}/" $RELEASE_DIR/$TARGET.app/Contents/Info.plist
        echo "[+] App is in" $RELEASE_DIR/$TARGET
    else # for MINGWIN or CYGWIN
        mv $RELEASE_DIR/bitmask.exe $RELEASE_DIR/$TARGET.exe
    fi
}

function buildDefault {
    echo "[+] Building Bitmask"
    if [ "$LRELEASE" != "no" ]
    then
        $LRELEASE bitmask.pro
    fi
    if [ "$BUILD_GOLIB" == "yes" ]
    then
        buildGoLib
    fi
    buildQmake

    $MAKE -C $QTBUILD clean
    $MAKE -C $QTBUILD $MAKEFLAGS all

    echo "[+] Done."
}

echo "[build.sh] VENDOR_PATH =" ${VENDOR_PATH}
for i in "$@"
do
case $i in
    --skip-golib)
    BUILD_GOLIB="no"
    shift # past argument=value
    ;;
    --just-golib)
    BUILD_GOLIB="just"
    shift # past argument=value
    ;;
    *)
          # unknown option
    ;;
esac
done

if [ "$BUILD_GOLIB" == "just" ]
then
    buildGoLib
else
    buildDefault
fi

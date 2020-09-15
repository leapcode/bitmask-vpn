#!/bin/sh

# Downloads qt5 source file and compiles it statically.
# See https://wohlsoft.ru/pgewiki/Building_static_Qt_5 for tips

QT5_V="5.12"
QT5_VER="5.12.9"
QT5_DIR="qt-everywhere-src-$QT5_VER"
QT5_TAR="$QT5_DIR.tar.xz"
QT5_URL="https://download.qt.io/archive/qt/$QT5_V/$QT5_VER/single/$QT5_TAR"


# TODO we could use -qt-freetype, but then we have to ship our own fonts.

CONFIG_FLAGS="-prefix $PWD/../qt5-static -release -opensource -confirm-license -platform linux-g++ \
              -fontconfig -system-freetype \
              -opengl \
              -no-ssl \
              --doubleconversion=qt \
              --zlib=qt \
              --libjpeg=no \
              --icu=no \
              --libpng=qt --pcre=qt --xcb=qt --harfbuzz=qt \
              -skip wayland -skip purchasing -skip serialbus -skip qtserialport -skip script -skip scxml -skip speech \
              -static \
              -optimize-size -nomake examples -nomake tests"

# --xcb=system

if [ -f "$QT5_TAR" ]; then
    echo "[+] $QT5_TAR already downloaded."
else
    echo "[+] Qt5 source tarball does not exist. Attempting to download..."
    wget -c $QT5_URL
fi

tar xf "$QT5_TAR"
cd $QT5_DIR

./configure ${CONFIG_FLAGS}
make -j 8 && make install

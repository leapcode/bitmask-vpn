#!/bin/bash
set -e
export DESTDIR=/bitmask-vpn.orig/deploy/
rm -rf /bitmask-vpn
cp -r /bitmask-vpn.orig /bitmask-vpn
cd /bitmask-vpn
make prepare
make build
case $XBUILD in
    win)
        make package_win
        ;;
    osx)
        make package_osx
        ;;
    yes)
        make packages
        ;;
esac
cp  /bitmask-vpn/deploy/* $DESTDIR

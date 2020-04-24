#!/bin/bash
set -e
export HOSTDIR=/bitmask-vpn.host
export GUESTDIR=/bitmask-vpn
export DESTDIR="${HOSTDIR}"/deploy/
rm -rf "${GUESTDIR}"
cp -r "${HOSTDIR}" "${GUESTDIR}"
cd "${GUESTDIR}"
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
cp  "${GUESTDIR}"/deploy/* $DESTDIR

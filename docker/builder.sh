#!/bin/bash
set -e
export HOSTDIR=/bitmask-vpn.host
export GUESTDIR=/bitmask-vpn
export DESTDIR="${HOSTDIR}"/deploy/
rm -rf "${GUESTDIR}"
cp -r "${HOSTDIR}" "${GUESTDIR}"
cd "${GUESTDIR}"
make prepare
case $TYPE in
    snap)
        echo "[+] Building SNAP"
        make package_snap
        ;;
    default)
        make build
        ;;
esac
case $XBUILD in
    win)
        if [ "$STAGE" = "1" ]; then
            echo ""
            echo "[+] Bulding WIN installer >>>>>>>>>>> STAGE 1"
            make package_win_stage_1
            echo ""
        fi
        if [ "$STAGE" = "2" ]; then
            echo ""
            echo "[+] Building WIN installer >>>>>>>>>> STAGE 2"
            make package_win_stage_2
            echo ""
        fi
        ;;
    osx)
        make package_osx
        ;;
    yes)
        make packages
        ;;
    default)
        echo "no XBUILD set..."
        ;;
esac
cp  "${GUESTDIR}"/deploy/* $DESTDIR

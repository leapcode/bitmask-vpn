#!/bin/bash

#############################################################################
# Builds OpenVPN statically against mbedtls (aka polarssl).
# Requirements:  cmake
# Output: ~/openvpn_build/sbin/openvpn-x.y.z
# License: GPLv3 or later
#############################################################################

set -e
#set -x

# [!] This needs to be updated for every release --------------------------
OPENVPN="openvpn-2.5.1"
OPENSSL="1.1.1j"
MBEDTLS="2.25.0"
LZO="lzo-2.10"
ZLIB="zlib-1.2.13"
LZO_SHA1="4924676a9bae5db58ef129dc1cebce3baa3c4b5d"
OPENSSL_SHA256="aaf2fcb575cdf6491b98ab4829abf78a3dec8402b8b81efc8f23c00d443981bf"
MBEDTLS_SHA256="f838f670f51070bc6b4ebf0c084affd9574652ded435b064969f36ce4e8b586d"
# -------------------------------------------------------------------------

platform='unknown'
unamestr=`uname`
if [[ "$unamestr" == 'Linux' ]]; then
   platform='linux'
elif [[ "$unamestr" == 'Darwin' ]]; then
   platform='osx'
fi

BUILDDIR="openvpn_build"
mkdir -p ~/$BUILDDIR && cd ~/$BUILDDIR

BASE=`pwd`
SRC=$BASE/src
mkdir -p $SRC

SHASUM="/usr/bin/shasum"

ZLIB_KEYS="https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x5ed46a6721d365587791e2aa783fcd8e58bcafba"
OPENVPN_KEYS="https://swupdate.openvpn.net/community/keys/security.key.asc"

WGET="wget --prefer-family=IPv4"
DEST=$BASE/install
LDFLAGS="-L$DEST/lib -L$DEST/usr/local/lib -W"
CPPFLAGS="-I$DEST/include"
CFLAGS="-D_FORTIFY_SOURCE=2 -O1 -Wformat -Wformat-security -fstack-protector -fPIE"
CXXFLAGS=$CFLAGS
CONFIGURE="./configure --prefix=/install"
MAKE="make -j4"


######## ####################################################################
# ZLIB # ####################################################################
######## ####################################################################

function build_zlib()
{
        gpg --fetch-keys $ZLIB_KEYS
	mkdir -p $SRC/zlib && cd $SRC/zlib

	if [ ! -f $ZLIB.tar.gz ]; then
	    $WGET https://zlib.net/$ZLIB.tar.gz
	    $WGET https://zlib.net/$ZLIB.tar.gz.asc
	fi
	tar zxvf $ZLIB.tar.gz
	cd $ZLIB

	LDFLAGS=$LDFLAGS \
	CPPFLAGS=$CPPFLAGS \
	CFLAGS=$CFLAGS \
	CXXFLAGS=$CXXFLAGS \
	./configure \
	--prefix=/install

	$MAKE
	make install DESTDIR=$BASE
}

######## ####################################################################
# LZO2 # ####################################################################
######## ####################################################################

function build_lzo2()
{
	mkdir -p $SRC/lzo2 && cd $SRC/lzo2
	if [ ! -f $LZO.tar.gz ]; then
	    $WGET http://www.oberhumer.com/opensource/lzo/download/$LZO.tar.gz
	fi
	sha1=`$SHASUM $LZO.tar.gz | cut -d' ' -f 1`
	if [ "${LZO_SHA1}" = "${sha1}" ]; then
	    echo "[+] sha1 verified ok"
	else
	    echo "[!] problem with sha1 verification"
	    exit 1
	fi
	tar zxvf $LZO.tar.gz
	cd $LZO

	LDFLAGS=$LDFLAGS \
	CPPFLAGS=$CPPFLAGS \
	CFLAGS=$CFLAGS \
	CXXFLAGS=$CXXFLAGS \
	$CONFIGURE --enable-static --disable-debug

	$MAKE
	make install DESTDIR=$BASE
}

########### ##################################################################
# OPENSSL # ##################################################################
########### ##################################################################

function build_openssl()
{
	cd $BASE
	mkdir -p $SRC/openssl && cd $SRC/openssl/
	if [ ! -f openssl-$OPENSSL.tar.gz ]; then
	    $WGET https://www.openssl.org/source/openssl-$OPENSSL.tar.gz
	fi
	sha256=`${SHASUM} -a 256 openssl-${OPENSSL}.tar.gz | cut -d' ' -f 1`
	
	if [ "${OPENSSL_SHA256}" = "${sha256}" ]; then
	    echo "[+] sha-256 verified ok"
	else
	    echo "[!] problem with sha-256 verification"
            echo "[ ] expected: " ${OPENSSL_SHA256}
            echo "[ ] got:      " ${sha256}
	    exit 1
	fi
	tar zxvf openssl-$OPENSSL.tar.gz
	cd openssl-$OPENSSL
	# Kudos to Jonathan K. Bullard from Tunnelblick.
	# TODO pass cc/arch if osx
	./Configure darwin64-x86_64-cc no-shared zlib no-asm --openssldir="$DEST"
	make build_libs build_apps openssl.pc libssl.pc libcrypto.pc
	make DESTDIR=$DEST install_sw
}

########### ##################################################################
# MBEDTLS # ##################################################################
########### ##################################################################

function build_mbedtls()
{
	mkdir -p $SRC/mbedtls && cd $SRC/mbedtls
	if [ ! -f v$MBEDTLS.tar.gz ]; then
	    $WGET https://github.com/ARMmbed/mbedtls/archive/v$MBEDTLS.tar.gz
	fi
	sha256=`${SHASUM} -a 256 v${MBEDTLS}.tar.gz | cut -d' ' -f 1`
	
	if [ "${MBEDTLS_SHA256}" = "${sha256}" ]; then
	    echo "[+] sha-256 verified ok"
	else
	    echo "[!] problem with sha-256 verification"
            echo "[ ] expected: " ${MBEDTLS_SHA256}
            echo "[ ] got:      " ${sha256}
	    exit 1
	fi
	tar zxvf v$MBEDTLS.tar.gz
	cd mbedtls-$MBEDTLS
        #scripts/config.pl full   ## available for mbedtls 2.16
        scripts/config.py full    ## available for mbedtls 2.25
	mkdir -p build
	cd build
	cmake ..
	$MAKE
	make install DESTDIR=$DEST
}

########### #################################################################
# OPENVPN # #################################################################
# OPENSSL # #################################################################
########### #################################################################

function build_openvpn_openssl()
{
	mkdir -p $SRC/openvpn && cd $SRC/openvpn
	gpg --fetch-keys $OPENVPN_KEYS
	if [ ! -f "$OPENVPN.tar.gz" ]; then
            $WGET https://build.openvpn.net/downloads/releases/$OPENVPN.tar.gz
            $WGET https://build.openvpn.net/downloads/releases/$OPENVPN.tar.gz.asc
	fi
	gpg --verify $OPENVPN.tar.gz.asc && echo "[+] gpg verification ok"
	tar zxvf $OPENVPN.tar.gz
	cd $OPENVPN

	
	CFLAGS="$CFLAGS -D __APPLE_USE_RFC_3542 -I$DEST/usr/local/include" \
	LZO_CFLAGS="-I$DEST/include" \
	LZO_LIBS="$DEST/lib/liblzo2.a" \
	OPENSSL_CFLAGS=-I$DEST/usr/local/include/ \
	OPENSSL_SSL_CFLAGS=-I$DEST/usr/local/include/ \
	OPENSSL_LIBS="$DEST/usr/local/lib/libssl.a $DEST/usr/local/lib/libcrypto.a $DEST/lib/libz.a" \
	OPENSSL_SSL_LIBS="$DEST/usr/local/lib/libssl.a" \
	OPENSSL_CRYPTO_LIBS="$DEST/usr/local/lib/libcrypto.a" \
	LDFLAGS=$LDFLAGS \
	CPPFLAGS=$CPPFLAGS \
	CXXFLAGS=$CXXFLAGS \
	$CONFIGURE \
	--disable-lz4 \
	--disable-unit-tests \
	--disable-plugin-auth-pam \
	--enable-small \
	--disable-debug
	$MAKE LIBS="-all-static"
	make install DESTDIR=$BASE/openvpn
	mkdir -p $BASE/sbin/
	cp $BASE/openvpn/install/sbin/openvpn $BASE/sbin/$OPENVPN
	strip $BASE/sbin/$OPENVPN
}


########### #################################################################
# OPENVPN # #################################################################
# MBEDTLS # #################################################################
########### #################################################################

function build_openvpn_mbedtls()
{
	mkdir -p $SRC/openvpn && cd $SRC/openvpn
	gpg --fetch-keys $OPENVPN_KEYS
	if [ ! -f $OPENVPN.tar.gz ]; then
            $WGET https://build.openvpn.net/downloads/releases/$OPENVPN.tar.gz
            $WGET https://build.openvpn.net/downloads/releases/$OPENVPN.tar.gz.asc
	fi
	gpg --verify $OPENVPN.tar.gz.asc && echo "[+] gpg verification ok"
	tar zxvf $OPENVPN.tar.gz
	cd $OPENVPN

	MBEDTLS_CFLAGS=-I$DEST/usr/local/include/ \
	MBEDTLS_LIBS="$DEST/usr/local/lib/libmbedtls.a $DEST/usr/local/lib/libmbedcrypto.a $DEST/usr/local/lib/libmbedx509.a" \
	LDFLAGS=$LDFLAGS \
	CPPFLAGS=$CPPFLAGS \
	CFLAGS="$CFLAGS -I$DEST/usr/local/include" \
	CXXFLAGS=$CXXFLAGS \
	$CONFIGURE \
	--disable-plugin-auth-pam \
	--with-crypto-library=mbedtls
	# TODO debug first
	#--enable-small \
	#--disable-debug

	$MAKE LIBS="-all-static -lz -llzo2"
	make install DESTDIR=$BASE/openvpn
	mkdir -p $BASE/sbin/
	cp $BASE/openvpn/install/sbin/openvpn $BASE/sbin/$OPENVPN
	strip $BASE/sbin/$OPENVPN
}

function build_all()
{
	echo "[+] Building" $OPENVPN
	build_zlib
	build_lzo2
	build_openssl
	build_openvpn_openssl
	#build_mbedtls  # broken, see #311
	#build_openvpn_mbedtls
}

function main()
{
    if [[ $platform == 'linux' ]]; then
      build_all
    fi
    if [[ $platform == 'osx' ]]; then
      build_all
    fi
}

main "$@"

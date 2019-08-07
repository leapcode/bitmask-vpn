#########################################
# (c) LEAP Encryption Access Project 2018
include vendor.mk
#########################################

#
# TODO review some of these targets, can go in the parent makefile
# 

SYSTRAY := 0xacab.org/leap/bitmask-vpn
STAGING := staging
SYSTRAY_BIN := bitmask-vpn
HELPER_BIN := bitmask_helper
BUILD_RELEASE?=no
WIN_CERT_PATH?=z:\leap\LEAP.pfx
WIN_CERT_PASS?=
OSX_CERT = "Developer ID Installer: LEAP Encryption Access Project"
DEB_VERSION = $(shell echo ${VERSION} | cut -d '-' -f 1,2)

ifeq ($(OS),Windows_NT)     # is Windows_NT on XP, 2000, 7, Vista, 10...
SYSTEM = Windows
else
SYSTEM = $(shell uname -s)
endif


# -----------------------------------------------------------------------------
# Windows 
# -----------------------------------------------------------------------------
ifeq ($(SYSTEM), Windows)
pkg_win: staging\nssm.exe staging\openvpn\openvpn.exe
	echo "[+] building windows"
	copy ..\bin\bitmask-vpn staging\bitmask-vpn.exe
	copy ..\bin\bitmask-helper staging\bitmask_helper.exe
	"C:\Program Files (x86)\NSIS\makensis.exe" windows/$(APPNAME)-installer.nsi
else
pkg_win: staging/nssm.exe staging/openvpn/openvpn.exe
	echo "[+] building windows"
	cp ../bin/bitmask-vpn staging/bitmask-vpn.exe
	cp ../bin/bitmask-helper staging/bitmask_helper.exe
	makensis windows/$(APPNAME)-installer.nsi
endif

sign_win:
	echo "[+] signing windows build"
	python windowns/sign.py $(WIN_CERT_PATH) $(WIN_CERT_PASS)

# -----------------------------------------------------------------------------
# OSX
# -----------------------------------------------------------------------------

pkg_osx:
	echo "[+] Building osx package..."
	cp ../bin/bitmask-vpn dist/$(APPNAME).app/Contents/MacOS/
	cp ../bin/bitmask-helper dist/$(APPNAME).app/Contents/MacOS/
	cp $(STAGING)/openvpn-osx dist/$(APPNAME).app/Contents/Resources/openvpn.leap
ifeq ($(SYSTEM), Darwin)
	osx/quickpkg --output dist/$(APPNAME)-$(VERSION)_unsigned.pkg --scripts osx/scripts/ dist/$(APPNAME).app/
	@if [ $(BUILD_RELEASE) = no ]; then\
		echo "[!] BUILD_RELEASE=no, skipping signature";\
	else\
		echo "[+] Signing the bundle";\
		productsign --sign $(OSX_CERT) dist/$(APPNAME)-$(VERSION)_unsigned.pkg dist/$(APPNAME)-$(VERSION).pkg;\
	fi
else
	mkdir dist/tmp
	mv dist/$(APPNAME).app dist/tmp
	mkbom -u 0 -g 80 dist/tmp dist/$(APPNAME)/Bom
	mv dist/tmp/$(APPNAME).app dist
	rmdir dist/tmp
	
	cd dist && find $(APPNAME).app | cpio -o --format odc --owner 0:80 | gzip -c > $(APPNAME)/Payload
	cd osx/scripts && find . | cpio -o --format odc --owner 0:80 | gzip -c > ../../dist/$(APPNAME)/Scripts
	xar --compression none -cf dist/RiseupVPN-$(VERSION)_unsigned.pkg dist/$(APPNAME)/
endif

# -----------------------------------------------------------------------------
# Linux 
# -----------------------------------------------------------------------------

pkg_snap:
	-@mkdir -p ../../deploy
	@echo "[+] building snap..."
	snapcraft build
	snapcraft snap
	@mv $(BINNAME)* ../../deploy

pkg_deb:
	echo "[+] building debian package version" ${DEB_VERSION}
	-@mkdir -p ../../deploy
	@if [ $(BUILD_RELEASE) = no ]; then\
		dch -v $(VERSION) -M "debian package generated from the git repository" && echo "[!] BUILD_RELEASE=no, incrementing changelog";\
	else\
		echo "[!] BUILD_RELEASE";\
	fi
	@mkdir -p build 
	# bitmask-vpn tar.gz should be placed in parent forlder after doing 'make prepare' with the
	# bitmask-vpn topmost Makefile.
	@cp ../bitmask-vpn_$(VERSION).tgz build/$(BINNAME)_${DEB_VERSION}.orig.tar.gz
	@cd build && tar xzf $(BINNAME)_${DEB_VERSION}.orig.tar.gz && mv bitmask-vpn_${VERSION}-src ${BINNAME}_${DEB_VERSION}
	@cp -r debian/ build/$(BINNAME)_$(DEB_VERSION)/
	@cd build/$(BINNAME)_$(DEB_VERSION) && debuild -us -uc
	@cp build/*.deb ../../deploy

# -----------------------------------------------------------------------------
# Utils
# -----------------------------------------------------------------------------

clean:
	rm -rf dist/ build/

staging\nssm.exe:
	xcopy /y "C:\ProgramData\chocolatey\lib\NSSM\tools\nssm.exe" $(STAGING)
staging/nssm.exe:
	wget https://nssm.cc/release/nssm-2.24.zip -O $(STAGING)/nssm.zip
	unzip $(STAGING)/nssm.zip -d $(STAGING)
	mv $(STAGING)/nssm-*/win32/nssm.exe $(STAGING)
	rm -rf $(STAGING)/nssm-* $(STAGING)/nssm.zip

staging\openvpn\openvpn.exe:
	if not exist staging\openvpn mkdir staging\openvpn
	wget https://build.openvpn.net/downloads/releases/latest/tap-windows-latest-stable.exe -O staging/openvpn/tap-windows.exe
# eventually, this should be built statically and cross compiled in the same pipeline that we build the installer. 
	wget https://downloads.leap.se/thirdparty/windows/openvpn-x86_64-w64-mingw32.tar.bz2 -O staging/openvpn/openvpn.tar.bz2
	7z e -y -ostaging/openvpn/ staging/openvpn/openvpn.tar.bz2
	7z e -y -r -ostaging/openvpn/ staging/openvpn/openvpn.tar *.dll
	7z e -y -r -ostaging/openvpn/ staging/openvpn/openvpn.tar *.exe
	copy .\staging\openvpn\openvpn.exe .\staging
	copy .\staging\openvpn\*.dll .\staging
staging/openvpn/openvpn.exe:
	mkdir -p staging/openvpn
	wget https://build.openvpn.net/downloads/releases/latest/tap-windows-latest-stable.exe -O $(STAGING)/openvpn/tap-windows.exe
	wget https://downloads.leap.se/thirdparty/windows/openvpn-x86_64-w64-mingw32.tar.bz2 -O $(STAGING)/openvpn/openvpn.tar.bz2
	tar xvjf $(STAGING)/openvpn/openvpn.tar.bz2 -C $(STAGING)/openvpn/
	cp $(STAGING)/openvpn/bin/openvpn.exe $(STAGING)/openvpn
	cp $(STAGING)/openvpn/bin/*.dll $(STAGING)
	cp $(STAGING)/openvpn/lib/engines-1_1/*.dll $(STAGING)


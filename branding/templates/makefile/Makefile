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
pkg_win: staging\openvpn\openvpn.exe
	echo "[+] building windows"
	copy ..\bin\bitmask-vpn staging\bitmask-vpn.exe
	copy ..\bin\bitmask-helper staging\bitmask_helper.exe
	"C:\Program Files (x86)\NSIS\makensis.exe" windows/$(APPNAME)-installer.nsi
else
pkg_win_stage_1: staging/openvpn/openvpn.exe
	echo "[+] building windows [stage1]"
	cp ../bin/windows/bitmask-vpn staging/bitmask-vpn.exe
	cp ../bin/windows/bitmask-helper-go staging/helper.exe
	touch windows/uninstall.exe
	makensis -DUNINSTALLER windows/$(APPNAME)-installer.nsi
	mv dist/produce-bitmask-uninstaller.exe ../../deploy/

pkg_win_stage_2: staging/openvpn/openvpn.exe
	echo "[+] building windows [stage2]"
	cp ../bin/windows/bitmask-vpn staging/bitmask-vpn.exe
	cp ../bin/windows/bitmask-helper-go staging/helper.exe
	cp ../windows/staging/uninstall-signed.exe windows/uninstall.exe
	makensis windows/$(APPNAME)-installer.nsi
	mv dist/$(APPNAME)-$(VERSION).exe ../../deploy/
endif

sign_win:
	echo "[+] signing windows build"
	python windows/sign.py $(WIN_CERT_PATH) $(WIN_CERT_PASS)

# -----------------------------------------------------------------------------
# OSX
# -----------------------------------------------------------------------------

pkg_osx:
	echo "[+] Building osx package..."
	cp ../bin/darwin/bitmask-vpn dist/$(APPNAME).app/Contents/MacOS/
	cp ../bin/darwin/bitmask-helper dist/$(APPNAME).app/Contents/MacOS/
	cp $(STAGING)/openvpn-osx dist/$(APPNAME).app/Contents/Resources/openvpn.leap
	chmod +x dist/$(APPNAME).app/Contents/Resources/openvpn.leap
ifeq ($(SYSTEM), Darwin)
	osx/quickpkg --output dist/$(APPNAME)-$(VERSION)_unsigned.pkg --scripts osx/scripts/ dist/$(APPNAME).app/
	@if [ $(BUILD_RELEASE) = no ]; then\
		echo "[!] BUILD_RELEASE=no, skipping signature";\
	else\
		echo "[+] Signing the bundle";\
		productsign --sign $(OSX_CERT) dist/$(APPNAME)-$(VERSION)_unsigned.pkg dist/$(APPNAME)-$(VERSION).pkg;\
	fi
else
	cd osx && APPNAME=$(APPNAME) VERSION=$(VERSION) ./cross-quickpkg
endif
	mv dist/$(APPNAME)-$(VERSION)*.pkg ../../deploy/

# -----------------------------------------------------------------------------
# Linux 
# -----------------------------------------------------------------------------

pkg_snap:
	-@mkdir -p ../../deploy
	@echo "[+] building snap..."
	cd ../..
	-@ln -s build/$(PROVIDER)/snap snap
	snapcraft build
	snapcraft snap
	-@rm snap
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
	@rm -rf build/${BINNAME}_${DEB_VERSION} build/bitmask-vpn_${VERSION}-src
	@cd build && tar xzf $(BINNAME)_${DEB_VERSION}.orig.tar.gz && mv bitmask-vpn_${VERSION}-src ${BINNAME}_${DEB_VERSION}
	@cp -r debian/ build/$(BINNAME)_$(DEB_VERSION)/
	@cd build/$(BINNAME)_$(DEB_VERSION) && debuild -us -uc
	@mv build/*.deb ../../deploy

# -----------------------------------------------------------------------------
# Utils
# -----------------------------------------------------------------------------

clean:
	rm -rf dist/ build/

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
	@mkdir -p staging/openvpn
	wget https://build.openvpn.net/downloads/releases/latest/tap-windows-latest-stable.exe -O $(STAGING)/openvpn/tap-windows.exe
	wget https://downloads.leap.se/thirdparty/windows/openvpn-x86_64-w64-mingw32.tar.bz2 -O $(STAGING)/openvpn/openvpn.tar.bz2
	@tar xjf $(STAGING)/openvpn/openvpn.tar.bz2 -C $(STAGING)/openvpn/
	cp $(STAGING)/openvpn/bin/openvpn.exe $(STAGING)/openvpn
	cp $(STAGING)/openvpn/bin/*.dll $(STAGING)
	cp $(STAGING)/openvpn/lib/engines-1_1/*.dll $(STAGING)


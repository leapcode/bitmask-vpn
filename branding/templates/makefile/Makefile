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

TGZ_PATH = $(shell pwd)/dist/$(BINNAME)-$(VERSION)
tgz:
	mkdir -p $(TGZ_PATH)
	git -C `go env GOPATH`/src/$(SYSTRAY) archive HEAD | tar -x -C $(TGZ_PATH)
	mkdir $(TGZ_PATH)/helpers
	wget -O $(TGZ_PATH)/helpers/bitmask-root https://0xacab.org/leap/bitmask-dev/raw/master/src/leap/bitmask/vpn/helpers/linux/bitmask-root
	chmod +x $(TGZ_PATH)/helpers/bitmask-root
	wget -O $(TGZ_PATH)/helpers/se.leap.bitmask.policy https://0xacab.org/leap/bitmask-dev/raw/master/src/leap/bitmask/vpn/helpers/linux/se.leap.bitmask.policy
	cd dist; tar cvzf $(BINNAME)-$(VERSION).tgz $(BINNAME)-$(VERSION)
	rm -r $(TGZ_PATH)

# -----------------------------------------------------------------------------
# Windows 
# -----------------------------------------------------------------------------
CROSS_FLAGS = CGO_ENABLED=1 GOARCH=386 GOOS=windows CC="/usr/bin/i686-w64-mingw32-gcc" CGO_LDFLAGS="-lssp" CXX="i686-w64-mingw32-c++"

openvpn_win:
	if not exist staging\openvpn mkdir staging\openvpn
	wget https://build.openvpn.net/downloads/releases/latest/tap-windows-latest-stable.exe -O staging/openvpn/tap-windows.exe
# eventually, this should be built statically and cross compiled in the same pipeline that we build the installer. 
	wget https://downloads.leap.se/thirdparty/windows/openvpn-x86_64-w64-mingw32.tar.bz2 -O staging/openvpn/openvpn.tar.bz2
	7z e -y -ostaging/openvpn/ staging/openvpn/openvpn.tar.bz2
	7z e -y -r -ostaging/openvpn/ staging/openvpn/openvpn.tar *.dll
	7z e -y -r -ostaging/openvpn/ staging/openvpn/openvpn.tar *.exe
	copy .\staging\openvpn\openvpn.exe .\staging
	copy .\staging\openvpn\*.dll .\staging
openvpn_cross_win:
	mkdir -p staging/openvpn
	wget https://build.openvpn.net/downloads/releases/latest/tap-windows-latest-stable.exe -O $(STAGING)/openvpn/tap-windows.exe
	wget https://downloads.leap.se/thirdparty/windows/openvpn-x86_64-w64-mingw32.tar.bz2 -O $(STAGING)/openvpn/openvpn.tar.bz2
	tar xvjf $(STAGING)/openvpn/openvpn.tar.bz2 -C $(STAGING)/openvpn/
	cp $(STAGING)/openvpn/bin/openvpn.exe $(STAGING)/openvpn
	cp $(STAGING)/openvpn/bin/*.dll $(STAGING)
	cp $(STAGING)/openvpn/lib/engines-1_1/*.dll $(STAGING)

# TODO no need to build systray in windows, but move this powershell hack to main build makefile
#powershell '$$gopath=go env GOPATH;$$version=git -C $$gopath/src/$(SYSTRAY) describe --tags; go build -ldflags "-H windowsgui -s -w -X main.version=$$version" -o $(STAGING)/$(SYSTRAY_BIN).exe $(SYSTRAY)/cmd/bitmask-vpn'

package_win: staging\nssm.exe helper_win systray_win
# I assume you did bootstrap openvpn_win manually already.
	echo "[+] building windows"
	"C:\Program Files (x86)\NSIS\makensis.exe" windows/$(APPNAME)-installer.nsi

sign_win:
	echo "[+] signing windows build"
	python windowns/sign.py $(WIN_CERT_PATH) $(WIN_CERT_PASS)

# TODO move to main build makefile
build_cross_win: staging/nssm.exe
	echo "!define VERSION $(VERSION)" > $(STAGING)/version.nsh
	$(CROSS_FLAGS) $(MAKE) helper_win
	mkdir -p dist
	make -C win VERSION=$(VERSION)
	makensis win/$(APPNAME)-installer.nsi

# -----------------------------------------------------------------------------
# OSX
# -----------------------------------------------------------------------------

package_osx:
	echo "[+] Building osx package..."
	osx/quickpkg --output dist/$(APPNAME)-$(VERSION)_unsigned.pkg --scripts osx/scripts/ dist/$(APPNAME).app/
	@if [ $(BUILD_RELEASE) = no ]; then\
		echo "[!] BUILD_RELEASE=no, skipping signature";\
	else\
		echo "[+] Signing the bundle";\
		productsign --sign $(OSX_CERT) dist/$(APPNAME)-$(VERSION)_unsigned.pkg dist/$(APPNAME)-$(VERSION).pkg;\
	fi

# -----------------------------------------------------------------------------
# Linux 
# -----------------------------------------------------------------------------

build_snap:
	echo "[+] building snap..."
	snapcraft build
	snapcraft snap
	mkdir -p dist
	mv $(BINNAME)* dist/

build_deb:  tgz
	echo "[+] building deb..."
	@if [ $(BUILD_RELEASE) = no ]; then\
		dch -v $(VERSION) -M "debian package generated from the git repository" && echo "[!] BUILD_RELEASE=no, incrementing changelog";\
	else\
		echo "[!] BUILD_RELEASE";\
	fi
	mkdir -p build
	cp dist/$(BINNAME)-$(VERSION).tgz build/$(BINNAME)_$(shell echo ${VERSION} | cut -d '-' -f 1-2).orig.tar.gz
	cd build && tar xzf $(BINNAME)_$(shell echo ${VERSION} | cut -d '-' -f 1-2).orig.tar.gz
	cp -r debian/ build/$(BINNAME)-$(VERSION)/
	cd build/$(BINNAME)-$(VERSION) && debuild -us -uc
	cp build/*.deb dist/
	git checkout -- debian/changelog

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

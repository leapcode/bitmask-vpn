#########################################################################
# Multiplatform build and packaging recipes for BitmaskVPN
# (c) LEAP Encryption Access Project, 2019-2020
#########################################################################

.PHONY: all get build icon locales generate_locales clean check_qtifw HAS-qtifw

XBUILD ?= no
SKIP_CACHECK ?= no
VENDOR_PATH ?= providers
APPNAME ?= $(shell VENDOR_PATH=${VENDOR_PATH} branding/scripts/getparam appname | tail -n 1)
TARGET ?= $(shell VENDOR_PATH=${VENDOR_PATH} branding/scripts/getparam binname | tail -n 1)
PROVIDER ?= $(shell grep ^'provider =' ${VENDOR_PATH}/vendor.conf | cut -d '=' -f 2 | tr -d "[:space:]")
VERSION ?= $(shell git describe)

# go paths
GOPATH = $(shell go env GOPATH)
TARGET_GOLIB=lib/libgoshim.a
SOURCE_GOLIB=gui/backend.go

# detect OS
ifeq ($(OS), Windows_NT)
PLATFORM = windows
else
UNAME = $(shell uname -s)
PLATFORM ?= $(shell echo ${UNAME} | awk "{print tolower(\$$0)}")
endif

QTBUILD = build/qt
INSTALLER = build/installer
INST_DATA = ${INSTALLER}/packages/bitmaskvpn/data/
OSX_CERT="Developer ID Installer: LEAP Encryption Access Project"
MACDEPLOYQT_OPTS = -appstore-compliant -qmldir=gui/qml -always-overwrite
# XXX expired cert -codesign="${OSX_CERT}"
	
SCRIPTS = branding/scripts
TEMPLATES = branding/templates

ifeq ($(PLATFORM), windows)
HAS_QTIFW := $(shell which binarycreator.exe)
else
HAS_QTIFW := $(shell PATH=$(PATH) which binarycreator)
endif
OPENVPN_BIN = "$(HOME)/openvpn_build/sbin/$(shell grep OPENVPN branding/thirdparty/openvpn/build_openvpn.sh | head -n 1 | cut -d = -f 2 | tr -d '"')"

#########################################################################
# go build
#########################################################################

install_go:
	# the version of go in bionic is too old. let's get something newer from a ppa.
	@sudo apt install software-properties-common
	@sudo add-apt-repository ppa:longsleep/golang-backports
	@sudo apt-get update
	@sudo apt-get install golang-go

depends:
	-@make depends$(UNAME)

dependsLinux:
	@sudo apt install golang pkg-config dh-golang golang-golang-x-text-dev cmake devscripts fakeroot debhelper curl g++ qt5-qmake qttools5-dev-tools qtdeclarative5-dev qml-module-qtquick-controls libqt5qml5 qtdeclarative5-dev qml-module-qt-labs-platform qml-module-qt-labs-qmlmodels qml-module-qtquick-extras qml-module-qtquick-dialogs
	@make -C docker deps
	@# debian needs also: snap install snapcraft --classic; snap install  multipass --beta --classic

dependsDarwin:
	# TODO - bootstrap homebrew if not there
	@brew install python3 golang make pkg-config curl
	@brew install --default-names gnu-sed

ifeq ($(PLATFORM), darwin)
EXTRA_FLAGS = MACOSX_DEPLOYMENT_TARGET=10.10 GOOS=darwin CC=clang
else
EXTRA_FLAGS =
endif

ifeq ($(PLATFORM), windows)
EXTRA_GO_LDFLAGS = "-H windowsgui"
endif


PKGFILES = $(shell find pkg -type f -name '*.go')
lib/%.a: $(PKGFILES)
	@./gui/build.sh --just-golib

golib: lib/libgoshim.a

build_gui:
	@XBUILD=no TARGET=${TARGET} VENDOR_PATH=${VENDOR_PATH}/${PROVIDER} gui/build.sh --skip-golib

build: golib build_helper build_openvpn build_gui

build_helper:
	@echo "PLATFORM: ${PLATFORM}"
	@mkdir -p build/bin/${PLATFORM}
	@go build -o build/bin/${PLATFORM}/bitmask-helper -ldflags "-X main.AppName=${APPNAME} -X main.Version=${VERSION} ${EXTRA_GO_LDFLAGS}" ./cmd/bitmask-helper/
	@echo "build helper done."

build_openvpn:
	@[ -f $(OPENVPN_BIN) ] && echo "OpenVPN already built at" $(OPENVPN_BIN) || ./branding/thirdparty/openvpn/build_openvpn.sh

build_installer: check_qtifw build
	@mkdir -p ${INST_DATA}
	@cp -r ${TEMPLATES}/qtinstaller/packages ${INSTALLER}
	@cp -r ${TEMPLATES}/qtinstaller/installer.pro ${INSTALLER}
	@cp -r ${TEMPLATES}/qtinstaller/config ${INSTALLER}
ifeq (${PLATFORM}, darwin)
	@mkdir -p ${INST_DATA}/helper
	@VERSION=${VERSION} VENDOR_PATH=${VENDOR_PATH} ${SCRIPTS}/gen-qtinstaller osx ${INSTALLER}
	@cp "${TEMPLATES}/osx/bitmask.pf.conf" ${INST_DATA}helper/bitmask.pf.conf
	@cp "${TEMPLATES}/osx/client.up.sh" ${INST_DATA}/
	@cp "${TEMPLATES}/osx/client.down.sh" ${INST_DATA}/
	@cp "${TEMPLATES}/qtinstaller/osx-data/post-install.py" ${INST_DATA}/
	@cp "${TEMPLATES}/qtinstaller/osx-data/uninstall.py" ${INST_DATA}/
	@cp "${TEMPLATES}/qtinstaller/osx-data/se.leap.bitmask-helper.plist" ${INST_DATA}/
	@cp build/bin/${PLATFORM}/bitmask-helper ${INST_DATA}/
	# FIXME our static openvpn build fails with an "Assertion failed at crypto.c". Needs to be fixed!!! - kali
	# a working (old) version:
	#@curl -L https://downloads.leap.se/thirdparty/osx/openvpn/openvpn -o build/${PROVIDER}/staging/openvpn-osx
	#FIXME FIXME @cp $(OPENVPN_BIN) ${INST_DATA}/openvpn.leap
	@rm -f ${INST_DATA}openvpn.leap && cp /usr/local/bin/openvpn ${INST_DATA}/openvpn.leap
	@echo "WARNING: workaround for broken static build. Shipping homebrew dynamically linked instead"
	@echo "[+] Running macdeployqt"
	@macdeployqt ${QTBUILD}/release/${PROVIDER}-vpn.app ${MACDEPLOYQT_OPTS}
	@cp -r "${QTBUILD}/release/${TARGET}.app"/ ${INST_DATA}/
endif
ifeq (${PLATFORM}, windows)
	@VERSION=${VERSION} ${SCRIPTS}/gen-qtinstaller windows ${INSTALLER}
	@cp build/bin/${PLATFORM}/bitmask-helper ${INST_DATA}helper.exe
	@cp branding/assets/${PROVIDER}/icon.ico ${INST_DATA}/icon.ico
	@cp ${QTBUILD}/release/${TARGET}.exe ${INST_DATA}${TARGET}.exe
	# FIXME get the signed binaries with curl from openvpn downloads page - see if we have to adapt the openvpn-build to install tap drivers etc from our installer.
	@cp "/c/Program Files/OpenVPN/bin/openvpn.exe" ${INST_DATA}
	@cp "/c/Program Files/OpenVPN/bin/"*.dll ${INST_DATA}
	# XXX add sign options
	@windeployqt --qmldir gui/qml ${INST_DATA}${TARGET}.exe
endif
ifeq (${PLATFORM}, linux)
	@VERSION=${VERSION} ${SCRIPTS}/gen-qtinstaller linux ${INSTALLER}
endif
	@echo "[+] All templates, binaries and libraries copied to build/installer."
	@echo "[+] Now building the installer."
	@cd build/installer && qmake INSTALLER=${APPNAME}-installer-${VERSION} && make

check_qtifw: 
ifdef HAS_QTIFW
	@echo "[+] Found QTIFW"
else
	$(error "[!] Cannot find QTIFW. Please install it and add it to your PATH")
endif

# ----------- FIXME ------- old build, reuse or delete -----------------------------

CROSS_WIN_FLAGS = CGO_ENABLED=1 GOARCH=386 GOOS=windows CC="/usr/bin/i686-w64-mingw32-gcc" CGO_LDFLAGS="-lssp" CXX="i686-w64-mingw32-c++"
PLATFORM_WIN = PLATFORM=windows
EXTRA_LDFLAGS_WIN = EXTRA_LDFLAGS="-H windowsgui" 
build_cross_win:
	@echo "[+] Cross-building for windows..."
	@$(CROSS_WIN_FLAGS) $(PLATFORM_WIN) $(EXTRA_LDFLAGS_WIN) $(MAKE) _buildparts
	# workaround for helper: we use the go compiler
	@echo "[+] Compiling helper with the Go compiler to work around missing stdout bug..."
	cd cmd/bitmask-helper && GOOS=windows GOARCH=386 go build -ldflags "-X main.version=`git describe --tags` -H windowsgui" -o ../../build/bin/windows/bitmask-helper-go

CROSS_OSX_FLAGS = MACOSX_DEPLOYMENT_TARGET=10.10 CGO_ENABLED=1 GOOS=darwin CC="o64-clang"
PLATFORM_OSX = PLATFORM=darwin
build_cross_osx:
	$(CROSS_OSX_FLAGS) $(PLATFORM_OSX) $(MAKE) _buildparts

# --------- FIXME -----------------------------------------------------------------------

clean:
	@rm -rf build/
	@unlink branding/assets/default || true

#########################################################################
# build them all
#########################################################################

build_all_providers:
	branding/scripts/build-all-providers

########################################################################
# tests
#########################################################################


test:
	@go test -tags "integration $(TAGS)" ./pkg/...

test_ui: golib
	@qmake -o tests/Makefile test.pro
	@make -C tests clean
	@make -C tests
	@./tests/build/test_ui


#########################################################################
# packaging templates
#########################################################################

vendor_init:
	@VENDOR_PATH=${VENDOR_PATH} ./branding/scripts/init

vendor_check:
	@VENDOR_PATH=${VENDOR_PATH} ./branding/scripts/check ${PROVIDER}
ifeq (${SKIP_CACHECK}, no)
	@VENDOR_PATH=${VENDOR_PATH} ${SCRIPTS}/check-ca-crt ${PROVIDER}
endif

vendor: gen_providers_json prepare_templates gen_pkg_snap gen_pkg_deb

gen_providers_json:
	@VENDOR_PATH=${VENDOR_PATH} branding/scripts/gen-providers-json gui/providers/providers.json

prepare_templates: generate tgz
	@mkdir -p build/${PROVIDER}/bin/ deploy
	@cp ${TEMPLATES}/makefile/Makefile build/${PROVIDER}/Makefile
	@VERSION=${VERSION} VENDOR_PATH=${VENDOR_PATH} ${SCRIPTS}/generate-vendor-make build/${PROVIDER}/vendor.mk

generate:
	@go generate gui/backend.go
	@go generate pkg/config/version/genver/gen.go

TGZ_NAME = bitmask-vpn_${VERSION}-src
TGZ_PATH = $(shell pwd)/build/${TGZ_NAME}
tgz:
	@mkdir -p $(TGZ_PATH)
	git archive HEAD | tar -x -C $(TGZ_PATH)
	@cd build/ && tar czf bitmask-vpn_$(VERSION).tgz ${TGZ_NAME}
	@rm -rf $(TGZ_PATH)

# FIXME port --------------------------------------------------------------------------------------------------

gen_pkg_deb:
	@cp -r ${TEMPLATES}/debian build/${PROVIDER}
	@VERSION=${VERSION} VENDOR_PATH=${VENDOR_PATH} ${SCRIPTS}/generate-debian build/${PROVIDER}/debian/data.json
	@mkdir -p build/${PROVIDER}/debian/icons/scalable && cp ${VENDOR_PATH}/${PROVIDER}/assets/icon.svg build/${PROVIDER}/debian/icons/scalable/icon.svg
	@cd build/${PROVIDER}/debian && python3 generate.py
	@cd build/${PROVIDER}/debian && rm app.desktop-template changelog-template rules-template control-template generate.py data.json && chmod +x rules

gen_pkg_snap:
	@cp -r ${TEMPLATES}/snap build/${PROVIDER}
	@VERSION=${VERSION} VENDOR_PATH=${VENDOR_PATH} ${SCRIPTS}/generate-snap build/${PROVIDER}/snap/data.json
	@cp helpers/se.leap.bitmask.snap.policy build/${PROVIDER}/snap/local/pre/
	@cp helpers/bitmask-root build/${PROVIDER}/snap/local/pre/
	@cd build/${PROVIDER}/snap && python3 generate.py
	@rm build/${PROVIDER}/snap/data.json build/${PROVIDER}/snap/snapcraft-template.yaml
	@mkdir -p build/${PROVIDER}/snap/gui && cp ${VENDOR_PATH}/${PROVIDER}/assets/icon.svg build/${PROVIDER}/snap/gui/icon.svg
	# FIXME is this png needed?? then add it to ASSETS_REQUIRED
	@cp ${VENDOR_PATH}/${PROVIDER}/assets/icon.png build/${PROVIDER}/snap/gui/${PROVIDER}-vpn.png
	@rm build/${PROVIDER}/snap/generate.py

# ---------------------------------------------------------------------------------------------------------------------


#########################################################################
# packaging action
#########################################################################

builder_image:
	@make -C docker build

packages: package_deb package_snap package_osx package_win

package_snap_in_docker:
	@make -C docker package_snap

package_snap:
	@unlink snap || true
	@ln -s build/${PROVIDER}/snap snap
	@make -C build/${PROVIDER} pkg_snap

package_deb:
	@make -C build/${PROVIDER} pkg_deb


#########################################################################
# icons & locales
#########################################################################

icon:
	@make -C icon


LANGS ?= $(foreach path,$(wildcard gui/i18n/main_*.ts),$(patsubst gui/i18n/main_%.ts,%,$(path)))

locales: $(foreach lang,$(LANGS),get_$(lang))

generate_locales:
	@lupdate bitmask.pro

get_%:
	@curl -L -X GET --user "api:${API_TOKEN}" "https://www.transifex.com/api/2/project/bitmask/resource/riseupvpn-test/translation/${subst -,_,$*}/?file" > gui/i18n/main_$*.ts

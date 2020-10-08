#########################################################################
# Multiplatform build and packaging recipes for BitmaskVPN
# (c) LEAP Encryption Access Project, 2019-2020
#########################################################################

.PHONY: all get build icon locales generate_locales clean check_qtifw HAS-qtifw

XBUILD ?= no
SKIP_CACHECK ?= no
PROVIDER ?= $(shell grep ^'provider =' branding/config/vendor.conf | cut -d '=' -f 2 | tr -d "[:space:]")
APPNAME ?= $(shell branding/scripts/getparam appname | tail -n 1)
TARGET ?= $(shell branding/scripts/getparam binname | tail -n 1)
PROVIDER_CONFIG ?= branding/config/vendor.conf
DEFAULT_PROVIDER = branding/assets/default/
VERSION ?= $(shell git describe)

# go paths
GOPATH = $(shell go env GOPATH)
TARGET_GOLIB=lib/libgoshim.a
SOURCE_GOLIB=gui/backend.go

# detect OS, we use it for dependencies
UNAME = $(shell uname -s)
PLATFORM ?= $(shell echo ${UNAME} | awk "{print tolower(\$$0)}")

QTBUILD = build/qt
INSTALLER = build/installer
WININST_DATA = branding/qtinstaller/packages/root.win_x86_64/data/
OSX_DATA = ${INSTALLER}/packages/bitmaskvpn/data/
OSX_CERT="Developer ID Installer: LEAP Encryption Access Project"
MACDEPLOYQT_OPTS = -appstore-compliant -qmldir=gui/qml -always-overwrite
# XXX expired cert -codesign="${OSX_CERT}"
	
SCRIPTS = branding/scripts
TEMPLATES = branding/templates

HAS_QTIFW := $(shell PATH=$(PATH) which binarycreator)
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

golib:
	# TODO stop building golib in gui/build.sh, it's redundant. 
	# we should port the buildGoLib parts of the gui/build.sh script here
	@echo "doing nothing"

build: golib build_helper build_openvpn
	@XBUILD=no TARGET=${TARGET} gui/build.sh

build_helper:
	@echo "PLATFORM: ${PLATFORM}"
	@mkdir -p build/bin/${PLATFORM}
	go build -o build/bin/${PLATFORM}/bitmask-helper -ldflags "-X main.AppName=${APPNAME} -X main.Version=${VERSION}" ./cmd/bitmask-helper/

build_old:
ifeq (${XBUILD}, yes)
	$(MAKE) build_cross_win
	$(MAKE) build_cross_osx
	$(MAKE) _build_xbuild_done
else ifeq (${XBUILD}, win)
	$(MAKE) build_cross_win
	$(MAKE) _build_done
else ifeq (${XBUILD}, osx)
	$(MAKE) build_cross_osx
	$(MAKE) _build_done
else
	@gui/build.sh
endif

build_openvpn:
	@[ -f $(OPENVPN_BIN) ] && echo "OpenVPN already built at" $(OPENVPN_BIN) || ./branding/thirdparty/openvpn/build_openvpn.sh

debug_installer:
	@VERSION=${VERSION} ${SCRIPTS}/gen-qtinstaller osx ${INSTALLER}

build_installer: check_qtifw build
	echo "mkdir osx data"	
	@mkdir -p ${OSX_DATA}
	@cp -r ${TEMPLATES}/qtinstaller/packages ${INSTALLER}
	@cp -r ${TEMPLATES}/qtinstaller/installer.pro ${INSTALLER}
	@cp -r ${TEMPLATES}/qtinstaller/config ${INSTALLER}
ifeq (${PLATFORM}, darwin)
	@mkdir -p ${OSX_DATA}/helper
	# TODO need to write this
	@VERSION=${VERSION} ${SCRIPTS}/gen-qtinstaller osx ${INSTALLER}
	@cp "${TEMPLATES}/osx/bitmask.pf.conf" ${OSX_DATA}/helper/bitmask.pf.conf
	@cp "${TEMPLATES}/osx/client.up.sh" ${OSX_DATA}/
	@cp "${TEMPLATES}/osx/client.down.sh" ${OSX_DATA}/
	@cp "${TEMPLATES}/qtinstaller/osx-data/post-install.py" ${OSX_DATA}/
	@cp "${TEMPLATES}/qtinstaller/osx-data/uninstall.py" ${OSX_DATA}/
	@cp "${TEMPLATES}/qtinstaller/osx-data/se.leap.bitmask-helper.plist" ${OSX_DATA}/
	@cp build/bin/${PLATFORM}/bitmask-helper ${OSX_DATA}/
	# FIXME our static openvpn build fails with an "Assertion failed at crypto.c". Needs to be fixed!!! - kali
	# a working (old) version:
	#@curl -L https://downloads.leap.se/thirdparty/osx/openvpn/openvpn -o build/${PROVIDER}/staging/openvpn-osx
	#@cp $(OPENVPN_BIN) ${OSX_DATA}/openvpn.leap
	@echo "WARNING: workaround for broken static build. Shipping homebrew dynamically linked instead"
	@rm -f ${OSX_DATA}openvpn.leap && cp /usr/local/bin/openvpn ${OSX_DATA}openvpn.leap
	@echo "[+] Running macdeployqt"
	@macdeployqt ${QTBUILD}/release/${PROVIDER}-vpn.app ${MACDEPLOYQT_OPTS}
	@cp -r "${QTBUILD}/release/${TARGET}.app"/ ${OSX_DATA}/
endif
ifeq (${PLATFORM}, windows)
	@${SCRIPTS}/gen-qtinstaller windows ${INSTALLER}
endif
ifeq (${PLATFORM}, linux)
	@${SCRIPTS}/gen-qtinstaller windows ${INSTALLER}
endif
	@echo "[+] All templates, binaries and libraries copied to build/installer."
	@echo "[+] Now building the installer."
	@cd build/installer && qmake INSTALLER=${APPNAME}-installer-${VERSION} && make

installer_win:
	# XXX refactor with build_installer
	cp helper.exe ${WININST_DATA}
	cp ${QTBUILD}/release/${TARGET}.exe ${WININST_DATA}${TARGET}.exe
	# XXX add sign step here
	windeployqt --qmldir gui/qml ${WININST_DATA}${TARGET}.exe
	"/c/Qt/QtIFW-3.2.2/bin/binarycreator.exe" -c ./branding/qtinstaller/config/config.xml -p ./branding/qtinstaller/packages build/${PROVIDER}-vpn-${VERSION}-installer.exe

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
	$(CROSS_WIN_FLAGS) $(PLATFORM_WIN) $(EXTRA_LDFLAGS_WIN) $(MAKE) _buildparts
	# workaround for helper: we use the go compiler
	@echo "[+] Compiling helper with the Go compiler to work around missing stdout bug..."
	cd cmd/bitmask-helper && GOOS=windows GOARCH=386 go build -ldflags "-X main.version=`git describe --tags` -H windowsgui" -o ../../build/bin/windows/bitmask-helper-go

CROSS_OSX_FLAGS = MACOSX_DEPLOYMENT_TARGET=10.10 CGO_ENABLED=1 GOOS=darwin CC="o64-clang"
PLATFORM_OSX = PLATFORM=darwin
build_cross_osx:
	$(CROSS_OSX_FLAGS) $(PLATFORM_OSX) $(MAKE) _buildparts

_build_done:
	@echo
	@echo 'Done. You can build your package now.'

_build_xbuild_done:
	@echo
	@echo 'Done. You can do "make packages" now.'

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
	@./branding/scripts/init
	# TODO we should do the prepare step here, store it in VENDOR_PATH

vendor_check:
	@./branding/scripts/check
	# TODO move ca-check here

vendor: gen_providers_json
	# TODO merge with prepare, after moving the gen_pkg to vendor_init

gen_providers_json:
	@python3 branding/scripts/gen-providers-json.py branding/config/vendor.conf gui/providers/providers.json

prepare: prepare_templates gen_pkg_win gen_pkg_osx gen_pkg_snap gen_pkg_deb prepare_done

prepare_templates: generate relink_default tgz
	@mkdir -p build/${PROVIDER}/bin/ deploy
	@cp ${TEMPLATES}/makefile/Makefile build/${PROVIDER}/Makefile
	@VERSION=${VERSION} PROVIDER_CONFIG=${PROVIDER_CONFIG} ${SCRIPTS}/generate-vendor-make.py build/${PROVIDER}/vendor.mk
ifeq (${SKIP_CACHECK}, no)
	@${SCRIPTS}/check-ca-crt.py ${PROVIDER} ${PROVIDER_CONFIG}
endif

generate:
	@go generate gui/backend.go
	@go generate pkg/config/version/genver/gen.go

relink_default:
ifneq (,$(wildcard ${DEFAULT_PROVIDER}))
	@cd branding/assets && unlink default
endif
	@cd branding/assets && ln -s ${PROVIDER} default

TGZ_NAME = bitmask-vpn_${VERSION}-src
TGZ_PATH = $(shell pwd)/build/${TGZ_NAME}
tgz:
	@mkdir -p $(TGZ_PATH)
	git archive HEAD | tar -x -C $(TGZ_PATH)
	@cd build/ && tar czf bitmask-vpn_$(VERSION).tgz ${TGZ_NAME}
	@rm -rf $(TGZ_PATH)

# XXX port/deprecate -----------------------------------------------
gen_pkg_win:
	@mkdir -p build/${PROVIDER}/windows/
	@cp -r ${TEMPLATES}/windows build/${PROVIDER}
	@VERSION=${VERSION} PROVIDER_CONFIG=${PROVIDER_CONFIG} ${SCRIPTS}/generate-win.py build/${PROVIDER}/windows/data.json
	@cd build/${PROVIDER}/windows && python3 generate.py

gen_pkg_deb:
	@cp -r ${TEMPLATES}/debian build/${PROVIDER}
	@VERSION=${VERSION} PROVIDER_CONFIG=${PROVIDER_CONFIG} ${SCRIPTS}/generate-debian.py build/${PROVIDER}/debian/data.json
	@mkdir -p build/${PROVIDER}/debian/icons/scalable && cp branding/assets/default/icon.svg build/${PROVIDER}/debian/icons/scalable/icon.svg
	@cd build/${PROVIDER}/debian && python3 generate.py
	@cd build/${PROVIDER}/debian && rm app.desktop-template changelog-template rules-template control-template generate.py data.json && chmod +x rules

gen_pkg_snap:
	@cp -r ${TEMPLATES}/snap build/${PROVIDER}
	@VERSION=${VERSION} PROVIDER_CONFIG=${PROVIDER_CONFIG} ${SCRIPTS}/generate-snap.py build/${PROVIDER}/snap/data.json
	@cp helpers/se.leap.bitmask.snap.policy build/${PROVIDER}/snap/local/pre/
	@cp helpers/bitmask-root build/${PROVIDER}/snap/local/pre/
	@cd build/${PROVIDER}/snap && python3 generate.py
	@rm build/${PROVIDER}/snap/data.json build/${PROVIDER}/snap/snapcraft-template.yaml
	@mkdir -p build/${PROVIDER}/snap/gui && cp branding/assets/default/icon.svg build/${PROVIDER}/snap/gui/icon.svg
	@cp branding/assets/default/icon.png build/${PROVIDER}/snap/gui/${PROVIDER}-vpn.png
	rm build/${PROVIDER}/snap/generate.py

prepare_done:
	@echo
	@echo 'Done. You can do "make build" now.'

#########################################################################
# packaging action
#########################################################################

builder_image:
	@make -C docker build

packages: package_deb package_snap package_osx package_win

package_snap_in_docker:
	@make -C docker package_snap

package_win_in_docker:
	@make -C docker package_win

package_snap:
	@unlink snap || true
	@ln -s build/${PROVIDER}/snap snap
	@make -C build/${PROVIDER} pkg_snap

package_deb:
	@make -C build/${PROVIDER} pkg_deb

package_osx:
	@echo "tbd"


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

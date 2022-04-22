#########################################################################
# Multiplatform build and packaging recipes for BitmaskVPN
# (c) LEAP Encryption Access Project, 2019-2021
#########################################################################

.PHONY: all get build icon locales generate_locales clean check_qtifw HAS-qtifw relink_vendor

XBUILD ?= no
RELEASE ?= no
QMAKE ?= qmake
LRELEASE ?= lrelease
SKIP_CACHECK ?= no
VENDOR_PATH ?= providers
APPNAME != VENDOR_PATH=${VENDOR_PATH} branding/scripts/getparam appname | tail -n 1
TARGET != VENDOR_PATH=${VENDOR_PATH} branding/scripts/getparam binname | tail -n 1
PROVIDER != grep ^'provider =' ${VENDOR_PATH}/vendor.conf | cut -d '=' -f 2 | cut -d ',' -f 1 | tr -d "[:space:]"
VERSION != git describe 2>/dev/null || echo -n "unknown"
WINCERTPASS ?= pass
OSXAPPPASS  ?= pass
OSXMORDORUID ?= uid

# go paths
GOPATH != go env GOPATH
TARGET_GOLIB=lib/libgoshim.a
SOURCE_GOLIB=gui/backend.go

# detect OS
UNAME != uname -s
PLATFORM != [ '$(UNAME)' = 'Windows_NT' ] && echo -n 'windows' || (echo ${UNAME} | awk "{print tolower(\$$0)}")

QTBUILD = build/qt
INSTALLER = build/installer
OSX_CERT="Developer ID Application: LEAP Encryption Access Project"
MACDEPLOYQT_OPTS = -appstore-compliant -always-overwrite -codesign="${OSX_CERT}"

ifeq ($(PLATFORM), darwin)
INST_ROOT =${INSTALLER}/packages/bitmaskvpn/data/
INST_DATA = ${INST_ROOT}/${APPNAME}.app/
else
INST_DATA = ${INSTALLER}/packages/bitmaskvpn/data/
endif

SCRIPTS = branding/scripts
TEMPLATES = branding/templates

TAP_WINDOWS = https://build.openvpn.net/downloads/releases/tap-windows-9.24.2-I601-Win10.exe

HAS_QTIFW != which binarycreator.exe 2>/dev/null || PATH=$(PATH) which binarycreator
OPENVPN_BIN != echo -n "$(HOME)/openvpn_build/sbin/$$(grep OPENVPN branding/thirdparty/openvpn/build_openvpn.sh | head -n 1 | cut -d = -f 2 | tr -d '"')"


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
	-@${MAKE} depends$(UNAME)

dependsLinux:
	@sudo apt install golang pkg-config dh-golang golang-golang-x-text-dev cmake devscripts fakeroot debhelper curl g++ qt5-qmake qttools5-dev-tools qtdeclarative5-dev qml-module-qtquick-controls2 libqt5qml5 qtdeclarative5-dev qtquickcontrols2-5-dev libqt5svg5-dev qml-module-qt-labs-platform qml-module-qtquick-extras qml-module-qtquick-dialogs
	@${MAKE} -C docker deps
	@# debian needs also: snap install snapcraft --classic; snap install  multipass --beta --classic

dependsDarwin:
	@brew install git golang make qt5
	#@brew install --default-names gnu-sed
	@brew link qt5

dependsCYGWIN_NT-10.0:
	@echo
	@echo "==================================WARNING=================================="
	@echo "You need to install all dependencies manually, please see README.md!"
	@echo "==================================WARNING=================================="
	@echo

EXTRA_FLAGS != [ $(PLATFORM) = 'darwin' ] && echo -n MACOSX_DEPLOYMENT_TARGET=10.10 GOOS=darwin CC=clang
EXTRA_GO_LDFLAGS != [ $(PLATFORM) = 'windows' ] && echo -n '-H=windowsgui'

ifeq ($(PLATFORM), windows)
PKGFILES = $(wildcard "pkg/*") # syntax err in windows with find 
else
PKGFILES != find pkg -type f -name '*.go'
endif

lib/%.a: $(PKGFILES)
	@XBUILD=no CC=${CC} CXX=${CXX} MAKE=${MAKE} AR=${AR} LD=${LD} ./gui/build.sh --just-golib

# FIXME move platform detection above! no place to uname here, just use $PLATFORM
#
MINGGW = 
ifeq ($(UNAME), MINGW64_NT-10.0)
MINGW = yes
endif
ifeq ($(UNAME), MINGW64_NT-10.0-19042)
MINGW = yes
endif

relink_vendor:
	@echo "============RELINK VENDOR============="
	@echo "PLATFORM: ${PLATFORM} (${UNAME})"
	@echo "VENDOR_PATH: ${VENDOR_PATH}"
	@echo "PROVIDER: ${PROVIDER}"
ifeq ($(PLATFORM), windows)
	@rm -rf providers/assets || true
ifeq ($(MINGW), yes)
ifeq ($(VENDOR_PATH), providers)
	@cp -r providers/${PROVIDER}/assets providers/assets || true
endif
endif # end mingw
ifeq ($(UNAME), CYGWIN_NT-10.0)
	@[ -L providers/assets ] || (CYGWIN=winsymlinks:nativestrict ln -s ${PROVIDER}/assets providers/assets)
endif # end cygwin
else # not windows: linux/osx
ifeq ($(VENDOR_PATH), providers)
	@unlink providers/assets || true
	@ln -s ${PROVIDER}/assets providers/assets || true
endif
endif
	@echo "============RELINK VENDOR============="

build_golib: lib/libgoshim.a

build_gui: build_golib relink_vendor
	@echo "==============BUILD GUI==============="
	@echo "TARGET: ${TARGET}"
	@echo "VENDOR_PATH: ${VENDOR_PATH}"
	@XBUILD=no CC=${CC} CXX=${CXX} MAKE=${MAKE} AR=${AR} LD=${LD} QMAKE=${QMAKE} LRELEASE=${LRELEASE} TARGET=${TARGET} VENDOR_PATH=${VENDOR_PATH} APPNAME=${APPNAME} gui/build.sh --skip-golib
	@echo "============BUILD GUI================="

build: build_helper build_gui

build_helper:
ifeq ($(PLATFORM), linux)
# no helper needed for linux, we use polkit/bitmask-root
else
	@echo "=============BUILDER HELPER==========="
	@echo "PLATFORM: ${PLATFORM}"
	@echo "APPNAME: ${APPNAME}"
	@echo "VERSION: ${VERSION}"
	@echo "EXTRA_GO_LDFLAGS: ${EXTRA_GO_LDFLAGS}"
	@mkdir -p build/bin/${PLATFORM}
	@go build -o build/bin/${PLATFORM}/bitmask-helper -ldflags "-X main.AppName=${APPNAME} -X main.Version=${VERSION} ${EXTRA_GO_LDFLAGS}" ./cmd/bitmask-helper/
	@echo "===========BUILDER HELPER============="
endif

build_openvpn:
	@[ -f $(OPENVPN_BIN) ] && echo "OpenVPN already built at" $(OPENVPN_BIN) || ./branding/thirdparty/openvpn/build_openvpn.sh

dosign:
ifeq (${PLATFORM}, windows)
	"c:\windows\system32\rcedit.exe" ${QTBUILD}/release/${TARGET}.exe --set-file-version ${VERSION}
	"c:\windows\system32\rcedit.exe" ${QTBUILD}/release/${TARGET}.exe --set-product-version ${VERSION}
	"c:\windows\system32\rcedit.exe" ${QTBUILD}/release/${TARGET}.exe --set-version-string CompanyName "LEAP Encryption Access Project"
	"c:\windows\system32\rcedit.exe" ${QTBUILD}/release/${TARGET}.exe --set-version-string FileDescription "${APPNAME}"
	"c:\windows\system32\signtool.exe" sign -debug -f "z:\leap\LEAP.pfx" -p ${WINCERTPASS} ${QTBUILD}/release/${TARGET}.exe
	# XXX need to deprecate helper and embrace interactive service
	cp build/bin/${PLATFORM}/bitmask-helper build/bin/${PLATFORM}/bitmask-helper.exe
	"c:\windows\system32\rcedit.exe" build/bin/${PLATFORM}/bitmask-helper.exe --set-file-version ${VERSION}
	"c:\windows\system32\rcedit.exe" build/bin/${PLATFORM}/bitmask-helper.exe --set-product-version ${VERSION}
	"c:\windows\system32\rcedit.exe" build/bin/${PLATFORM}/bitmask-helper.exe --set-version-string ProductName "bitmask-helper-v2"
	"c:\windows\system32\rcedit.exe" build/bin/${PLATFORM}/bitmask-helper.exe --set-version-string CompanyName "LEAP Encryption Access Project"
	"c:\windows\system32\rcedit.exe" build/bin/${PLATFORM}/bitmask-helper.exe --set-version-string FileDescription "Administrative helper for ${APPNAME}"
	"c:\windows\system32\signtool.exe" sign -debug -f "z:\leap\LEAP.pfx" -p ${WINCERTPASS} build/bin/${PLATFORM}/bitmask-helper.exe
endif

checksign:
ifeq (${PLATFORM}, windows)
	@"c:\windows\system32\sigcheck.exe" ${QTBUILD}/release/${TARGET}.exe
	@"c:\windows\system32\sigcheck.exe" build/bin/${PLATFORM}/bitmask-helper.exe
	@"c:\windows\system32\sigcheck.exe" "/c/Program Files/OpenVPN/bin/openvpn.exe"
endif

installer: check_qtifw checksign
	@mkdir -p ${INST_DATA}
	@cp -r ${TEMPLATES}/qtinstaller/packages ${INSTALLER}
	@cp -r ${TEMPLATES}/qtinstaller/installer.pro ${INSTALLER}
	@cp -r ${TEMPLATES}/qtinstaller/config ${INSTALLER}
	@cp ${VENDOR_PATH}/assets/icon.ico ${INSTALLER}/config/installer-icon.ico
	@cp ${VENDOR_PATH}/assets/icon.icns ${INSTALLER}/config/installer-icon.icns
	@cp ${VENDOR_PATH}/assets/installer-logo.png ${INSTALLER}/config/installer-logo.png
ifeq (${PLATFORM}, darwin)
	@mkdir -p ${INST_DATA}/helper
	@VERSION=${VERSION} VENDOR_PATH=${VENDOR_PATH} ${SCRIPTS}/gen-qtinstaller osx ${INSTALLER}
	@cp "${TEMPLATES}/osx/bitmask.pf.conf" ${INST_DATA}helper/bitmask.pf.conf
	@cp "${TEMPLATES}/osx/client.up.sh" ${INST_DATA}/
	@cp "${TEMPLATES}/osx/client.down.sh" ${INST_DATA}/
	@cp "${TEMPLATES}/qtinstaller/osx-data/post-install.py" ${INST_ROOT}/
	@cp "${TEMPLATES}/qtinstaller/osx-data/uninstall.py" ${INST_ROOT}/
	@cp "${TEMPLATES}/qtinstaller/osx-data/se.leap.bitmask-helper.plist" ${INST_DATA}
	@cp $(OPENVPN_BIN) ${INST_DATA}/openvpn.leap
	@cp build/bin/${PLATFORM}/bitmask-helper ${INST_DATA}/
ifeq (${RELEASE}, yes)
	@echo "[+] Running macdeployqt (release mode)"
	@macdeployqt ${QTBUILD}/release/${PROVIDER}-vpn.app -qmldir=gui/components ${MACDEPLOYQT_OPTS}
else
	@echo "[+] Running macdeployqt (debug mode)"
	@macdeployqt ${QTBUILD}/release/${PROVIDER}-vpn.app -qmldir=gui/components
endif
	@cp -r "${QTBUILD}/release/${TARGET}.app"/ ${INST_DATA}/
endif
ifeq (${PLATFORM}, windows)
	@VERSION=${VERSION} VENDOR_PATH=${VENDOR_PATH} ${SCRIPTS}/gen-qtinstaller windows ${INSTALLER}
	@cp build/bin/${PLATFORM}/bitmask-helper.exe ${INST_DATA}helper.exe
ifeq (${VENDOR_PATH}, providers)
	@cp ${VENDOR_PATH}/${PROVIDER}/assets/icon.ico ${INST_DATA}/icon.ico
else
	@cp ${VENDOR_PATH}/assets/icon.ico ${INST_DATA}/icon.ico
endif
	@cp ${QTBUILD}/release/${TARGET}.exe ${INST_DATA}${TARGET}.exe
	@cp "/c/Program Files/OpenVPN/bin/openvpn.exe" ${INST_DATA}
	@cp "/c/Program Files/OpenVPN/bin/"*.dll ${INST_DATA}
ifeq (${RELEASE}, yes)
	#@windeployqt --release --qmldir gui/components ${INST_DATA}${TARGET}.exe
	#FIXME -- cannot find platform plugin
	@windeployqt --qmldir gui/components ${INST_DATA}${TARGET}.exe
else
	@windeployqt --qmldir gui/components ${INST_DATA}${TARGET}.exe
endif
	# TODO stage it to shave some time
	@wget ${TAP_WINDOWS} -O ${INST_DATA}/tap-windows.exe
	# XXX this is a workaround for missing libs after windeployqt ---
	@cp /c/Qt/5.15.2/mingw81_64/bin/libgcc_s_seh-1.dll ${INST_DATA}
	@cp /c/Qt/5.15.2/mingw81_64/bin/libstdc++-6.dll ${INST_DATA}
	@cp /c/Qt/5.15.2/mingw81_64/bin/libwinpthread-1.dll ${INST_DATA}
	@cp -r /c/Qt/5.15.2/mingw81_64/qml ${INST_DATA}
endif
ifeq (${PLATFORM}, linux)
	@VERSION=${VERSION} ${SCRIPTS}/gen-qtinstaller linux ${INSTALLER}
endif
	@echo "[+] All templates, binaries and libraries copied to build/installer."
	@echo "[+] Now building the installer."
	@cd build/installer && ${QMAKE} VENDOR_PATH=${VENDOR_PATH} INSTALLER=${APPNAME}-installer-${VERSION} && ${MAKE}

sign_installer:
ifeq (${PLATFORM}, windows)
	# TODO add flag to skip signing for regular builds
	"c:\windows\system32\signtool.exe" sign -f "z:\leap\LEAP.pfx" -p ${WINCERTPASS} build/installer/${APPNAME}-installer-${VERSION}.exe
endif
ifeq (${PLATFORM}, darwin)
	gsed -i "s/com.yourcompany.installerbase/se.leap.bitmask.${TARGET}/g" build/installer/${APPNAME}-installer-${VERSION}.app/Contents/Info.plist
	codesign -s ${OSX_CERT} --options "runtime" build/installer/${APPNAME}-installer-${VERSION}.app
	ditto -ck --rsrc --sequesterRsrc build/installer/${APPNAME}-installer-${VERSION}.app build/installer/${APPNAME}-installer-${VERSION}.zip
endif

notarize_all:
	APPNAME=${APPNAME} VERSION=${VERSION} TARGET=${TARGET} OSXAPPPASS=${OSXAPPPASS} branding/scripts/osx-stapler.sh

# --------------------
# TODO test and remove

notarize_installer:
# courtesy of https://skyronic.com/2019/07/app-notarization-for-qt-applications/
ifeq (${PLATFORM}, darwin)
	xcrun altool --notarize-app -t osx -f build/installer/${APPNAME}-installer-${VERSION}.zip --primary-bundle-id="se.leap.bitmask.${TARGET}" -u "info@leap.se" -p ${OSXAPPPASS}
endif

notarize_check:
ifeq (${PLATFORM}, darwin)
	xcrun altool --notarization-info ${OSXMORDORUID} -u "info@leap.se" -p ${OSXAPPPASS}
endif

notarize_staple:
ifeq (${PLATFORM}, darwin)
	xcrun stapler staple build/installer/${APPNAME}-installer-${VERSION}.app
endif

create_dmg:
ifeq (${PLATFORM}, darwin)
	@create-dmg deploy/${APPNAME}-${VERSION}.dmg build/installer/${APPNAME}-installer-${VERSION}.app
endif
# --------------------



check_qtifw:
ifdef HAS_QTIFW
	@echo "[+] Found QTIFW"
else
	$(error "[!] Cannot find QTIFW. Please install it and add it to your PATH")
endif

clean:
	@rm -rf lib/*
	@rm -rf build/
	@unlink branding/assets/default || true


########################################################################
# tests
#########################################################################

qmllint:
	@qmllint gui/*.qml
	@qmllint gui/components/*.qml

qmlfmt:
	# needs https://github.com/jesperhh/qmlfmt in your path
	@qmlfmt -w gui/qml/*.qml

test:
	@go test -tags "integration $(TAGS)" ./pkg/...

test_ui: build_golib
	@${QMAKE} -o tests/Makefile test.pro
	@${MAKE} -C tests clean
	@${MAKE} -C tests
ifeq ($(PLATFORM), windows)
	@./tests/build/test_ui.exe
else
	@./tests/build/test_ui
endif


#########################################################################
# packaging templates
#########################################################################

bump_snap:
	@sed -i 's/^version:.*$$/version: ${VERSION}/' snap/snapcraft.yaml
	@sed -i 's/^.*echo .*version.txt$$/        echo ${VERSION} > $$SNAPCRAFT_PRIME\/snap\/version.txt/' snap/snapcraft.yaml

local_snap:
	# just to be able to debug stuff locally in the same way as it's really built @canonical
	# but multipass is the way to go, nowadays
	@snapcraft --debug --use-lxd

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


gen_pkg_deb:
ifeq (${PLATFORM}, linux)
	@cp -r ${TEMPLATES}/debian build/${PROVIDER}
	@VERSION=${VERSION} VENDOR_PATH=${VENDOR_PATH} ${SCRIPTS}/generate-debian build/${PROVIDER}/debian/data.json
ifeq (${VENDOR_PATH}, providers)
	@mkdir -p build/${PROVIDER}/debian/icons/scalable && cp ${VENDOR_PATH}/${PROVIDER}/assets/icon.svg build/${PROVIDER}/debian/icons/scalable/icon.svg
else
	@mkdir -p build/${PROVIDER}/debian/icons/scalable && cp ${VENDOR_PATH}/assets/icon.svg build/${PROVIDER}/debian/icons/scalable/icon.svg
endif
	@cd build/${PROVIDER}/debian && python3 generate.py
	@cd build/${PROVIDER}/debian && rm app.desktop-template changelog-template rules-template control-template generate.py data.json && chmod +x rules
endif

gen_pkg_snap:
ifeq (${PLATFORM}, linux)
	@cp -r ${TEMPLATES}/snap build/${PROVIDER}
	@VERSION=${VERSION} VENDOR_PATH=${VENDOR_PATH} ${SCRIPTS}/generate-snap build/${PROVIDER}/snap/data.json
	@cp pkg/pickle/helpers/se.leap.bitmask.snap.policy build/${PROVIDER}/snap/local/pre/
	@cp pkg/pickle/helpers/bitmask-root build/${PROVIDER}/snap/local/pre/
	@cd build/${PROVIDER}/snap && python3 generate.py
	@rm build/${PROVIDER}/snap/data.json build/${PROVIDER}/snap/snapcraft-template.yaml
	@mkdir -p build/${PROVIDER}/snap/gui
ifeq (${VENDOR_PATH}, providers)
	@cp ${VENDOR_PATH}/${PROVIDER}/assets/icon.svg build/${PROVIDER}/snap/gui/icon.svg
	@cp ${VENDOR_PATH}/${PROVIDER}/assets/icon.png build/${PROVIDER}/snap/gui/${PROVIDER}-vpn.png
else
	@cp ${VENDOR_PATH}/assets/icon.svg build/${PROVIDER}/snap/gui/icon.svg
	@cp ${VENDOR_PATH}/assets/icon.png build/${PROVIDER}/snap/gui/${PROVIDER}-vpn.png
endif
	@rm build/${PROVIDER}/snap/generate.py
endif


#########################################################################
# packaging action
#########################################################################
run:
	./build/qt/release/riseup-vpn

builder_image:
	@${MAKE} -C docker build

packages: package_deb package_snap package_osx package_win

package_win_release: build dosign installer sign_installer

package_win: build installer

package_snap_in_docker:
	@${MAKE} -C docker package_snap

package_snap:
	@unlink snap || true
	@cp build/${PROVIDER}/snap/local/${TARGET}.desktop build/${PROVIDER}/snap/gui/
	@ln -s build/${PROVIDER}/snap snap
	@${MAKE} -C build/${PROVIDER} pkg_snap

package_deb:
	@${MAKE} -C build/${PROVIDER} pkg_deb

sign_artifact:
	@find ${FILE} -type f -not -name "*.asc" -print0 | xargs -0 -n1 -I{} sha256sum -b "{}" | sed 's/*deploy\///' > ${FILE}.sha256
	@gpg --clear-sign --armor ${FILE}.sha256

upload_artifact:
	scp ${FILE} downloads.leap.se:./
	scp ${FILE}.sha256.asc downloads.leap.se:./


#########################################################################
# icons & locales
#########################################################################

icon:
	@${MAKE} -C icon


LANGS ?= $(foreach path,$(wildcard gui/i18n/main_*.ts),$(patsubst gui/i18n/main_%.ts,%,$(path)))

locales: $(foreach lang,$(LANGS),get_$(lang))

generate_locales:
	@lupdate bitmask.pro

get_%:
	@curl -L -X GET --user "api:${API_TOKEN}" "https://www.transifex.com/api/2/project/bitmask/resource/bitmask-desktop/translation/${subst -,_,$*}/?file" > gui/i18n/main_$*.ts

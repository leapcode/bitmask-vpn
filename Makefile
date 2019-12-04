#########################################################################
# Multiplatform build and packaging recipes for BitmaskVPN
# (c) LEAP Encryption Access Project, 2019
#########################################################################

.PHONY: all get build build_bitmaskd icon locales generate_locales clean

TAGS ?= gtk_3_18

XBUILD ?= no
SKIP_CACHECK ?= no
PROVIDER ?= $(shell grep ^'provider =' branding/config/vendor.conf | cut -d '=' -f 2 | tr -d "[:space:]")
PROVIDER_CONFIG ?= branding/config/vendor.conf
DEFAULT_PROVIDER = branding/assets/default/
VERSION ?= $(shell git describe)

# go paths
GOPATH = $(shell go env GOPATH)
SYSTRAY = 0xacab.org/leap/bitmask-vpn
GOSYSTRAY = ${GOPATH}/src/${SYSTRAY}

# detect OS, we use it for dependencies
UNAME = $(shell uname -s)
PLATFORM ?= $(shell echo ${UNAME} | awk "{print tolower(\$$0)}")

TEMPLATES = branding/templates
SCRIPTS = branding/scripts

all: icon locales build


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
	@go get -u golang.org/x/text/cmd/gotext github.com/cratonica/2goarray

dependsLinux:
	@sudo apt install libgtk-3-dev libappindicator3-dev golang pkg-config dh-golang golang-golang-x-text-dev cmake devscripts fakeroot debhelper curl
	@# debian needs also: snap install snapcraft --classic; snap install  multipass --beta --classic

dependsDarwin:
	# TODO - bootstrap homebrew if not there
	@brew install python3 golang make pkg-config upx curl
	@brew install --default-names gnu-sed

dependsCygwin:
	@choco install -y golang python nssm nsis wget 7zip

build:
	$(MAKE) _buildparts
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
endif

_buildparts: $(foreach path,$(wildcard cmd/*),build_$(patsubst cmd/%,%,$(path)))

build_%:
	@mkdir -p build/bin/${PLATFORM}
	go build -tags $(TAGS) -ldflags "-s -w -X main.version=`git describe --tags`" -o build/bin/${PLATFORM}/$* ./cmd/$*
	-@rm -rf build/${PROVIDER}/staging/${PLATFORM} && mkdir -p build/${PROVIDER}/staging/${PLATFORM}
	-@ln -s ../../../bin/${PLATFORM}/$* build/${PROVIDER}/staging/${PLATFORM}/$*

test:
	@go test -tags "integration $(TAGS)" ./...

build_bitmaskd:
	@go build -tags "$(TAGS) bitmaskd" -ldflags "-X main.version=`git describe --tags`" ./cmd/*

build_win:
	powershell -Command '$$version=git describe --tags; go build -ldflags "-H windowsgui -X main.version=$$version" ./cmd/*'

CROSS_WIN_FLAGS = CGO_ENABLED=1 GOARCH=386 GOOS=windows CC="/usr/bin/i686-w64-mingw32-gcc" CGO_LDFLAGS="-lssp" CXX="i686-w64-mingw32-c++"
PLATFORM_WIN = PLATFORM=windows
build_cross_win:
	$(CROSS_WIN_FLAGS) $(PLATFORM_WIN) $(MAKE) _buildparts

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

clean:
	@rm -rf build/
	@unlink branding/assets/default

#########################################################################
# build them all
#########################################################################

build_all_providers:
	branding/scripts/build-all-providers

#########################################################################
# packaging templates
#########################################################################

prepare: prepare_templates gen_pkg_win gen_pkg_osx gen_pkg_snap gen_pkg_deb prepare_done

prepare_templates: generate relink_default tgz
	@mkdir -p build/${PROVIDER}/bin/
	@cp ${TEMPLATES}/makefile/Makefile build/${PROVIDER}/Makefile
	@VERSION=${VERSION} PROVIDER_CONFIG=${PROVIDER_CONFIG} ${SCRIPTS}/generate-vendor-make.py build/${PROVIDER}/vendor.mk
ifeq (${SKIP_CACHECK}, no)
	@${SCRIPTS}/check-ca-crt.py ${PROVIDER} ${PROVIDER_CONFIG}
endif

generate:
	@go generate cmd/bitmask-vpn/main.go

relink_default:
ifneq (,$(wildcard ${DEFAULT_PROVIDER}))
	@cd branding/assets && unlink default
endif
	@cd branding/assets && ln -s ${PROVIDER} default

TGZ_NAME = bitmask-vpn_${VERSION}-src
TGZ_PATH = $(shell pwd)/build/${TGZ_NAME}
tgz:
	@mkdir -p $(TGZ_PATH)
	@git archive HEAD | tar -x -C $(TGZ_PATH)
	@cd build/ && tar cvzf bitmask-vpn_$(VERSION).tgz ${TGZ_NAME}
	@rm -rf $(TGZ_PATH)

gen_pkg_win:
	@mkdir -p build/${PROVIDER}/windows/
	@cp -r ${TEMPLATES}/windows build/${PROVIDER}
	@VERSION=${VERSION} PROVIDER_CONFIG=${PROVIDER_CONFIG} ${SCRIPTS}/generate-win.py build/${PROVIDER}/windows/data.json
	@cd build/${PROVIDER}/windows && python3 generate.py
	# TODO create/copy build/PROVIDER/assets/
	# TODO create/copy build/PROVIDER/staging/

gen_pkg_osx:
	@mkdir -p build/${PROVIDER}/osx/scripts
	@mkdir -p build/${PROVIDER}/staging
ifeq (,$(wildcard build/${PROVIDER}/assets))
	@ln -s ../../branding/assets/default build/${PROVIDER}/assets
endif
ifeq (,$(wildcard build/${PROVIDER}/staging/openvpn-osx))
	@curl -L https://downloads.leap.se/thirdparty/osx/openvpn/openvpn -o build/${PROVIDER}/staging/openvpn-osx
endif
	@cp -r ${TEMPLATES}/osx build/${PROVIDER}
	@VERSION=${VERSION} PROVIDER_CONFIG=${PROVIDER_CONFIG} ${SCRIPTS}/generate-osx.py build/${PROVIDER}/osx/data.json
	@cd build/${PROVIDER}/osx && python3 generate.py
	@cd build/${PROVIDER}/osx/scripts && chmod +x preinstall postinstall

gen_pkg_snap:
	@cp -r ${TEMPLATES}/snap build/${PROVIDER}
	@VERSION=${VERSION} PROVIDER_CONFIG=${PROVIDER_CONFIG} ${SCRIPTS}/generate-snap.py build/${PROVIDER}/snap/data.json
	@cp helpers/se.leap.bitmask.snap.policy build/${PROVIDER}/snap/local/pre/
	@cp helpers/bitmask-root build/${PROVIDER}/snap/local/pre/
	@cd build/${PROVIDER}/snap && python3 generate.py
	@rm build/${PROVIDER}/snap/data.json build/${PROVIDER}/snap/snapcraft-template.yaml
	@mkdir -p build/${PROVIDER}/snap/gui && cp branding/assets/default/icon.svg build/${PROVIDER}/snap/gui/icon.svg
	@cp branding/assets/default/icon.png build/${PROVIDER}/snap/gui/${PROVIDER}-vpn.png

gen_pkg_deb:
	@cp -r ${TEMPLATES}/debian build/${PROVIDER}
	@VERSION=${VERSION} PROVIDER_CONFIG=${PROVIDER_CONFIG} ${SCRIPTS}/generate-debian.py build/${PROVIDER}/debian/data.json
	@mkdir -p build/${PROVIDER}/debian/icons/scalable && cp branding/assets/default/icon.svg build/${PROVIDER}/debian/icons/scalable/icon.svg
	@cd build/${PROVIDER}/debian && python3 generate.py
	@cd build/${PROVIDER}/debian && rm app.desktop-template changelog-template rules-template control-template generate.py data.json && chmod +x rules

prepare_done:
	@echo
	@echo 'Done. You can do "make build" now.'

#########################################################################
# packaging action
#########################################################################

packages: package_deb package_snap package_osx package_win

package_snap:
	@make -C build/${PROVIDER} pkg_snap

package_deb:
	@make -C build/${PROVIDER} pkg_deb

package_win:
	@make -C build/${PROVIDER} pkg_win

package_osx:
	@make -C build/${PROVIDER} pkg_osx



#########################################################################
# icons & locales
#########################################################################

icon:
	@make -C icon


LANGS ?= $(foreach path,$(wildcard locales/*),$(patsubst locales/%,%,$(path)))
empty :=
space := $(empty) $(empty)
lang_list := $(subst $(space),,$(foreach lang,$(LANGS),$(lang),))

locales: $(foreach lang,$(LANGS),get_$(lang)) cmd/bitmask-vpn/catalog.go

generate_locales:
	@gotext update -lang=$(lang_list) ./pkg/systray ./pkg/bitmask
	@make -C tools/transifex

locales/%/out.gotext.json: pkg/systray/systray.go pkg/systray/notificator.go pkg/bitmask/standalone.go pkg/bitmask/bitmaskd.go
	@gotext update -lang=$* ./pkg/systray ./pkg/bitmask

cmd/bitmask-vpn/catalog.go: $(foreach lang,$(LANGS),locales/$(lang)/messages.gotext.json)
	@gotext update -lang=$(lang_list) -out cmd/bitmask-vpn/catalog.go ./pkg/systray ./pkg/bitmask

get_%: locales/%/out.gotext.json
	@make -C tools/transifex build
	@curl -L -X GET --user "api:${API_TOKEN}" "https://www.transifex.com/api/2/project/bitmask/resource/RiseupVPN/translation/${subst -,_,$*}/?file" | tools/transifex/transifex t2g locales/$*/

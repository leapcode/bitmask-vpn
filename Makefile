.PHONY: all get build build_bitmaskd icon locales generate_locales clean

TAGS ?= gtk_3_18

PROVIDER ?= $(shell grep ^'provider =' branding/config/vendor.conf | cut -d '=' -f 2 | tr -d "[:space:]")
PROVIDER_CONFIG ?= branding/config/vendor.conf
DEFAULT_PROVIDER = branding/assets/default/
VERSION ?= $(shell git describe)

# detect OS, we use it for dependencies
UNAME = `uname`

all: icon locales get build

depends:
	-@make depends$(UNAME)
	go get -u golang.org/x/text/cmd/gotext github.com/cratonica/2goarray

dependsLinux:
	sudo apt install libgtk-3-dev libappindicator3-dev golang pkg-config cmake

dependsDarwin:
	# TODO - bootstrap homebrew if not there
	brew install python3 golang make pkg-config upx
	brew install --default-names gnu-sed

dependsCygwin:
	choco install -y golang python nssm nsis wget 7zip

get:
	go get -tags $(TAGS) ./...
	go get -tags "$(TAGS) bitmaskd" ./...

generate:
	go generate cmd/bitmask-vpn/main.go

relink_default:
ifneq (,$(wildcard ${DEFAULT_PROVIDER}))
	cd branding/assets && unlink default
endif
	cd branding/assets && ln -s ${PROVIDER} default

prepare: generate relink_default
	mkdir -p build/${PROVIDER}/bin/
	cp branding/templates/makefile/Makefile build/${PROVIDER}/Makefile
	VERSION=${VERSION} PROVIDER_CONFIG=${PROVIDER_CONFIG} branding/scripts/generate-vendor-make.py build/${PROVIDER}/vendor.mk
	branding/scripts/check-ca-crt.py ${PROVIDER} ${PROVIDER_CONFIG}
	# FIXME trouble in win - better get into repo
	#-@make icon

gen_pkg_win:
	mkdir -p build/${PROVIDER}/windows/
	cp -r branding/templates/windows build/${PROVIDER}
	VERSION=${VERSION} PROVIDER_CONFIG=${PROVIDER_CONFIG} branding/scripts/generate-win.py build/${PROVIDER}/windows/data.json
	cd build/${PROVIDER}/windows && python3 generate.py
	# TODO create/copy build/PROVIDER/assets/
	# TODO create/copy build/PROVIDER/staging/

gen_pkg_osx:
	mkdir -p build/${PROVIDER}/osx/scripts
	mkdir -p build/${PROVIDER}/staging
ifeq (,$(wildcard build/${PROVIDER}/assets))
	ln -s ../../branding/assets/default build/${PROVIDER}/assets
endif
ifeq (,$(wildcard build/${PROVIDER}/staging/openvpn-osx))
	curl -L https://downloads.leap.se/thirdparty/osx/openvpn/openvpn -o build/${PROVIDER}/staging/openvpn-osx
endif
	cp -r branding/templates/osx build/${PROVIDER}
	VERSION=${VERSION} PROVIDER_CONFIG=${PROVIDER_CONFIG} branding/scripts/generate-osx.py build/${PROVIDER}/osx/data.json
	cd build/${PROVIDER}/osx && python3 generate.py
	cd build/${PROVIDER}/osx/scripts && chmod +x preinstall postinstall

gen_pkg_snap:
	cp -r branding/templates/snap build/${PROVIDER}
	VERSION=${VERSION} PROVIDER_CONFIG=${PROVIDER_CONFIG} branding/scripts/generate-snap.py build/${PROVIDER}/snap/data.json
	cd build/${PROVIDER}/snap && python3 generate.py
	rm build/${PROVIDER}/snap/data.json build/${PROVIDER}/snap/snapcraft-template.yaml
	mkdir -p build/${PROVIDER}/snap/gui && cp branding/assets/default/icon.svg build/${PROVIDER}/snap/gui/icon.svg
	# TODO missing hooks

gen_pkg_deb:
	echo "debian..."

gen_pkg_all: prepare gen_pkg_win gen_pkg_osx gen_pkg_snap gen_pkg_deb

build: $(foreach path,$(wildcard cmd/*),build_$(patsubst cmd/%,%,$(path)))

build_%:
	go build -tags $(TAGS) -ldflags "-X main.version=`git describe --tags`" -o $* ./cmd/$*
	# FIXME does not work in osx
	# strip $*
	mkdir -p build/bin
	mv $* build/bin/
	mkdir -p build/${PROVIDER}/staging
	ln -s ../../bin/$* build/${PROVIDER}/staging/$*

test:
	go test -tags "integration $(TAGS)" ./...

build_bitmaskd:
	go build -tags "$(TAGS) bitmaskd" -ldflags "-X main.version=`git describe --tags`" ./cmd/*

build_win:
	powershell -Command '$$version=git describe --tags; go build -ldflags "-H windowsgui -X main.version=$$version" ./cmd/*'

clean:
	make -C icon clean
	rm -f build/${PROVIDER}/bin/bitmask-*
	unlink branding/assets/default

icon:
	make -C icon


LANGS ?= $(foreach path,$(wildcard locales/*),$(patsubst locales/%,%,$(path)))
empty :=
space := $(empty) $(empty)
lang_list := $(subst $(space),,$(foreach lang,$(LANGS),$(lang),))

locales: $(foreach lang,$(LANGS),get_$(lang)) cmd/bitmask-vpn/catalog.go

generate_locales:
	gotext update -lang=$(lang_list) ./pkg/systray ./pkg/bitmask
	make -C tools/transifex

locales/%/out.gotext.json: pkg/systray/systray.go pkg/systray/notificator.go pkg/bitmask/standalone.go pkg/bitmask/bitmaskd.go
	gotext update -lang=$* ./pkg/systray ./pkg/bitmask

cmd/bitmask-vpn/catalog.go: $(foreach lang,$(LANGS),locales/$(lang)/messages.gotext.json)
	gotext update -lang=$(lang_list) -out cmd/bitmask-vpn/catalog.go ./pkg/systray ./pkg/bitmask

get_%: locales/%/out.gotext.json
	make -C tools/transifex build
	curl -L -X GET --user "api:${API_TOKEN}" "https://www.transifex.com/api/2/project/bitmask/resource/RiseupVPN/translation/${subst -,_,$*}/?file" | tools/transifex/transifex t2g locales/$*/

.PHONY: all get build build_bitmaskd icon locales generate_locales clean

TAGS ?= gtk_3_18

PROVIDER ?= $(shell grep ^'provider =' branding/config/vendor.conf | cut -d '=' -f 2 | tr -d "[:space:]")
PROVIDER_CONFIG ?= branding/config/vendor.conf
DEFAULT_PROVIDER = branding/assets/default/
VERSION ?= $(shell git describe)

all: icon locales get build

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
	branding/scripts/check-ca-crt.py ${PROVIDER} ${PROVIDER_CONFIG}

prepare_win:
	mkdir -p build/${PROVIDER}/windows/
	cp -r branding/templates/windows build/${PROVIDER}
	VERSION=${VERSION} PROVIDER_CONFIG=${PROVIDER_CONFIG} branding/scripts/generate-win.py build/${PROVIDER}/windows/data.json
	cd build/${PROVIDER}/windows && python3 generate.py
	# TODO create build/PROVIDER/assets/
	# TODO create build/PROVIDER/staging/

prepare_osx:
	echo "osx..."

prepare_snap:
	echo "snap..."

prepare_debian:
	echo "debian..."

prepare_all: prepare prepare_win prepare_osx prepare_snap

build: $(foreach path,$(wildcard cmd/*),build_$(patsubst cmd/%,%,$(path)))

build_%:
	go build -tags $(TAGS) -ldflags "-X main.version=`git describe --tags`" -o $* ./cmd/$*
	strip $*
	mkdir -p build/bin
	mv $* build/bin/

test:
	go test -tags "integration $(TAGS)" ./...

build_bitmaskd:
	go build -tags "$(TAGS) bitmaskd" -ldflags "-X main.version=`git describe --tags`" ./cmd/*

build_win:
	powershell -Command '$$version=git describe --tags; go build -ldflags "-H windowsgui -X main.version=$$version" ./cmd/*'

clean:
	make -C icon clean
	rm build/bitmask-vpn

icon:
	make -C icon

get_deps:
	sudo apt install libgtk-3-dev libappindicator3-dev golang pkg-config


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

.PHONY: all get build build_bitmaskd icon locales generate_locales clean

TAGS ?= gtk_3_18

PROVIDER ?= $(shell grep ^'provider =' branding/config/vendor.conf | cut -d '=' -f 2 | tr -d "[:space:]")
DEFAULT_PROVIDER = branding/assets/default/

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
	branding/scripts/check-ca-crt.py ${PROVIDER} branding/config/vendor.conf

build: $(foreach path,$(wildcard cmd/*),build_$(patsubst cmd/%,%,$(path)))

build_%:
	go build -tags $(TAGS) -ldflags "-X main.version=`git describe --tags`" -o $* ./cmd/$*

test:
	go test -tags "integration $(TAGS)" ./...

build_bitmaskd:
	go build -tags "$(TAGS) bitmaskd" -ldflags "-X main.version=`git describe --tags`" ./cmd/*

build_win:
	powershell -Command '$$version=git describe --tags; go build -ldflags "-H windowsgui -X main.version=$$version" ./cmd/*'

clean:
	make -C icon clean
	rm bitmask-vpn

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

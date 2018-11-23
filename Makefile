.PHONY: all get build build_bitmaskd icon locales generate_locales clean

TAGS ?= gtk_3_18

all: icon locales get build

get:
	go get -tags $(TAGS) ./...
	go get -tags "$(TAGS) bitmaskd" ./...

build:
	go build -tags $(TAGS) -ldflags "-X main.version=`git describe --tags`"

test:
	go test -tags "integration $(TAGS)" ./...

build_bitmaskd:
	go build -tags "$(TAGS) bitmaskd" -ldflags "-X main.version=`git describe --tags`"

build_win:
	powershell -Command '$$version=git describe --tags; go build -ldflags "-H windowsgui -X main.version=$$version"'

clean:
	make -C icon clean
	rm bitmask-systray

icon:
	make -C icon

get_deps:
	 sudo apt install libgtk-3-dev libappindicator3-dev golang pkg-config


LANGS ?= $(foreach path,$(wildcard locales/*),$(patsubst locales/%,%,$(path)))
empty :=
space := $(empty) $(empty)
lang_list := $(subst $(space),,$(foreach lang,$(LANGS),$(lang),))

locales: $(foreach lang,$(LANGS),get_$(lang)) catalog.go

generate_locales: $(foreach lang,$(LANGS),locales/$(lang)/out.gotext.json)
	make -C transifex

locales/%/out.gotext.json: systray.go notificator.go
	gotext update -lang=$*

catalog.go: $(foreach lang,$(LANGS),locales/$(lang)/messages.gotext.json)
	gotext update -lang=$(lang_list) -out catalog.go

get_%:
	make -C transifex build
	curl -L -X GET --user "api:${API_TOKEN}" "https://www.transifex.com/api/2/project/bitmask/resource/RiseupVPN/translation/${subst -,_,$*}/?file" | transifex/transifex t2g locales/$*/

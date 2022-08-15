OBFS4_ENDPOINT ?= 
OBFS4_CERT ?= 

certs:
	curl -k https://black.riseup.net/ca.crt > /tmp/ca.crt
	curl -k https://api.black.riseup.net/3/cert > /tmp/cert.pem

build-client:
	go get ./...
	go build -o obfsvpn-client ./cmd/client/

run-client:
	./obfsvpn-client -c ${OBFS4_CERT}

run-client-kcp:
	KCP=1 ./obfsvpn-client -c ${OBFS4_CERT}

run-openvpn:
	./scripts/run-openvpn-client.sh

check:
	curl https://wtfismyip.com/json

.PHONY: lint
lint: go-fmt go-vet go-sec

go-fmt:
	gofmt -s -l .

go-vet:
	go vet ./...

go-sec:
	gosec ./...

GOMOBILE = $(shell go env GOPATH)/bin/gomobile

build-android:
	go get -u golang.org/x/mobile/cmd/gomobile
	$(GOMOBILE) init	
	$(GOMOBILE) bind -x -target android -o mobile/android/obfsvpn.aar ./client/

test:
	go test ./...


OBFS4_ENDPOINT ?= 
OBFS4_CERT ?= 
TAG ?= 0.1
OBFSVPN_IMAGE_NAME ?= leap/obfsvpn-test-server
OBFSVPN_CLIENT_IMAGE_NAME ?= leap/obfsvpn-test-client

certs:
	#curl -k https://black.riseup.net/ca.crt > /tmp/ca.crt
	curl -k https://api.black.riseup.net/3/cert > /tmp/cert.pem

build-client:
	go get ./...
	CGO_ENABLED=0 go build -o obfsvpn-client ./cmd/client/

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

check-yawning-obfs4:
	./scripts/check-obfs4-ver.sh

obfsvpn-server-container:
	docker build -t ${OBFSVPN_IMAGE_NAME}:${TAG} -f images/obfsvpn/Dockerfile \
		. --no-cache

obfsvpn-client-container:
	docker build -t ${OBFSVPN_CLIENT_IMAGE_NAME}:${TAG} \
		-f images/obfsvpn-client/Dockerfile \
		. --no-cache

integration: obfsvpn-server-container obfsvpn-client-container
	docker network create test-net
	docker run -d --rm --privileged -p 4430:4430/tcp \
		-v obfsvpn_data:/opt/Dockovpn_data:Z -e OBFS4_HOST=0.0.0.0 \
		--name obfsvpn-server --net test-net ${OBFSVPN_IMAGE_NAME}:${TAG}
	docker run --rm --privileged -v obfsvpn_data:/vpn:Z --net test-net \
		--name obfsvpn-client ${OBFSVPN_CLIENT_IMAGE_NAME}:${TAG}


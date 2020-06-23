.PHONY: build test clean prepare update docker

GO = CGO_ENABLED=0 GO111MODULE=on go

MICROSERVICES=cmd/device-coap-go

.PHONY: $(MICROSERVICES)

DOCKERS=docker_device_coap_go
.PHONY: $(DOCKERS)

VERSION=$(shell cat ./VERSION)
GIT_SHA=$(shell git rev-parse HEAD)
GOFLAGS=-ldflags "-X github.com/edgexfoundry/device-coap-go.Version=$(VERSION)"

build: $(MICROSERVICES)
	$(GO) build ./...

cmd/device-coap-go:
	$(GO) build $(GOFLAGS) -o $@ ./cmd

test:
	$(GO) test ./... -coverprofile=coverage.out
	$(GO) vet ./...
	gofmt -l .
	[ "`gofmt -l .`" = "" ]
	./bin/test-attribution-txt.sh
	./bin/test-go-mod-tidy.sh

clean:
	rm -f $(MICROSERVICES)

docker: $(DOCKERS)

docker_device_coap_go:
	docker build \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-device-coap-go:$(GIT_SHA) \
		-t edgexfoundry/docker-device-coap-go:$(VERSION)-dev \
		.

-include environ.inc
.PHONY: deps dev build install image release test clean

export CGO_ENABLED=0
VERSION=$(shell git describe --abbrev=0 --tags 2>/dev/null || echo "$VERSION")
COMMIT=$(shell git rev-parse --short HEAD || echo "$COMMIT")
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
GOCMD=go

DESTDIR=/usr/local/bin

ifeq ($(BRANCH), main)
IMAGE := prologic/todo
TAG := latest
else
IMAGE := prologic/todo
TAG := dev
endif

all: preflight build

preflight:
	@./preflight.sh

deps:

dev : DEBUG=1
dev : build
	@./todo

build:
	@$(GOCMD) build $(FLAGS) -tags "netgo static_build" -installsuffix netgo \
		-ldflags "-w \
		-X $(shell go list).Version=$(VERSION) \
		-X $(shell go list).Commit=$(COMMIT)" \
		.

install: build
	@install -D -m 755 todo $(DESTDIR)/todo

ifeq ($(PUBLISH), 1)
image:
	@docker build --build-arg VERSION="$(VERSION)" --build-arg COMMIT="$(COMMIT)" -t $(IMAGE):$(TAG) .
	@docker push $(IMAGE):$(TAG)
else
image:
	@docker build --build-arg VERSION="$(VERSION)" --build-arg COMMIT="$(COMMIT)" -t $(IMAGE):$(TAG) .
endif

release:
	@./tools/release.sh

fmt:
	@$(GOCMD) fmt ./...

test:
	@CGO_ENABLED=1 $(GOCMD) test -v -cover -race ./...

coverage:
	@CGO_ENABLED=1 $(GOCMD) test -v -cover -race -cover -coverprofile=coverage.out  ./...
	@$(GOCMD) tool cover -html=coverage.out

clean:
	@git clean -f -d -X

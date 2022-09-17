VERSION=0.1.0

LDFLAGS="-X 'uptime.Version=${VERSION}' -X 'uptime.CommitHash=$(shell git rev-parse --short HEAD)'"

.DEFAULT_GOAL := install

help: ## self documenting help output
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
.PHONY: help

init: ## configure go and git
	@echo "nothing required yet"
.PHONY: init

fmt: ## format code
	go fmt ./...
.PHONY: fmt

vet: fmt ## vet code
	go vet ./...
.PHONY: vet

test: vet ## test code
	go test -v ./...
.PHONY: test

UNAME := $(shell uname -m)

install: test ## install provider
	go build -o ~/.terraform.d/plugins/integral.com.au/local/uptime/${VERSION}/darwin_${UNAME}/terraform-provider-uptime \
		-ldflags ${LDFLAGS}
.PHONY: install

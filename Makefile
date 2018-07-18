GO   := GO15VENDOREXPERIMENT=1 go
pkgs  = $(shell $(GO) list ./... | grep -v /vendor/)

all: format test build

test:
	@echo ">> running tests"
	@$(GO) test -race $(pkgs)

format:
	@echo ">> formatting code"
	@$(GO) fmt $(pkgs)

vet:
	@echo ">> vetting code"
	@$(GO) vet $(pkgs)

build:
	@echo ">> building binary"
	@./scripts/build.sh

docker:
	@docker build -t keti:$(shell git rev-parse --short HEAD) .

release:
	@echo ">> building binaries"
	@./scripts/release

.PHONY: all format build test vet docker assets

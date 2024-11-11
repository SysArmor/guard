PROJECT_PKG = $(shell pwd)
CLI_EXE     = guard
CLI_PKG     = $(PROJECT_PKG)
GIT_COMMIT  = $(shell git rev-parse HEAD)
GIT_TAG     = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
GIT_DIRTY   = $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "clean")
GO_EXE      = go
DEB_PKGS    = $(PROJECT_PKG)/client/debian
DEB_BIN	 	= $(DEB_PKGS)/usr/bin
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

LDFLAGS = -w

.PHONY: help
help:  
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[%\/0-9A-Za-z_-]+:.*?##/ { printf "  \033[36m%-45s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: generate-cert
generate-cert: ## Generate ca and cert files
	ssh-keygen -C CA -f ca -b 4096

.PHONY: build-swagger
build-swagger: ## Build docs from swagger
	swag init -g ./server/docs/swag/main.go -o ./server/docs/swag/docs

.PHONY: run-docs
run-docs: build-swagger ## Run docs server
	go run ./server/docs/swag/.

.PHONY: build-client
build-client: ## Build client
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(CLI_EXE)-client $(CLI_PKG)/client/.

.PHONY: build-server
build-server: ## Build server
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(CLI_EXE)-server $(CLI_PKG)/server/.

.PHONY: build-deb
build-deb: build-client ## Build debian package
	chmod +x $(CLI_EXE)-client
	mkdir -p $(DEB_BIN)
	mv $(CLI_EXE)-client $(DEB_BIN)/$(CLI_EXE)	
	dpkg-deb --build $(PROJECT_PKG)/client/debian $(CLI_EXE).deb
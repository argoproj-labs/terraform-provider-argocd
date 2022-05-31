TEST?=./...
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
PKG_NAME=pass

BINARY=terraform-provider-argocd
VERSION = $(shell git describe --always)

default: build-all

build-all: linux windows darwin freebsd

install: fmtcheck
	go install

linux: fmtcheck
	@mkdir -p bin/
	GOOS=linux GOARCH=amd64 go build -v -o bin/$(BINARY)_$(VERSION)_linux_amd64
	GOOS=linux GOARCH=386 go build -v -o bin/$(BINARY)_$(VERSION)_linux_x86

windows: fmtcheck
	@mkdir -p bin/
	GOOS=windows GOARCH=amd64 go build -v -o bin/$(BINARY)_$(VERSION)_windows_amd64
	GOOS=windows GOARCH=386 go build -v -o bin/$(BINARY)_$(VERSION)_windows_x86

darwin: fmtcheck
	@mkdir -p bin/
	GOOS=darwin GOARCH=amd64 go build -v -o bin/$(BINARY)_$(VERSION)_darwin_amd64
	GOOS=darwin GOARCH=386 go build -v -o bin/$(BINARY)_$(VERSION)_darwin_x86

freebsd: fmtcheck
	@mkdir -p bin/
	GOOS=freebsd GOARCH=amd64 go build -v -o bin/$(BINARY)_$(VERSION)_freebsd_amd64
	GOOS=freebsd GOARCH=386 go build -v -o bin/$(BINARY)_$(VERSION)_freebsd_x86

release: clean linux windows darwin freebsd
	for f in $(shell ls bin/); do zip bin/$${f}.zip bin/$${f}; done

clean:
	git clean -fXd -e \!vendor -e \!vendor/**/*

test: fmtcheck
	go test $(TEST) -timeout=30s -parallel=4

testacc_prepare_env:
	sh scripts/testacc_prepare_env.sh

testacc:
	sh scripts/testacc.sh

testacc_clean_env:
	kind delete cluster --name argocd

fmt:
	@echo "==> Fixing source code with gofmt..."
	gofmt -s -w ./$(PKG_NAME)

# Currently required by tf-deploy compile
fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

lint:
	@echo "==> Checking source code against linters..."
	@GOGC=30 golangci-lint run ./$(PKG_NAME)

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

vendor:
	go mod tidy
	go mod vendor

vet:
	go vet $<

.PHONY: build test testacc_prepare_env testacc testacc_clean_env fmt fmtcheck lint test-compile vendor

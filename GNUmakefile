default: build

# To see which other env vars are set, please refer to internal/testhelpers/suite.go
TEST_FILTER?=

# Versions of Argo CD and K3s the acceptance tests run against. Override them to test against another version, e.g.
# `make testacc ARGOCD_VERSION=v3.4.8`. The IDE run configs (.vscode/launch.json, .run/) duplicate these values, keep
# them in sync.
ARGOCD_VERSION?=v3.5.6
K3S_VERSION?=v1.34.3-k3s3

export

build:
	go build -v -o bin/ ./...

install: build
	go install -v ./...

# See https://golangci-lint.run/
lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=4 -run="$(TEST_FILTER)" ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout 20m -run="$(TEST_FILTER)" ./...

clean:
	git clean -fXd -e \!vendor -e \!vendor/**/* -e \!.vscode

.PHONY: build install lint generate fmt test testacc clean

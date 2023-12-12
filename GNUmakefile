default: build

ARGOCD_INSECURE?=true
ARGOCD_SERVER?=127.0.0.1:8080
ARGOCD_AUTH_USERNAME?=admin
ARGOCD_AUTH_PASSWORD?=acceptancetesting
ARGOCD_VERSION?=v2.9.3

export

build:
	go build -v ./...

install: build
	go install -v ./...

# See https://golangci-lint.run/
lint:
	golangci-lint run

generate:
	go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=4 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout 20m ./...

testacc_clean_env:
	kind delete cluster --name argocd

testacc_prepare_env:
	sh scripts/testacc_prepare_env.sh
	
clean:
	git clean -fXd -e \!vendor -e \!vendor/**/* -e \!.vscode

.PHONY: build install lint generate fmt test testacc testacc_clean_env testacc_prepare_env clean
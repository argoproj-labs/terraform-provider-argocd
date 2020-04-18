TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)

default: build

fmt:
	gofmt -w $(GOFMT_FILES)

build: fmt
	go install

test: fmt
	go test -i $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc_prepare_env:
	sh scripts/testacc_prepare_env.sh

testacc:
	sh scripts/testacc.sh

testacc_clean_env:
	kind delete cluster --name argocd

.PHONY: build test testacc_prepare_env testacc testacc_clean_env fmt

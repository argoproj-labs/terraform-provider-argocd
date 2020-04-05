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
	kind create cluster --name argocd && \
	kubectl create namespace argocd && \
	kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/v1.5.0/manifests/install.yaml && \
	kubectl patch -n argocd secret/argocd-secret --type merge -p '{"stringData":{"admin.password":"$$2a$10$Oc4i/CTPPCrbbeAIrwgmzeCg.wzEtCZd2HQz5gZOnPNlBekm.FVta"}}' && \
	kubectl apply -n argocd -f manifests/argocd-project.yml && \
	kubectl wait --for=condition=available --timeout=300s deployment/argocd-server -n argocd && \
	export KPF_PID=$(kubectl port-forward -n argocd service/argocd-server --address 127.0.0.1 8080:443)

testacc:
	ARGOCD_INSECURE=true ARGOCD_SERVER=localhost:8080 \
	ARGOCD_AUTH_USERNAME=admin ARGOCD_AUTH_PASSWORD=acceptancetesting \
	TF_ACC=1 go test $(TEST) -v -timeout 5m

testacc_clean_env:
	kind delete cluster --name argocd

.PHONY: build test testacc_prepare_env testacc testacc_clean_env fmt

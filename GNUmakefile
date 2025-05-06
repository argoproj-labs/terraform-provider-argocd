default: build

ARGOCD_INSECURE?=true
ARGOCD_SERVER?=127.0.0.1:8080
ARGOCD_AUTH_USERNAME?=admin
ARGOCD_AUTH_PASSWORD?=acceptancetesting
ARGOCD_VERSION?=v3.0.0

export

build:
	go build -v ./...

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
	go test -v -cover -timeout=120s -parallel=4 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout 20m ./...

testacc_clean_env:
	kind delete cluster --name argocd

testacc_prepare_env:
	echo "\n--- Clearing current kube context\n"
	kubectl config unset current-context

	echo "\n--- Kustomize sanity checks\n"
	kustomize version || exit 1

	echo "\n--- Create Kind cluster\n"
	kind create cluster --config kind-config.yml 

	echo "\n--- Kind sanity checks\n"
	kubectl get nodes -o wide
	kubectl get pods --all-namespaces -o wide
	kubectl get services --all-namespaces -o wide

	echo "\n--- Fetch ArgoCD installation manifests\n"
	curl https://raw.githubusercontent.com/argoproj/argo-cd/${ARGOCD_VERSION}/manifests/install.yaml > manifests/install/argocd.yml


	echo "\n--- Install ArgoCD ${ARGOCD_VERSION}\n"
	kustomize build manifests/install | kubectl apply -f -

	echo "\n--- Wait until CRDs are established\n"
	kubectl wait --for=condition=Established crd/applications.argoproj.io --timeout=60s
	kubectl wait --for=condition=Established crd/applicationsets.argoproj.io --timeout=60s
	kubectl wait --for=condition=Established crd/appprojects.argoproj.io --timeout=60s

	echo "\n--- Install ArgoCD test data\n"
	kubectl apply -f manifests/testdata/

	echo "\n--- Wait for ArgoCD components to be ready...\n"
	kubectl wait --for=condition=available --timeout=600s deployment/argocd-server -n argocd
	kubectl wait --for=condition=available --timeout=30s deployment/argocd-repo-server -n argocd
	kubectl wait --for=condition=available --timeout=30s deployment/argocd-dex-server -n argocd
	kubectl wait --for=condition=available --timeout=30s deployment/argocd-redis -n argocd
	
clean:
	git clean -fXd -e \!vendor -e \!vendor/**/* -e \!.vscode

.PHONY: build install lint generate fmt test testacc testacc_clean_env testacc_prepare_env clean

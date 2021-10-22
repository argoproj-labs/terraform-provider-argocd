#!/bin/sh
# shellcheck disable=SC2016,SC2028

export PATH=$PATH:.

echo "\n--- Kustomize sanity checks\n"
kustomize version || exit 1

echo "\n--- Create Kind cluster\n"
kind create cluster --name argocd --config scripts/kind-config.yml --image kindest/node:${ARGOCD_KUBERNETES_VERSION:-v1.19.7}

echo "\n--- Kind sanity checks\n"
kubectl get nodes -o wide
kubectl get pods --all-namespaces -o wide
kubectl get services --all-namespaces -o wide

if [[ -z "${ARGOCD_CI}" ]]; then
  echo "\n--- Load already available container images from local registry into Kind (local development only)\n"
  kind load docker-image redis:6.2.4-alpine --name argocd
  kind load docker-image ghcr.io/dexidp/dex:v2.27.0 --name argocd
  kind load docker-image alpine:3 --name argocd
  kind load docker-image quay.io/argoproj/argocd:${ARGOCD_VERSION:-v1.8.7} --name argocd
fi

echo "\n--- Install ArgoCD ${ARGOCD_VERSION:-v1.8.7}\n"
curl https://raw.githubusercontent.com/argoproj/argo-cd/${ARGOCD_VERSION:-v1.8.7}/manifests/install.yaml > manifests/install/argocd.yml &&
kustomize build manifests/install | kubectl apply -f - &&
kubectl apply -f manifests/testdata/ &&

echo "\n--- Wait for ArgoCD components to be ready...\n"
kubectl wait --for=condition=available --timeout=600s deployment/argocd-server -n argocd
kubectl wait --for=condition=available --timeout=30s deployment/argocd-repo-server -n argocd
kubectl wait --for=condition=available --timeout=30s deployment/argocd-dex-server -n argocd
kubectl wait --for=condition=available --timeout=30s deployment/argocd-redis -n argocd

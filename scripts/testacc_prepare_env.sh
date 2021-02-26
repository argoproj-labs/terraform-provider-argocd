#!/bin/sh
# shellcheck disable=SC2016,SC2028

export PATH=$PATH:.

echo '--- Kustomize sanity checks'
kustomize version || exit 1

echo '--- Create Kind cluster\n\n'
kind create cluster --name argocd --config scripts/kind-config.yml --image kindest/node:${ARGOCD_KUBERNETES_VERSION:-v1.19.7}

echo '--- Kind sanity checks\n\n'
kubectl get nodes -o wide
kubectl get pods --all-namespaces -o wide
kubectl get services --all-namespaces -o wide

echo '--- Load already available container images from local registry into Kind (local development only)'
kind load docker-image redis:5.0.10-alpine --name argocd
kind load docker-image ghcr.io/dexidp/dex:v2.27.0 --name argocd
kind load docker-image banzaicloud/vault-operator:1.3.3 --name argocd
kind load docker-image argoproj/argocd:${ARGOCD_VERSION:-v1.8.3} --name argocd

echo '--- Install ArgoCD ${ARGOCD_VERSION:-v1.8.3}\n\n'
kustomize build manifests/install | kubectl apply -f - &&
kubectl apply -f manifests/testdata/ &&

echo '--- Wait for ArgoCD components to be ready...'
kubectl wait --for=condition=available --timeout=600s deployment/argocd-server -n argocd
kubectl wait --for=condition=available --timeout=30s deployment/argocd-repo-server -n argocd
kubectl wait --for=condition=available --timeout=30s deployment/argocd-dex-server -n argocd
kubectl wait --for=condition=available --timeout=30s deployment/argocd-redis -n argocd

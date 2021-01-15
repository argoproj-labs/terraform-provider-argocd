#!/bin/sh
# shellcheck disable=SC2016,SC2028

export PATH=$PATH:.

echo '--- Kustomize sanity checks'
kustomize version || exit 1

echo '--- Create Kind cluster\n\n'
kind create cluster --name argocd --config scripts/kind-config.yml --image kindest/node:${ARGOCD_KUBERNETES_VERSION:-v1.18.8}

echo '--- Kind sanity checks\n\n'
kubectl get nodes -o wide
kubectl get pods --all-namespaces -o wide
kubectl get services --all-namespaces -o wide

echo '--- Load already available container images from local registry into Kind'
kind load docker-image redis:5.0.3 --name argocd
kind load docker-image quay.io/dexidp/dex:v2.22.0 --name argocd
kind load docker-image argoproj/argocd:${ARGOCD_VERSION:-v1.6.1} --name argocd

echo '--- Install ArgoCD ${ARGOCD_VERSION:-v1.6.1}\n\n'
kustomize build manifests/install | kubectl apply -f - &&
kubectl apply -f manifests/testdata/ &&

echo '--- Wait for ArgoCD components to be ready...'
kubectl wait --for=condition=available --timeout=600s deployment/argocd-server -n argocd
kubectl wait --for=condition=available --timeout=30s deployment/argocd-repo-server -n argocd
kubectl wait --for=condition=available --timeout=30s deployment/argocd-application-controller -n argocd
kubectl wait --for=condition=available --timeout=30s deployment/argocd-dex-server -n argocd
kubectl wait --for=condition=available --timeout=30s deployment/argocd-redis -n argocd

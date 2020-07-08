#!/bin/sh
# shellcheck disable=SC2016,SC2028

export PATH=$PATH:.

echo '--- Kustomize sanity checks'
kustomize version || exit 1

echo '--- Create Kind cluster\n\n'
kind create cluster --name argocd

echo '--- Kind sanity checks\n\n'
kubectl get nodes -o wide
kubectl get pods --all-namespaces -o wide
kubectl get services --all-namespaces -o wide

echo '--- Install ArgoCD ${ARGOCD_VERSION:-v1.6.1}\n\n'
kustomize build manifests/install | kubectl apply -f - &&
kubectl apply -f manifests/testdata/ &&

echo '--- Wait for ArgoCD server to be ready...'
kubectl wait --for=condition=available --timeout=600s deployment/argocd-server -n argocd

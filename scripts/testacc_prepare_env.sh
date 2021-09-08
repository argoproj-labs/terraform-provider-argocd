#!/bin/sh
# shellcheck disable=SC2016,SC2028

export PATH=$PATH:.

echo '--- Kustomize sanity checks'
kustomize version || exit 1

echo '--- Create local Kubernetes cluster\n\n'
k3d cluster create argocd -i rancher/k3s:v1.20.10-k3s1 -p "8080:30123@server[0]"

echo '--- Kubernetes sanity checks\n\n'
kubectl get nodes -o wide
kubectl get pods --all-namespaces -o wide
kubectl get services --all-namespaces -o wide

echo '--- Load already available container images from local registry into Kubernetes (local development only)'
k3d image import -t -c argocd redis:6.2.4-alpine
k3d image import -t -c argocd bitnami/redis:6.2.5
k3d image import -t -c argocd ghcr.io/dexidp/dex:v2.27.0
k3d image import -c argocd quay.io/argoproj/argocd:${ARGOCD_VERSION:-v2.1.2}

echo '--- Install ArgoCD ${ARGOCD_VERSION:-v2.1.2}\n\n'
kustomize build manifests/install | kubectl apply -f - &&
kubectl apply -f manifests/testdata/ &&

echo '--- Wait for ArgoCD components to be ready...'
kubectl wait --for=condition=available --timeout=600s deployment/argocd-server -n argocd
kubectl wait --for=condition=available --timeout=30s deployment/argocd-repo-server -n argocd
kubectl wait --for=condition=available --timeout=30s deployment/argocd-dex-server -n argocd
kubectl wait --for=condition=available --timeout=30s deployment/argocd-redis -n argocd
kubectl wait --for=condition=available --timeout=120s deployment/private-git-repository -n argocd
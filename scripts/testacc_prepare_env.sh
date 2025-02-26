#!/bin/sh
# shellcheck disable=SC2016,SC2028
set -e

export PATH=$PATH:.

echo "\n--- Clearing current kube context\n"
kubectl config unset current-context

echo "\n--- Kustomize sanity checks\n"
kustomize version || exit 1

echo "\n--- Create Kind cluster\n"
kind create cluster --config scripts/kind-config.yml 

echo "\n--- Kind sanity checks\n"
kubectl get nodes -o wide
kubectl get pods --all-namespaces -o wide
kubectl get services --all-namespaces -o wide

echo "\n--- Fetch ArgoCD installation manifests\n"
curl https://raw.githubusercontent.com/argoproj/argo-cd/${ARGOCD_VERSION}/manifests/install.yaml > manifests/install/argocd.yml

echo "\n--- Install ArgoCD ${ARGOCD_VERSION}\n"
kustomize build manifests/install | kubectl apply -f - && \
  kubectl apply -f manifests/testdata/

echo "\n--- Wait for ArgoCD components to be ready...\n"
kubectl wait --for=condition=available --timeout=600s deployment/argocd-server -n argocd
kubectl wait --for=condition=available --timeout=30s deployment/argocd-repo-server -n argocd
kubectl wait --for=condition=available --timeout=30s deployment/argocd-dex-server -n argocd
kubectl wait --for=condition=available --timeout=30s deployment/argocd-redis -n argocd

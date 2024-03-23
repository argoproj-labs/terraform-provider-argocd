#!/bin/sh
# shellcheck disable=SC2016,SC2028
set -e

export PATH=$PATH:.

argocd_version=${ARGOCD_VERSION:-v2.8.13}
k8s_version=${ARGOCD_KUBERNETES_VERSION:-v1.27.11}

echo "\n--- Clearing current kube context\n"
kubectl config unset current-context

echo "\n--- Kustomize sanity checks\n"
kustomize version || exit 1

echo "\n--- Create Kind cluster\n"
kind create cluster --name argocd --config scripts/kind-config.yml --image kindest/node:$k8s_version

echo "\n--- Kind sanity checks\n"
kubectl get nodes -o wide
kubectl get pods --all-namespaces -o wide
kubectl get services --all-namespaces -o wide

echo "\n--- Fetch ArgoCD installation manifests\n"
curl https://raw.githubusercontent.com/argoproj/argo-cd/$argocd_version/manifests/install.yaml > manifests/install/argocd.yml

if [ -z "${ARGOCD_CI}" ]; then
  echo "\n--- Load local container images from into Kind (local development only)\n"
  docker pull quay.io/argoproj/argocd:$argocd_version
  kind load docker-image quay.io/argoproj/argocd:$argocd_version --name argocd
  
  dex_version=$(cat manifests/install/argocd.yml| grep "image: ghcr.io/dexidp/dex" | cut -d":" -f3)
  docker pull ghcr.io/dexidp/dex:$dex_version
  kind load docker-image ghcr.io/dexidp/dex:$dex_version --name argocd
  
  redis_version=$(cat manifests/install/argocd.yml| grep "image: redis" | cut -d":" -f3)
  docker pull redis:$redis_version
  kind load docker-image redis:$redis_version --name argocd
fi

echo "\n--- Install ArgoCD $argocd_version\n"
kustomize build manifests/install | kubectl apply -f - && \
  kubectl apply -f manifests/testdata/

echo "\n--- Wait for ArgoCD components to be ready...\n"
kubectl wait --for=condition=available --timeout=600s deployment/argocd-server -n argocd
kubectl wait --for=condition=available --timeout=30s deployment/argocd-repo-server -n argocd
kubectl wait --for=condition=available --timeout=30s deployment/argocd-dex-server -n argocd
kubectl wait --for=condition=available --timeout=30s deployment/argocd-redis -n argocd

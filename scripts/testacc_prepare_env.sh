#!/bin/sh
# shellcheck disable=SC2016,SC2039

printf '--- Create Kind cluster\n\n'
PATH=$PATH:$(go env GOPATH)/bin kind create cluster --name argocd

printf '--- Kind sanity checks\n\n'
kubectl get nodes -o wide
kubectl get pods --all-namespaces -o wide
kubectl get services --all-namespaces -o wide

printf '--- Install ArgoCD\n\n'
kubectl create namespace argocd &&
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/v1.5.0/manifests/install.yaml &&
kubectl patch -n argocd secret/argocd-secret --type merge -p '{"stringData":{"admin.password":"$2a$10$O7VHb/85434QLWAep6.pye/z454DE3R2IWbCIJ7q5V/nTXUdPEBZC"}}' &&
kubectl apply -n argocd -f manifests/argocd-project.yml &&
kubectl wait --for=condition=available --timeout=600s deployment/argocd-server -n argocd

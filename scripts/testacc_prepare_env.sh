#!/bin/sh

# Create Kind cluster
PATH=$PATH:$(go env GOPATH)/bin kind create cluster --name argocd

# Kind sanity checks
kubectl get nodes -o wide
kubectl get pods --all-namespaces -o wide
kubectl get services --all-namespaces -o wide

# Install ArgoCD
kubectl create namespace argocd &&
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/v1.5.0/manifests/install.yaml &&
kubectl patch -n argocd secret/argocd-secret --type merge -p '{"stringData":{"admin.password":"'$(echo -n acceptancetesting|bcrypt-cli)'"}}' &&
kubectl apply -n argocd -f manifests/argocd-project.yml &&
kubectl wait --for=condition=available --timeout=600s deployment/argocd-server -n argocd

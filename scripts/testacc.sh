#!/bin/sh

ARGOCD_INSECURE=true \
ARGOCD_SERVER=127.0.0.1:8080 \
ARGOCD_AUTH_USERNAME=admin \
ARGOCD_AUTH_PASSWORD=acceptancetesting \
ARGOCD_CONTEXT=kind-argocd \
TF_ACC=1 go test $(go list ./... |grep -v 'vendor') -v -timeout 5m

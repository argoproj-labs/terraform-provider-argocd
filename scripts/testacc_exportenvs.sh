#!/bin/sh

set -e

export ARGOCD_INSECURE=true
export ARGOCD_SERVER=127.0.0.1:8080
export ARGOCD_AUTH_USERNAME=admin
export ARGOCD_AUTH_PASSWORD=acceptancetesting
export ARGOCD_CONTEXT=kind-argocd

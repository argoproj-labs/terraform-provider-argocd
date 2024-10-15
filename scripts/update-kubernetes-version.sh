#!/usr/bin/env bash
# Source: 
# https://github.com/kubernetes/kubernetes/issues/79384#issuecomment-521493597
# https://github.com/argoproj/argo-cd/blob/master/hack/update-kubernetes-version.sh
set -euo pipefail

if [ -z "${1:-}" ]; then
  echo "Example usage: ./hack/update-kubernetes-version.sh v1.26.11"
  exit 1
fi
VERSION=${1#"v"}
MODS=($(
  curl -sS https://raw.githubusercontent.com/kubernetes/kubernetes/v${VERSION}/go.mod |
    sed -n 's|.*k8s.io/\(.*\) => ./staging/src/k8s.io/.*|k8s.io/\1|p'
))
for MOD in "${MODS[@]}"; do
  echo "Updating $MOD..." >&2
  V=$(
    go mod download -json "${MOD}@kubernetes-${VERSION}" |
      sed -n 's|.*"Version": "\(.*\)".*|\1|p'
  )
  go mod edit "-replace=${MOD}=${MOD}@${V}"
done
go get "k8s.io/kubernetes@v${VERSION}"
go mod tidy

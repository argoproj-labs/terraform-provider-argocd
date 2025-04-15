package sync

import "sync"

// GPGKeysMutex is used to handle concurrent access to ArgoCD GPG keys which are
// stored in the `argocd-gpg-keys-cm` ConfigMap resource
var GPGKeysMutex = &sync.RWMutex{}
var RepositoryMutex = &sync.RWMutex{}

package sync

import "sync"

// GPGKeysMutex is used to handle concurrent access to ArgoCD GPG keys which are
// stored in the `argocd-gpg-keys-cm` ConfigMap resource
var GPGKeysMutex = &sync.RWMutex{}

// AccountTokensMutex is used to handle concurrent access to ArgoCD account token operations
var AccountTokensMutex = &sync.RWMutex{}

// AccountsMutex is used to handle concurrent access to ArgoCD account operations
// (password updates, account modifications)
var AccountsMutex = &sync.RWMutex{}

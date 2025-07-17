package sync

import "sync"

// GPGKeysMutex is used to handle concurrent access to ArgoCD GPG keys which are
// stored in the `argocd-gpg-keys-cm` ConfigMap resource
var GPGKeysMutex = &sync.RWMutex{}

// RepositoryMutex is used to handle concurrent access to ArgoCD repositories
var RepositoryMutex = &sync.RWMutex{}

// CertificateMutex is used to handle concurrent access to ArgoCD repository certificates
var CertificateMutex = &sync.RWMutex{}

// RepositoryCredentialsMutex is used to handle concurrent access to ArgoCD repository credentials
var RepositoryCredentialsMutex = &sync.RWMutex{}

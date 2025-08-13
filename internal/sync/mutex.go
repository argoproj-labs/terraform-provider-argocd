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

// tokenMutexProjectMap is used to handle concurrent access to ArgoCD project tokens per project
var tokenMutexProjectMap = make(map[string]*sync.RWMutex)

// tokenMutexProjectMapMutex protects access to TokenMutexProjectMap itself
var tokenMutexProjectMapMutex = &sync.Mutex{}

// GetProjectMutex safely gets or creates a mutex for a project
func GetProjectMutex(projectName string) *sync.RWMutex {
	tokenMutexProjectMapMutex.Lock()
	defer tokenMutexProjectMapMutex.Unlock()

	if mutex, exists := tokenMutexProjectMap[projectName]; exists {
		return mutex
	}

	tokenMutexProjectMap[projectName] = &sync.RWMutex{}

	return tokenMutexProjectMap[projectName]
}

package argocd

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/argoproj/argo-cd/pkg/apiclient"
	"github.com/argoproj/argo-cd/pkg/apiclient/project"
	"github.com/argoproj/argo-cd/pkg/apiclient/version"
)

const (
	featureTokenIDs = iota
)

var (
	featureVersionConstraintsMap = map[int]*semver.Version{
		featureTokenIDs: semver.MustParse("1.5.3"),
	}
)

type ServerInterface struct {
	ApiClient            apiclient.Client
	ProjectClient        project.ProjectServiceClient
	ServerVersion        *semver.Version
	ServerVersionMessage *version.VersionMessage
}

// Checks that a specific feature is available for the current ArgoCD server version.
// 'feature' argument must match one of the predefined feature* constants.
func (p ServerInterface) isFeatureSupported(feature int) (bool, error) {
	versionConstraint, ok := featureVersionConstraintsMap[feature]
	if !ok {
		return false, fmt.Errorf("feature constraint is not handled by the provider")
	}
	if i := versionConstraint.Compare(p.ServerVersion); i == 1 {
		return false, nil
	}
	return true, nil
}

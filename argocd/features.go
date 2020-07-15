package argocd

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/argoproj/argo-cd/pkg/apiclient"
	"github.com/argoproj/argo-cd/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/pkg/apiclient/project"
	"github.com/argoproj/argo-cd/pkg/apiclient/repocreds"
	"github.com/argoproj/argo-cd/pkg/apiclient/repository"
	"github.com/argoproj/argo-cd/pkg/apiclient/version"
)

const (
	featureApplicationLevelSyncOptions = iota
	featureRepositoryGet
	featureTokenIDs
)

var (
	featureVersionConstraintsMap = map[int]*semver.Version{
		featureApplicationLevelSyncOptions: semver.MustParse("1.5.0"),
		featureRepositoryGet:               semver.MustParse("1.6.0"),
		featureTokenIDs:                    semver.MustParse("1.5.3"),
	}
)

type ServerInterface struct {
	ApiClient            *apiclient.Client
	ApplicationClient    *application.ApplicationServiceClient
	ProjectClient        *project.ProjectServiceClient
	RepositoryClient     *repository.RepositoryServiceClient
	RepoCredsClient      *repocreds.RepoCredsServiceClient
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

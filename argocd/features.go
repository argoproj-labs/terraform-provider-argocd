package argocd

import (
	"context"
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/argoproj/argo-cd/util/io"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/project"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/repocreds"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/repository"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/version"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	ClusterClient        *cluster.ClusterServiceClient
	ProjectClient        *project.ProjectServiceClient
	RepositoryClient     *repository.RepositoryServiceClient
	RepoCredsClient      *repocreds.RepoCredsServiceClient
	ServerVersion        *semver.Version
	ServerVersionMessage *version.VersionMessage
	ProviderData         *schema.ResourceData
}

func (p *ServerInterface) initClients() error {
	d := p.ProviderData
	apiClient, err := initApiClient(d)
	if err != nil {
		return err
	}
	p.ApiClient = &apiClient

	_, clusterClient, err := apiClient.NewClusterClient()
	if err != nil {
		return err
	}
	p.ClusterClient = &clusterClient

	_, applicationClient, err := apiClient.NewApplicationClient()
	if err != nil {
		return err
	}
	p.ApplicationClient = &applicationClient

	_, projectClient, err := apiClient.NewProjectClient()
	if err != nil {
		return err
	}
	p.ProjectClient = &projectClient

	_, repositoryClient, err := apiClient.NewRepoClient()
	if err != nil {
		return err
	}
	p.RepositoryClient = &repositoryClient

	_, repoCredsClient, err := apiClient.NewRepoCredsClient()
	if err != nil {
		return err
	}
	p.RepoCredsClient = &repoCredsClient

	acCloser, versionClient, err := apiClient.NewVersionClient()
	if err != nil {
		return err
	}
	defer io.Close(acCloser)

	serverVersionMessage, err := versionClient.Version(context.Background(), &empty.Empty{})
	if err != nil {
		return err
	}
	if serverVersionMessage == nil {
		return fmt.Errorf("could not get server version information")
	}
	p.ServerVersionMessage = serverVersionMessage
	serverVersion, err := semver.NewVersion(serverVersionMessage.Version)
	if err != nil {
		return fmt.Errorf("could not parse server semantic version: %s", serverVersionMessage.Version)
	}
	p.ServerVersion = serverVersion

	return nil
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

package argocd

import (
	"context"
	"fmt"
	"sync"

	"github.com/Masterminds/semver"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/account"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/applicationset"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/certificate"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/project"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/repocreds"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/repository"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/session"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/version"
	"github.com/argoproj/argo-cd/v2/util/io"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	featureApplicationLevelSyncOptions = iota
	featureIgnoreDiffJQPathExpressions
	featureRepositoryGet
	featureTokenIDs
	featureProjectScopedClusters
	featureProjectScopedRepositories
	featureClusterMetadata
	featureRepositoryCertificates
	featureApplicationHelmSkipCrds
	featureExecLogsPolicy
	featureProjectSourceNamespaces
	featureMultipleApplicationSources
	featureApplicationSet
	featureApplicationSetProgressiveSync
)

var featureVersionConstraintsMap = map[int]*semver.Version{
	featureApplicationLevelSyncOptions:   semver.MustParse("1.5.0"),
	featureIgnoreDiffJQPathExpressions:   semver.MustParse("2.1.0"),
	featureRepositoryGet:                 semver.MustParse("1.6.0"),
	featureTokenIDs:                      semver.MustParse("1.5.3"),
	featureProjectScopedClusters:         semver.MustParse("2.2.0"),
	featureProjectScopedRepositories:     semver.MustParse("2.2.0"),
	featureClusterMetadata:               semver.MustParse("2.2.0"),
	featureRepositoryCertificates:        semver.MustParse("1.2.0"),
	featureApplicationHelmSkipCrds:       semver.MustParse("2.3.0"),
	featureExecLogsPolicy:                semver.MustParse("2.4.4"),
	featureProjectSourceNamespaces:       semver.MustParse("2.5.0"),
	featureMultipleApplicationSources:    semver.MustParse("2.6.3"), // Whilst the feature was introduced in 2.6.0 there was a bug that affects refresh of applications (and hence `wait` within this provider) that was only fixed in https://github.com/argoproj/argo-cd/pull/12576
	featureApplicationSet:                semver.MustParse("2.5.0"),
	featureApplicationSetProgressiveSync: semver.MustParse("2.6.0"),
}

type ServerInterface struct {
	AccountClient        account.AccountServiceClient
	ApiClient            apiclient.Client
	ApplicationClient    application.ApplicationServiceClient
	ApplicationSetClient applicationset.ApplicationSetServiceClient
	CertificateClient    certificate.CertificateServiceClient
	ClusterClient        cluster.ClusterServiceClient
	ProjectClient        project.ProjectServiceClient
	RepositoryClient     repository.RepositoryServiceClient
	RepoCredsClient      repocreds.RepoCredsServiceClient
	SessionClient        session.SessionServiceClient

	ServerVersion        *semver.Version
	ServerVersionMessage *version.VersionMessage
	ProviderData         *schema.ResourceData

	sync.Mutex
	initialized bool
}

func (p *ServerInterface) initClients(ctx context.Context) error {
	if p.initialized {
		return nil
	}

	d := p.ProviderData

	p.Lock()
	defer p.Unlock()

	if p.ApiClient == nil {
		apiClient, err := initApiClient(ctx, d)
		if err != nil {
			return err
		}

		p.ApiClient = apiClient
	}

	if p.AccountClient == nil {
		_, accountClient, err := p.ApiClient.NewAccountClient()
		if err != nil {
			return err
		}

		p.AccountClient = accountClient
	}

	if p.ClusterClient == nil {
		_, clusterClient, err := p.ApiClient.NewClusterClient()
		if err != nil {
			return err
		}

		p.ClusterClient = clusterClient
	}

	if p.CertificateClient == nil {
		_, certClient, err := p.ApiClient.NewCertClient()
		if err != nil {
			return err
		}

		p.CertificateClient = certClient
	}

	if p.ApplicationClient == nil {
		_, applicationClient, err := p.ApiClient.NewApplicationClient()
		if err != nil {
			return err
		}

		p.ApplicationClient = applicationClient
	}

	if p.ApplicationSetClient == nil {
		_, applicationSetClient, err := (p.ApiClient).NewApplicationSetClient()
		if err != nil {
			return err
		}

		p.ApplicationSetClient = applicationSetClient
	}

	if p.ProjectClient == nil {
		_, projectClient, err := p.ApiClient.NewProjectClient()
		if err != nil {
			return err
		}

		p.ProjectClient = projectClient
	}

	if p.RepositoryClient == nil {
		_, repositoryClient, err := p.ApiClient.NewRepoClient()
		if err != nil {
			return err
		}

		p.RepositoryClient = repositoryClient
	}

	if p.RepoCredsClient == nil {
		_, repoCredsClient, err := p.ApiClient.NewRepoCredsClient()
		if err != nil {
			return err
		}

		p.RepoCredsClient = repoCredsClient
	}

	if p.SessionClient == nil {
		_, sessionClient, err := p.ApiClient.NewSessionClient()
		if err != nil {
			return err
		}

		p.SessionClient = sessionClient
	}

	acCloser, versionClient, err := p.ApiClient.NewVersionClient()
	if err != nil {
		return err
	}
	defer io.Close(acCloser)

	serverVersionMessage, err := versionClient.Version(ctx, &empty.Empty{})
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
	p.initialized = true

	return nil
}

// Checks that a specific feature is available for the current ArgoCD server version.
// 'feature' argument must match one of the predefined feature* constants.
func (p *ServerInterface) isFeatureSupported(feature int) (bool, error) {
	versionConstraint, ok := featureVersionConstraintsMap[feature]
	if !ok {
		return false, fmt.Errorf("feature constraint is not handled by the provider")
	}

	if i := versionConstraint.Compare(p.ServerVersion); i == 1 {
		return false, nil
	}

	return true, nil
}

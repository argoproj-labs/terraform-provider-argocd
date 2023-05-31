package argocd

import (
	"context"
	"fmt"
	"sync"

	"github.com/Masterminds/semver/v3"
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
	"github.com/oboukili/terraform-provider-argocd/internal/features"
)

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
func (p *ServerInterface) isFeatureSupported(feature features.Feature) bool {
	fc, ok := features.ConstraintsMap[feature]

	return ok && fc.MinVersion.Compare(p.ServerVersion) != 1
}

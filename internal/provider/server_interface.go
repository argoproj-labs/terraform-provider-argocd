package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/Masterminds/semver/v3"
	"github.com/argoproj-labs/terraform-provider-argocd/internal/diagnostics"
	"github.com/argoproj-labs/terraform-provider-argocd/internal/features"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/account"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/applicationset"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/certificate"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/gpgkey"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/project"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/repocreds"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/repository"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/session"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/version"
	"github.com/argoproj/argo-cd/v3/util/io"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/protobuf/types/known/emptypb"
	"k8s.io/apimachinery/pkg/util/runtime"
)

var runtimeErrorHandlers []runtime.ErrorHandler

type ServerInterface struct {
	AccountClient        account.AccountServiceClient
	ApiClient            apiclient.Client
	ApplicationClient    application.ApplicationServiceClient
	ApplicationSetClient applicationset.ApplicationSetServiceClient
	CertificateClient    certificate.CertificateServiceClient
	ClusterClient        cluster.ClusterServiceClient
	GPGKeysClient        gpgkey.GPGKeyServiceClient
	ProjectClient        project.ProjectServiceClient
	RepoCredsClient      repocreds.RepoCredsServiceClient
	RepositoryClient     repository.RepositoryServiceClient
	SessionClient        session.SessionServiceClient

	ServerVersion        *semver.Version
	ServerVersionMessage *version.VersionMessage

	config      ArgoCDProviderConfig
	initialized bool
	sync.RWMutex
}

func NewServerInterface(c ArgoCDProviderConfig) *ServerInterface {
	return &ServerInterface{
		config: c,
	}
}

func (si *ServerInterface) InitClients(ctx context.Context) diag.Diagnostics {
	si.Lock()
	defer si.Unlock()

	if si.initialized {
		return nil
	}

	opts, d := si.config.getApiClientOptions(ctx)
	if d.HasError() {
		return d
	}

	ac, err := apiclient.NewClient(opts)
	if err != nil {
		return diagnostics.Error("failed to create new API client", err)
	}

	var diags diag.Diagnostics

	_, si.AccountClient, err = ac.NewAccountClient()
	if err != nil {
		diags.Append(diagnostics.Error("failed to initialize account client", err)...)
	}

	_, si.ApplicationClient, err = ac.NewApplicationClient()
	if err != nil {
		diags.Append(diagnostics.Error("failed to initialize application client", err)...)
	}

	_, si.ApplicationSetClient, err = ac.NewApplicationSetClient()
	if err != nil {
		diags.Append(diagnostics.Error("failed to initialize application set client", err)...)
	}

	_, si.CertificateClient, err = ac.NewCertClient()
	if err != nil {
		diags.Append(diagnostics.Error("failed to initialize certificate client", err)...)
	}

	_, si.ClusterClient, err = ac.NewClusterClient()
	if err != nil {
		diags.Append(diagnostics.Error("failed to initialize cluster client", err)...)
	}

	_, si.GPGKeysClient, err = ac.NewGPGKeyClient()
	if err != nil {
		diags.Append(diagnostics.Error("failed to initialize GPG keys client", err)...)
	}

	_, si.ProjectClient, err = ac.NewProjectClient()
	if err != nil {
		diags.Append(diagnostics.Error("failed to initialize project client", err)...)
	}

	_, si.RepositoryClient, err = ac.NewRepoClient()
	if err != nil {
		diags.Append(diagnostics.Error("failed to initialize repository client", err)...)
	}

	_, si.RepoCredsClient, err = ac.NewRepoCredsClient()
	if err != nil {
		diags.Append(diagnostics.Error("failed to initialize repository credentials client", err)...)
	}

	_, si.SessionClient, err = ac.NewSessionClient()
	if err != nil {
		diags.Append(diagnostics.Error("failed to initialize session client", err)...)
	}

	acCloser, versionClient, err := ac.NewVersionClient()
	if err != nil {
		diags.Append(diagnostics.Error("failed to initialize version client", err)...)
	} else {
		defer io.Close(acCloser)

		serverVersionMessage, err := versionClient.Version(ctx, &emptypb.Empty{})
		if err != nil {
			return diagnostics.Error("failed to read server version", err)
		}

		if serverVersionMessage == nil {
			return diagnostics.Error("could not get server version information", nil)
		}

		si.ServerVersionMessage = serverVersionMessage

		serverVersion, err := semver.NewVersion(serverVersionMessage.Version)
		if err != nil {
			diags.Append(diagnostics.Error(fmt.Sprintf("could not parse server semantic version: %s", serverVersionMessage.Version), nil)...)
		}

		si.ServerVersion = serverVersion
	}

	si.initialized = !diags.HasError()

	return diags
}

// Checks that the server version meets the minimum version required for the feature.
func (si *ServerInterface) IsVersionSupported(fc features.FeatureConstraint) bool {
	if fc.MinVersion == nil {
		return true
	}

	return fc.MinVersion.Compare(si.ServerVersion) != 1
}

// Checks that a specific feature is available for the current ArgoCD server version.
// 'feature' argument must match one of the predefined feature* constants.
func (si *ServerInterface) IsFeatureSupported(feature features.Feature) bool {
	fc, ok := features.ConstraintsMap[feature]

	return ok && fc.MinVersion.Compare(si.ServerVersion) != 1
}

func getDefaultString(s types.String, envKey string) string {
	if !s.IsNull() && !s.IsUnknown() {
		return s.ValueString()
	}

	return os.Getenv(envKey)
}

func getDefaultBool(ctx context.Context, b types.Bool, envKey string) bool {
	if !b.IsNull() && !b.IsUnknown() {
		return b.ValueBool()
	}

	env, ok := os.LookupEnv(envKey)
	if !ok {
		return false
	}

	pb, err := strconv.ParseBool(env)
	if err == nil {
		return pb
	}

	tflog.Warn(ctx, fmt.Sprintf("failed to parse env var %s with value %s as bool. Will default to `false`.", envKey, env))

	return false
}

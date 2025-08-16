package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/providervalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ArgoCDProvider satisfies various provider interfaces.
var _ provider.Provider = (*ArgoCDProvider)(nil)

type ArgoCDProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func New(version string) provider.Provider {
	return &ArgoCDProvider{
		version: version,
	}
}

func (p *ArgoCDProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "argocd"
}

func (p *ArgoCDProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"auth_token": schema.StringAttribute{
				Description: "ArgoCD authentication token, takes precedence over `username`/`password`. Can be set through the `ARGOCD_AUTH_TOKEN` environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"username": schema.StringAttribute{
				Description: "Authentication username. Can be set through the `ARGOCD_AUTH_USERNAME` environment variable.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "Authentication password. Can be set through the `ARGOCD_AUTH_PASSWORD` environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"core": schema.BoolAttribute{
				Description: "Configure direct access using Kubernetes API server.\n\n  " +
					"**Warning**: this feature works by starting a local ArgoCD API server that talks directly to the Kubernetes API using the **current context " +
					"in the default kubeconfig** (`~/.kube/config`). This behavior cannot be overridden using either environment variables or the `kubernetes` block " +
					"in the provider configuration at present).\n\n  If the server fails to start (e.g. your kubeconfig is misconfigured) then the provider will " +
					"fail as a result of the `argocd` module forcing it to exit and no logs will be available to help you debug this. The error message will be " +
					"similar to\n  > `The plugin encountered an error, and failed to respond to the plugin.(*GRPCProvider).ReadResource call. The plugin logs may " +
					"contain more details.`\n\n  To debug this, you will need to login via the ArgoCD CLI using `argocd login --core` and then running an operation. " +
					"E.g. `argocd app list`.",
				Optional: true,
			},
			"server_addr": schema.StringAttribute{
				Description: "ArgoCD server address with port. Can be set through the `ARGOCD_SERVER` environment variable.",
				Optional:    true,
			},
			"port_forward": schema.BoolAttribute{
				Description: "Connect to a random argocd-server port using port forwarding.",
				Optional:    true,
			},
			"port_forward_with_namespace": schema.StringAttribute{
				Description: "Namespace name which should be used for port forwarding.",
				Optional:    true,
			},
			"use_local_config": schema.BoolAttribute{
				Description: "Use the authentication settings found in the local config file. Useful when you have previously logged in using SSO. Conflicts with `auth_token`, `username` and `password`.",
				Optional:    true,
			},
			"config_path": schema.StringAttribute{
				Description: "Override the default config path of `$HOME/.config/argocd/config`. Only relevant when `use_local_config`. Can be set through the `ARGOCD_CONFIG_PATH` environment variable.",
				Optional:    true,
			},
			"context": schema.StringAttribute{
				Description: "Context to choose when using a local ArgoCD config file. Only relevant when `use_local_config`. Can be set through `ARGOCD_CONTEXT` environment variable.",
				Optional:    true,
			},
			"cert_file": schema.StringAttribute{
				Description: "Additional root CA certificates file to add to the client TLS connection pool.",
				Optional:    true,
			},
			"client_cert_file": schema.StringAttribute{
				Description: "Client certificate.",
				Optional:    true,
			},
			"client_cert_key": schema.StringAttribute{
				Description: "Client certificate key.",
				Optional:    true,
			},
			"grpc_web": schema.BoolAttribute{
				Description: "Whether to use gRPC web proxy client. Useful if Argo CD server is behind proxy which does not support HTTP2.",
				Optional:    true,
			},
			"grpc_web_root_path": schema.StringAttribute{
				Description: "Use the gRPC web proxy client and set the web root, e.g. `argo-cd`. Useful if the Argo CD server is behind a proxy at a non-root path.",
				Optional:    true,
			},
			"headers": schema.SetAttribute{
				Description: "Additional headers to add to each request to the ArgoCD server.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"insecure": schema.BoolAttribute{
				Description: "Whether to skip TLS server certificate. Can be set through the `ARGOCD_INSECURE` environment variable.",
				Optional:    true,
			},
			"plain_text": schema.BoolAttribute{
				Description: "Whether to initiate an unencrypted connection to ArgoCD server.",
				Optional:    true,
			},
			"user_agent": schema.StringAttribute{
				Description: "User-Agent request header override.",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"kubernetes": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				Description: "Kubernetes configuration overrides.  Only relevant when `port_forward = true` or `port_forward_with_namespace = \"foo\"`. The kubeconfig file that is used can be overridden using the [`KUBECONFIG` environment variable](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/#the-kubeconfig-environment-variable)).",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"host": schema.StringAttribute{
							Optional:    true,
							Description: "The hostname (in form of URI) of the Kubernetes API. Can be sourced from `KUBE_HOST`.",
						},
						"username": schema.StringAttribute{
							Description: "The username to use for HTTP basic authentication when accessing the Kubernetes API. Can be sourced from `KUBE_USER`.",
							Optional:    true,
						},
						"password": schema.StringAttribute{
							Description: "The password to use for HTTP basic authentication when accessing the Kubernetes API. Can be sourced from `KUBE_PASSWORD`.",
							Optional:    true,
							Sensitive:   true,
						},
						"insecure": schema.BoolAttribute{
							Description: "Whether server should be accessed without verifying the TLS certificate. Can be sourced from `KUBE_INSECURE`.",
							Optional:    true,
						},
						"client_certificate": schema.StringAttribute{
							Description: "PEM-encoded client certificate for TLS authentication. Can be sourced from `KUBE_CLIENT_CERT_DATA`.",
							Optional:    true,
						},
						"client_key": schema.StringAttribute{
							Description: "PEM-encoded client certificate key for TLS authentication. Can be sourced from `KUBE_CLIENT_KEY_DATA`.",
							Optional:    true,
							Sensitive:   true,
						},
						"cluster_ca_certificate": schema.StringAttribute{
							Description: "PEM-encoded root certificates bundle for TLS authentication. Can be sourced from `KUBE_CLUSTER_CA_CERT_DATA`.",
							Optional:    true,
						},
						"config_context": schema.StringAttribute{
							Description: "Context to choose from the config file. Can be sourced from `KUBE_CTX`.",
							Optional:    true,
						},
						"config_context_auth_info": schema.StringAttribute{
							Description: "",
							Optional:    true,
						},
						"config_context_cluster": schema.StringAttribute{
							Description: "",
							Optional:    true,
						},
						"token": schema.StringAttribute{
							Description: "Token to authenticate an service account. Can be sourced from `KUBE_TOKEN`.",
							Optional:    true,
							Sensitive:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"exec": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							Description: "Configuration block to use an [exec-based credential plugin](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#client-go-credential-plugins), e.g. call an external command to receive user credentials.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"api_version": schema.StringAttribute{
										Description: "API version to use when decoding the ExecCredentials resource, e.g. `client.authentication.k8s.io/v1beta1`.",
										Required:    true,
									},
									"command": schema.StringAttribute{
										Description: "Command to execute.",
										Required:    true,
									},
									"env": schema.MapAttribute{
										Description: "List of arguments to pass when executing the plugin.",
										Optional:    true,
										ElementType: types.StringType,
									},
									"args": schema.ListAttribute{
										Description: "Map of environment variables to set when executing the plugin.",
										Optional:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (p *ArgoCDProvider) ConfigValidators(ctx context.Context) []provider.ConfigValidator {
	return []provider.ConfigValidator{
		// Don't mix/match different mechanisms used to determine which server to speak to (i.e. how ArgoCD API server is exposed or whether to expose it locally)
		providervalidator.Conflicting(
			path.MatchRoot("port_forward"),
			path.MatchRoot("port_forward_with_namespace"),
			path.MatchRoot("server_addr"),
			path.MatchRoot("use_local_config"),
			path.MatchRoot("core"),
		),
		// Don't mix/match different authentication mechanisms
		providervalidator.Conflicting(
			path.MatchRoot("auth_token"),
			path.MatchRoot("password"),
			path.MatchRoot("use_local_config"),
			path.MatchRoot("core"),
		),
	}
}

func (p *ArgoCDProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ArgoCDProviderConfig

	// Read configuration into model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	server := NewServerInterface(config)

	resp.DataSourceData = server
	resp.ResourceData = server
}

func (p *ArgoCDProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewGPGKeyResource,
		NewRepositoryResource,
		NewRepositoryCertificateResource,
		NewRepositoryCredentialsResource,
		NewProjectResource,
		NewProjectTokenResource,
	}
}

func (p *ArgoCDProvider) DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewArgoCDApplicationDataSource,
	}
}

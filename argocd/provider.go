package argocd

import (
	"context"
	"sync"

	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	// Import to initialize client auth plugins.

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// Used to handle concurrent access to ArgoCD common configuration
var tokenMutexConfiguration = &sync.RWMutex{}

// Used to handle concurrent access to ArgoCD clusters
var tokenMutexClusters = &sync.RWMutex{}

// Used to handle concurrent access to each ArgoCD project
var tokenMutexProjectMap = make(map[string]*sync.RWMutex, 0)

// Used to handle concurrent access to ArgoCD secrets
var tokenMutexSecrets = &sync.RWMutex{}

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"server_addr": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ArgoCD server address with port. Can be set through the `ARGOCD_SERVER` environment variable.",
			},
			"auth_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ArgoCD authentication token, takes precedence over `username`/`password`. Can be set through the `ARGOCD_AUTH_TOKEN` environment variable.",
				Sensitive:   true,
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Authentication username. Can be set through the `ARGOCD_AUTH_USERNAME` environment variable.",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Authentication password. Can be set through the `ARGOCD_AUTH_PASSWORD` environment variable.",
				Sensitive:   true,
			},
			"cert_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Additional root CA certificates file to add to the client TLS connection pool.",
			},
			"client_cert_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Client certificate.",
			},
			"client_cert_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Client certificate key.",
			},
			"plain_text": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether to initiate an unencrypted connection to ArgoCD server.",
			},
			"context": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Context to choose when using a local ArgoCD config file. Only relevant when `use_local_config`. Can be set through `ARGOCD_CONTEXT` environment variable.",
			},
			"user_agent": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User-Agent request header override.",
			},
			"core": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: "Configure direct access using Kubernetes API server.\n\n  " +
					"**Warning**: this feature works by starting a local ArgoCD API server that talks directly to the Kubernetes API using the **current context " +
					"in the default kubeconfig** (`~/.kube/config`). This behavior cannot be overridden using either environment variables or the `kubernetes` block " +
					"in the provider configuration at present).\n\n  If the server fails to start (e.g. your kubeconfig is misconfigured) then the provider will " +
					"fail as a result of the `argocd` module forcing it to exit and no logs will be available to help you debug this. The error message will be " +
					"similar to\n  > `The plugin encountered an error, and failed to respond to the plugin.(*GRPCProvider).ReadResource call. The plugin logs may " +
					"contain more details.`\n\n  To debug this, you will need to login via the ArgoCD CLI using `argocd login --core` and then running an operation. " +
					"E.g. `argocd app list`.",
			},
			"grpc_web": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether to use gRPC web proxy client. Useful if Argo CD server is behind proxy which does not support HTTP2.",
			},
			"grpc_web_root_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Use the gRPC web proxy client and set the web root, e.g. `argo-cd`. Useful if the Argo CD server is behind a proxy at a non-root path.",
			},
			"use_local_config": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Use the authentication settings found in the local config file. Useful when you have previously logged in using SSO. Conflicts with `auth_token`, `username` and `password`.",
			},
			"config_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Override the default config path of `$HOME/.config/argocd/config`. Only relevant when `use_local_config`. Can be set through the `ARGOCD_CONFIG_PATH` environment variable.",
			},
			"port_forward": {
				Type:        schema.TypeBool,
				Description: "Connect to a random argocd-server port using port forwarding.",
				Optional:    true,
			},
			"port_forward_with_namespace": {
				Type:        schema.TypeString,
				Description: "Namespace name which should be used for port forwarding.",
				Optional:    true,
			},
			"headers": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Additional headers to add to each request to the ArgoCD server.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether to skip TLS server certificate. Can be set through the `ARGOCD_INSECURE` environment variable.",
			},
			"kubernetes": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Kubernetes configuration overrides.  Only relevant when `port_forward = true` or `port_forward_with_namespace = \"foo\"`. The kubeconfig file that is used can be overridden using the [`KUBECONFIG` environment variable](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/#the-kubeconfig-environment-variable)).",
				Elem:        kubernetesResource(),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"argocd_account_token":   resourceArgoCDAccountToken(),
			"argocd_application":     resourceArgoCDApplication(),
			"argocd_application_set": resourceArgoCDApplicationSet(),
			"argocd_cluster":         resourceArgoCDCluster(),
		},
		ConfigureContextFunc: func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
			config, diags := argoCDProviderConfigFromResourceData(ctx, d)

			server := NewServerInterface(config)

			return server, diags
		},
	}
}

func kubernetesResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The hostname (in form of URI) of the Kubernetes API. Can be sourced from `KUBE_HOST`.",
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The username to use for HTTP basic authentication when accessing the Kubernetes API. Can be sourced from `KUBE_USER`.",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The password to use for HTTP basic authentication when accessing the Kubernetes API. Can be sourced from `KUBE_PASSWORD`.",
				Sensitive:   true,
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether server should be accessed without verifying the TLS certificate. Can be sourced from `KUBE_INSECURE`.",
			},
			"client_certificate": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "PEM-encoded client certificate for TLS authentication. Can be sourced from `KUBE_CLIENT_CERT_DATA`.",
			},
			"client_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "PEM-encoded client certificate key for TLS authentication. Can be sourced from `KUBE_CLIENT_KEY_DATA`.",
				Sensitive:   true,
			},
			"cluster_ca_certificate": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "PEM-encoded root certificates bundle for TLS authentication. Can be sourced from `KUBE_CLUSTER_CA_CERT_DATA`.",
			},
			"config_context": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Context to choose from the config file. Can be sourced from `KUBE_CTX`.",
			},
			"config_context_auth_info": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "",
			},
			"config_context_cluster": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "",
			},
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Token to authenticate an service account. Can be sourced from `KUBE_TOKEN`.",
				Sensitive:   true,
			},
			"exec": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Configuration block to use an [exec-based credential plugin](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#client-go-credential-plugins), e.g. call an external command to receive user credentials.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"api_version": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "API version to use when decoding the ExecCredentials resource, e.g. `client.authentication.k8s.io/v1beta1`.",
						},
						"command": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Command to execute.",
						},
						"env": {
							Type:        schema.TypeMap,
							Optional:    true,
							Description: "List of arguments to pass when executing the plugin.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"args": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Map of environment variables to set when executing the plugin.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func argoCDProviderConfigFromResourceData(ctx context.Context, d *schema.ResourceData) (ArgoCDProviderConfig, diag.Diagnostics) {
	c := ArgoCDProviderConfig{
		AuthToken:                getStringFromResourceData(d, "auth_token"),
		CertFile:                 getStringFromResourceData(d, "cert_file"),
		ClientCertFile:           getStringFromResourceData(d, "client_cert_file"),
		ClientCertKey:            getStringFromResourceData(d, "client_cert_key"),
		ConfigPath:               getStringFromResourceData(d, "config_path"),
		Context:                  getStringFromResourceData(d, "context"),
		Core:                     getBoolFromResourceData(d, "core"),
		GRPCWeb:                  getBoolFromResourceData(d, "grpc_web"),
		GRPCWebRootPath:          getStringFromResourceData(d, "grpc_web_root_path"),
		Insecure:                 getBoolFromResourceData(d, "insecure"),
		Password:                 getStringFromResourceData(d, "password"),
		PlainText:                getBoolFromResourceData(d, "plain_text"),
		PortForward:              getBoolFromResourceData(d, "port_forward"),
		PortForwardWithNamespace: getStringFromResourceData(d, "port_forward_with_namespace"),
		ServerAddr:               getStringFromResourceData(d, "server_addr"),
		UseLocalConfig:           getBoolFromResourceData(d, "use_local_config"),
		UserAgent:                getStringFromResourceData(d, "user_agent"),
		Username:                 getStringFromResourceData(d, "username"),
	}

	headers, diags := getStringSetFromResourceData(ctx, d, "headers")
	c.Headers = headers

	k8s, ds := kubernetesConfigFromResourceData(ctx, d)
	c.Kubernetes = k8s

	diags.Append(ds...)

	return c, pluginSDKDiags(diags)
}

func kubernetesConfigFromResourceData(ctx context.Context, d *schema.ResourceData) ([]Kubernetes, fwdiag.Diagnostics) {
	if _, ok := d.GetOk("kubernetes"); !ok {
		return nil, nil
	}

	k8s := Kubernetes{
		ClientCertificate:     getStringFromResourceData(d, "kubernetes.0.client_certificate"),
		ClientKey:             getStringFromResourceData(d, "kubernetes.0.client_key"),
		ClusterCACertificate:  getStringFromResourceData(d, "kubernetes.0.cluster_ca_certificate"),
		ConfigContext:         getStringFromResourceData(d, "kubernetes.0.config_context"),
		ConfigContextAuthInfo: getStringFromResourceData(d, "kubernetes.0.config_context_auth_info"),
		ConfigContextCluster:  getStringFromResourceData(d, "kubernetes.0.config_context_cluster"),
		Host:                  getStringFromResourceData(d, "kubernetes.0.host"),
		Insecure:              getBoolFromResourceData(d, "kubernetes.0.insecure"),
		Password:              getStringFromResourceData(d, "kubernetes.0.password"),
		Token:                 getStringFromResourceData(d, "kubernetes.0.token"),
		Username:              getStringFromResourceData(d, "kubernetes.0.username"),
	}

	var diags fwdiag.Diagnostics

	k8s.Exec, diags = kubernetesExecConfigFromResourceData(ctx, d)

	return []Kubernetes{k8s}, diags
}

func kubernetesExecConfigFromResourceData(ctx context.Context, d *schema.ResourceData) ([]KubernetesExec, fwdiag.Diagnostics) {
	if _, ok := d.GetOk("kubernetes.0.exec"); !ok {
		return nil, nil
	}

	exec := KubernetesExec{
		APIVersion: getStringFromResourceData(d, "kubernetes.0.exec.0.api_version"),
		Command:    getStringFromResourceData(d, "kubernetes.0.exec.0.command"),
	}

	args, diags := getStringListFromResourceData(ctx, d, "kubernetes.0.exec.0.args")
	exec.Args = args

	env, ds := getStringMapFromResourceData(ctx, d, "kubernetes.0.exec.0.env")
	exec.Env = env

	diags.Append(ds...)

	return []KubernetesExec{exec}, diags
}

func getStringFromResourceData(d *schema.ResourceData, key string) types.String {
	if v, ok := d.GetOk(key); ok {
		return types.StringValue(v.(string))
	}

	return types.StringNull()
}

func getBoolFromResourceData(d *schema.ResourceData, key string) types.Bool {
	if v, ok := d.GetOk(key); ok {
		return types.BoolValue(v.(bool))
	}

	return types.BoolNull()
}

func getStringListFromResourceData(ctx context.Context, d *schema.ResourceData, key string) (types.List, fwdiag.Diagnostics) {
	if v, ok := d.GetOk(key); ok {
		return types.ListValueFrom(ctx, types.StringType, v.([]interface{}))
	}

	return types.ListNull(types.StringType), nil
}

func getStringMapFromResourceData(ctx context.Context, d *schema.ResourceData, key string) (types.Map, fwdiag.Diagnostics) {
	if v, ok := d.GetOk(key); ok {
		return types.MapValueFrom(ctx, types.StringType, v.(map[string]interface{}))
	}

	return types.MapNull(types.StringType), nil
}

func getStringSetFromResourceData(ctx context.Context, d *schema.ResourceData, key string) (types.Set, fwdiag.Diagnostics) {
	if v, ok := d.GetOk(key); ok {
		return types.SetValueFrom(ctx, types.StringType, v.(*schema.Set).List())
	}

	return types.SetNull(types.StringType), nil
}

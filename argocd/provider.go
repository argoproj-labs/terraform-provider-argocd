package argocd

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"sync"

	"github.com/argoproj/argo-cd/v2/cmd/argocd/commands/headless"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/session"
	"github.com/argoproj/argo-cd/v2/util/io"
	"github.com/argoproj/argo-cd/v2/util/localconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	apimachineryschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/runtime"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// Used to handle concurrent access to ArgoCD common configuration
var tokenMutexConfiguration = &sync.RWMutex{}

// Used to handle concurrent access to ArgoCD clusters
var tokenMutexClusters = &sync.RWMutex{}

// Used to handle concurrent access to each ArgoCD project
var tokenMutexProjectMap = make(map[string]*sync.RWMutex, 0)

var runtimeErrorHandlers []func(error)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"server_addr": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARGOCD_SERVER", nil),
				Description: "ArgoCD server address with port. Can be set through the `ARGOCD_SERVER` environment variable.",
				AtLeastOneOf: []string{
					"core",
					"port_forward",
					"port_forward_with_namespace",
					"use_local_config",
				},
			},
			"auth_token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARGOCD_AUTH_TOKEN", nil),
				Description: "ArgoCD authentication token, takes precedence over `username`/`password`. Can be set through the `ARGOCD_AUTH_TOKEN` environment variable.",
				ConflictsWith: []string{
					"config_path",
					"core",
					"password",
					"use_local_config",
					"username",
				},
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARGOCD_AUTH_USERNAME", nil),
				Description: "Authentication username. Can be set through the `ARGOCD_AUTH_USERNAME` environment variable.",
				ConflictsWith: []string{
					"auth_token",
					"config_path",
					"core",
					"use_local_config",
				},
				AtLeastOneOf: []string{
					"core",
					"password",
					"auth_token",
					"use_local_config",
				},
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARGOCD_AUTH_PASSWORD", nil),
				Description: "Authentication password. Can be set through the `ARGOCD_AUTH_PASSWORD` environment variable.",
				ConflictsWith: []string{
					"auth_token",
					"config_path",
					"core",
					"use_local_config",
				},
				AtLeastOneOf: []string{
					"username",
					"core",
					"auth_token",
					"use_local_config",
				},
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
				Default:     false,
				Description: "Whether to initiate an unencrypted connection to ArgoCD server.",
			},
			"context": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARGOCD_CONTEXT", nil),
				Description: "Context to choose when using a local ArgoCD config file. Only relevant when `use_local_config`. Can be set through `ARGOCD_CONTEXT` environment variable.",
				ConflictsWith: []string{
					"core",
					"username",
					"password",
					"auth_token",
				},
			},
			"user_agent": {
				Type:     schema.TypeString,
				Optional: true,
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
				ConflictsWith: []string{
					"auth_token",
					"use_local_config",
					"password",
					"username",
				},
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
				ConflictsWith: []string{
					"auth_token",
					"core",
					"password",
					"username",
				},
			},
			"config_path": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARGOCD_CONFIG_PATH", nil),
				Description: "Override the default config path of `$HOME/.config/argocd/config`. Only relevant when `use_local_config`. Can be set through the `ARGOCD_CONFIG_PATH` environment variable.",
				ConflictsWith: []string{
					"auth_token",
					"core",
					"password",
					"username",
				},
			},
			"port_forward": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"port_forward_with_namespace": {
				Type:     schema.TypeString,
				Optional: true,
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
				DefaultFunc: schema.EnvDefaultFunc("ARGOCD_INSECURE", false),
				Description: "Whether to skip TLS server certificate. Can be set through the `ARGOCD_INSECURE` environment variable.",
			},
			"kubernetes": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Kubernetes configuration.",
				Elem:        kubernetesResource(),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"argocd_application":            resourceArgoCDApplication(),
			"argocd_repository_certificate": resourceArgoCDRepositoryCertificates(),
			"argocd_cluster":                resourceArgoCDCluster(),
			"argocd_project":                resourceArgoCDProject(),
			"argocd_project_token":          resourceArgoCDProjectToken(),
			"argocd_repository":             resourceArgoCDRepository(),
			"argocd_repository_credentials": resourceArgoCDRepositoryCredentials(),
		},
		ConfigureFunc: func(d *schema.ResourceData) (interface{}, error) {
			server := ServerInterface{ProviderData: d}
			return &server, nil
		},
	}
}

func initApiClient(ctx context.Context, d *schema.ResourceData) (apiClient apiclient.Client, err error) {
	var opts apiclient.ClientOptions

	if v, ok := d.GetOk("server_addr"); ok {
		opts.ServerAddr = v.(string)
	}

	if v, ok := d.GetOk("use_local_config"); ok {
		if v.(bool) {
			if v, ok := d.GetOk("config_path"); ok {
				opts.ConfigPath = v.(string)
			} else if opts.ConfigPath, err = localconfig.DefaultLocalConfigPath(); err != nil {
				return nil, err
			}
		}
	}

	if v, ok := d.GetOk("plain_text"); ok {
		opts.PlainText = v.(bool)
	}

	if v, ok := d.GetOk("insecure"); ok {
		opts.Insecure = v.(bool)
	}

	if v, ok := d.GetOk("cert_file"); ok {
		opts.CertFile = v.(string)
	}

	if v, ok := d.GetOk("client_cert_file"); ok {
		opts.ClientCertFile = v.(string)
	}

	if v, ok := d.GetOk("client_cert_key"); ok {
		opts.ClientCertKeyFile = v.(string)
	}

	if v, ok := d.GetOk("context"); ok {
		opts.Context = v.(string)
	}

	if v, ok := d.GetOk("user_agent"); ok {
		opts.UserAgent = v.(string)
	}

	if v, ok := d.GetOk("grpc_web"); ok {
		opts.GRPCWeb = v.(bool)
	}

	if v, ok := d.GetOk("grpc_web_root_path"); ok {
		opts.GRPCWebRootPath = v.(string)
	}

	if v, ok := d.GetOk("port_forward"); ok {
		opts.PortForward = v.(bool)
	}

	if v, ok := d.GetOk("port_forward_with_namespace"); ok {
		opts.PortForwardNamespace = v.(string)
	}

	if v, ok := d.GetOk("headers"); ok {
		_headers := v.(*schema.Set).List()

		var headers = make([]string, len(_headers))

		for i, _header := range _headers {
			headers[i] = _header.(string)
		}

		opts.Headers = headers
	}

	if _, ok := d.GetOk("kubernetes"); ok {
		opts.KubeOverrides = &clientcmd.ConfigOverrides{}

		if v, ok := k8sGetOk(d, "insecure"); ok {
			opts.KubeOverrides.ClusterInfo.InsecureSkipTLSVerify = v.(bool)
		}

		if v, ok := k8sGetOk(d, "cluster_ca_certificate"); ok {
			opts.KubeOverrides.ClusterInfo.CertificateAuthorityData = bytes.NewBufferString(v.(string)).Bytes()
		}

		if v, ok := k8sGetOk(d, "client_certificate"); ok {
			opts.KubeOverrides.AuthInfo.ClientCertificateData = bytes.NewBufferString(v.(string)).Bytes()
		}

		kubectx, ctxOk := k8sGetOk(d, "config_context")
		authInfo, authInfoOk := k8sGetOk(d, "config_context_auth_info")
		cluster, clusterOk := k8sGetOk(d, "config_context_cluster")

		if ctxOk || authInfoOk || clusterOk {
			if ctxOk {
				opts.KubeOverrides.CurrentContext = kubectx.(string)
			}

			opts.KubeOverrides.Context = clientcmdapi.Context{}
			if authInfoOk {
				opts.KubeOverrides.Context.AuthInfo = authInfo.(string)
			}

			if clusterOk {
				opts.KubeOverrides.Context.Cluster = cluster.(string)
			}
		}

		if v, ok := k8sGetOk(d, "host"); ok {
			// Server has to be the complete address of the kubernetes cluster (scheme://hostname:port), not just the hostname,
			// because `overrides` are processed too late to be taken into account by `defaultServerUrlFor()`.
			// This basically replicates what defaultServerUrlFor() does with config but for overrides,
			// see https://github.com/kubernetes/client-go/blob/v12.0.0/rest/url_utils.go#L85-L87
			hasCA := len(opts.KubeOverrides.ClusterInfo.CertificateAuthorityData) != 0
			hasCert := len(opts.KubeOverrides.AuthInfo.ClientCertificateData) != 0
			defaultTLS := hasCA || hasCert || opts.KubeOverrides.ClusterInfo.InsecureSkipTLSVerify

			var host *url.URL

			host, _, err = rest.DefaultServerURL(v.(string), "", apimachineryschema.GroupVersion{}, defaultTLS)
			if err != nil {
				return nil, err
			}

			opts.KubeOverrides.ClusterInfo.Server = host.String()
		}

		if v, ok := k8sGetOk(d, "username"); ok {
			opts.KubeOverrides.AuthInfo.Username = v.(string)
		}

		if v, ok := k8sGetOk(d, "password"); ok {
			opts.KubeOverrides.AuthInfo.Password = v.(string)
		}

		if v, ok := k8sGetOk(d, "client_key"); ok {
			opts.KubeOverrides.AuthInfo.ClientKeyData = bytes.NewBufferString(v.(string)).Bytes()
		}

		if v, ok := k8sGetOk(d, "token"); ok {
			opts.KubeOverrides.AuthInfo.Token = v.(string)
		}

		if v, ok := k8sGetOk(d, "exec"); ok {
			exec := &clientcmdapi.ExecConfig{}
			if spec, ok := v.([]interface{})[0].(map[string]interface{}); ok {
				exec.InteractiveMode = clientcmdapi.IfAvailableExecInteractiveMode
				exec.APIVersion = spec["api_version"].(string)
				exec.Command = spec["command"].(string)
				exec.Args = expandStringSlice(spec["args"].([]interface{}))

				for kk, vv := range spec["env"].(map[string]interface{}) {
					exec.Env = append(exec.Env, clientcmdapi.ExecEnvVar{Name: kk, Value: vv.(string)})
				}
			} else {
				return nil, fmt.Errorf("failed to parse exec")
			}

			opts.KubeOverrides.AuthInfo.Exec = exec
		}
	}

	authToken, authTokenOk := d.GetOk("auth_token")
	switch authTokenOk {
	case true:
		opts.AuthToken = authToken.(string)
	case false:
		userName, userNameOk := d.GetOk("username")
		password, passwordOk := d.GetOk("password")

		if userNameOk && passwordOk {
			apiClient, err = apiclient.NewClient(&opts)
			if err != nil {
				return apiClient, err
			}

			closer, sc, err := apiClient.NewSessionClient()
			if err != nil {
				return apiClient, err
			}

			defer io.Close(closer)

			sessionOpts := session.SessionCreateRequest{
				Username: userName.(string),
				Password: password.(string),
			}

			resp, err := sc.Create(ctx, &sessionOpts)
			if err != nil {
				return apiClient, err
			}

			opts.AuthToken = resp.Token
		}
	}

	if v, ok := d.Get("core").(bool); ok && v {
		opts.ServerAddr = "kubernetes"
		opts.Core = true

		// HACK: `headless.StartLocalServer` manipulates this global variable
		// when starting the local server without checking it's length/contents
		// which leads to a panic if called multiple times. So, we need to
		// ensure we "reset" it before calling the method.
		if runtimeErrorHandlers == nil {
			runtimeErrorHandlers = runtime.ErrorHandlers
		} else {
			runtime.ErrorHandlers = runtimeErrorHandlers
		}

		err := headless.StartLocalServer(ctx, &opts, "", nil, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to start local server: %w", err)
		}
	}

	return apiclient.NewClient(&opts)
}

func kubernetesResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_HOST", ""),
				Description: "The hostname (in form of URI) of the Kubernetes API. Can be sourced from `KUBE_HOST`.",
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_USER", ""),
				Description: "The username to use for HTTP basic authentication when accessing the Kubernetes API. Can be sourced from `KUBE_USER`.",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_PASSWORD", ""),
				Description: "The password to use for HTTP basic authentication when accessing the Kubernetes API. Can be sourced from `KUBE_PASSWORD`.",
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_INSECURE", false),
				Description: "Whether server should be accessed without verifying the TLS certificate. Can be sourced from `KUBE_INSECURE`.",
			},
			"client_certificate": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_CLIENT_CERT_DATA", ""),
				Description: "PEM-encoded client certificate for TLS authentication. Can be sourced from `KUBE_CLIENT_CERT_DATA`.",
			},
			"client_key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_CLIENT_KEY_DATA", ""),
				Description: "PEM-encoded client certificate key for TLS authentication. Can be sourced from `KUBE_CLIENT_KEY_DATA`.",
			},
			"cluster_ca_certificate": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_CLUSTER_CA_CERT_DATA", ""),
				Description: "PEM-encoded root certificates bundle for TLS authentication. Can be sourced from `KUBE_CLUSTER_CA_CERT_DATA`.",
			},
			"config_context": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_CTX", ""),
				Description: "Context to choose from the config file. Can be sourced from `KUBE_CTX`.",
			},
			"config_context_auth_info": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_CTX_AUTH_INFO", ""),
				Description: "",
			},
			"config_context_cluster": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_CTX_CLUSTER", ""),
				Description: "",
			},
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_TOKEN", ""),
				Description: "Token to authenticate an service account. Can be sourced from `KUBE_TOKEN`.",
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

func k8sGetOk(d *schema.ResourceData, key string) (interface{}, bool) {
	var k8sPrefix = "kubernetes.0."
	value, ok := d.GetOk(k8sPrefix + key)

	// For boolean attributes the zero value is Ok
	switch value.(type) {
	case bool:
		// TODO: replace deprecated GetOkExists with SDK v2 equivalent
		// https://github.com/hashicorp/terraform-plugin-sdk/pull/350
		value, ok = d.GetOkExists(k8sPrefix + key)
	}

	// fix: DefaultFunc is not being triggered on TypeList
	s := kubernetesResource().Schema[key]
	if !ok && s.DefaultFunc != nil {
		value, _ = s.DefaultFunc()

		switch v := value.(type) {
		case string:
			ok = len(v) != 0
		case bool:
			ok = v
		}
	}

	return value, ok
}

func expandStringSlice(s []interface{}) []string {
	result := make([]string, len(s))

	for k, v := range s {
		// Handle the Terraform parser bug which turns empty strings in lists to nil.
		if v == nil {
			result[k] = ""
		} else {
			result[k] = v.(string)
		}
	}

	return result
}

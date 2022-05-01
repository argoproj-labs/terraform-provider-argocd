package argocd

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/session"
	"github.com/argoproj/argo-cd/v2/util/io"
	"github.com/argoproj/argo-cd/v2/util/localconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	apimachineryschema "k8s.io/apimachinery/pkg/runtime/schema"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var apiClientConnOpts apiclient.ClientOptions

// Used to handle concurrent access to ArgoCD common configuration
var tokenMutexConfiguration = &sync.RWMutex{}

// Used to handle concurrent access to ArgoCD clusters
var tokenMutexClusters = &sync.RWMutex{}

// Used to handle concurrent access to each ArgoCD project
var tokenMutexProjectMap = make(map[string]*sync.RWMutex, 0)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"server_addr": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARGOCD_SERVER", nil),
			},
			"auth_token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARGOCD_AUTH_TOKEN", nil),
				ConflictsWith: []string{
					"username",
					"password",
					"use_local_config",
					"config_path",
				},
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARGOCD_AUTH_USERNAME", nil),
				ConflictsWith: []string{
					"auth_token",
					"use_local_config",
					"config_path",
				},
				AtLeastOneOf: []string{
					"password",
					"auth_token",
					"use_local_config",
				},
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARGOCD_AUTH_PASSWORD", nil),
				ConflictsWith: []string{
					"auth_token",
					"use_local_config",
					"config_path",
				},
				AtLeastOneOf: []string{
					"username",
					"auth_token",
					"use_local_config",
				},
			},
			"cert_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"plain_text": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"context": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARGOCD_CONTEXT", nil),
			},
			"user_agent": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"grpc_web": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"use_local_config": {
				Type:     schema.TypeBool,
				Optional: true,
				ConflictsWith: []string{
					"username",
					"password",
					"auth_token",
				},
			},
			"config_path": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARGOCD_CONFIG_PATH", nil),
				ConflictsWith: []string{
					"username",
					"password",
					"auth_token",
				},
			},
			"grpc_web_root_path": {
				Type:     schema.TypeString,
				Optional: true,
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
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARGOCD_INSECURE", false),
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

func initApiClient(d *schema.ResourceData) (
	apiClient apiclient.Client,
	err error) {

	var opts apiclient.ClientOptions

	if v, ok := d.GetOk("server_addr"); ok {
		opts.ServerAddr = v.(string)
	}

	if v, ok := d.GetOk("use_local_config"); ok {
		if v.(bool) {
			if v, ok := d.GetOk("config_path"); ok {
				opts.ConfigPath = v.(string)
			} else {
				path, err := localconfig.DefaultLocalConfigPath()
				if err != nil {
					return nil, err
				}
				opts.ConfigPath = path
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
		opts.Headers = v.([]string)
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
		if v, ok := k8sGetOk(d, "host"); ok {
			// Server has to be the complete address of the kubernetes cluster (scheme://hostname:port), not just the hostname,
			// because `overrides` are processed too late to be taken into account by `defaultServerUrlFor()`.
			// This basically replicates what defaultServerUrlFor() does with config but for overrides,
			// see https://github.com/kubernetes/client-go/blob/v12.0.0/rest/url_utils.go#L85-L87
			hasCA := len(opts.KubeOverrides.ClusterInfo.CertificateAuthorityData) != 0
			hasCert := len(opts.KubeOverrides.AuthInfo.ClientCertificateData) != 0
			defaultTLS := hasCA || hasCert || opts.KubeOverrides.ClusterInfo.InsecureSkipTLSVerify
			host, _, err := rest.DefaultServerURL(v.(string), "", apimachineryschema.GroupVersion{}, defaultTLS)
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
				log.Printf("[ERROR] Failed to parse exec")
				return nil, fmt.Errorf("failed to parse exec")
			}
			opts.KubeOverrides.AuthInfo.Exec = exec
		}
	}

	// Export provider API client connections options for use in other spawned api clients
	apiClientConnOpts = opts

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
			resp, err := sc.Create(context.Background(), &sessionOpts)
			if err != nil {
				return apiClient, err
			}
			opts.AuthToken = resp.Token
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
				Description: "The hostname (in form of URI) of Kubernetes master.",
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_USER", ""),
				Description: "The username to use for HTTP basic authentication when accessing the Kubernetes master endpoint.",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_PASSWORD", ""),
				Description: "The password to use for HTTP basic authentication when accessing the Kubernetes master endpoint.",
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_INSECURE", false),
				Description: "Whether server should be accessed without verifying the TLS certificate.",
			},
			"client_certificate": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_CLIENT_CERT_DATA", ""),
				Description: "PEM-encoded client certificate for TLS authentication.",
			},
			"client_key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_CLIENT_KEY_DATA", ""),
				Description: "PEM-encoded client certificate key for TLS authentication.",
			},
			"cluster_ca_certificate": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_CLUSTER_CA_CERT_DATA", ""),
				Description: "PEM-encoded root certificates bundle for TLS authentication.",
			},
			"config_paths": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "A list of paths to kube config files. Can be set with KUBE_CONFIG_PATHS environment variable.",
			},
			"config_path": {
				Type:          schema.TypeString,
				Optional:      true,
				DefaultFunc:   schema.EnvDefaultFunc("KUBE_CONFIG_PATH", nil),
				Description:   "Path to the kube config file. Can be set with KUBE_CONFIG_PATH.",
				ConflictsWith: []string{"kubernetes.0.config_paths"},
			},
			"config_context": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_CTX", ""),
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
				Description: "Token to authenticate an service account",
			},
			"exec": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"api_version": {
							Type:     schema.TypeString,
							Required: true,
						},
						"command": {
							Type:     schema.TypeString,
							Required: true,
						},
						"env": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"args": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
				Description: "",
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
	result := make([]string, len(s), len(s))
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

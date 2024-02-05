package argocd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
)

func clusterSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Description: "Name of the cluster. If omitted, will use the server address.",
			Optional:    true,
			DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
				if k == "name" {
					name, nameOk := d.GetOk("name")
					server, serverOk := d.GetOk("server")
					// Actual value is same as 'server' but not explicitly set
					if nameOk && serverOk && name == server && oldValue == server && newValue == "" {
						return true
					}
				}
				return false
			},
		},
		"server": {
			Type:        schema.TypeString,
			Description: "Server is the API server URL of the Kubernetes cluster.",
			Optional:    true,
			DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
				return oldValue == strings.TrimRight(newValue, "/")
			},
		},
		"shard": {
			Type:        schema.TypeString,
			Description: "Optional shard number. Calculated on the fly by the application controller if not specified.",
			Optional:    true,
		},
		"namespaces": {
			Type:        schema.TypeList,
			Description: "List of namespaces which are accessible in that cluster. Cluster level resources would be ignored if namespace list is not empty.",
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"config": {
			Type:        schema.TypeList,
			Description: "Cluster information for connecting to a cluster.",
			Required:    true,
			MinItems:    1,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"aws_auth_config": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"cluster_name": {
									Type:        schema.TypeString,
									Description: "AWS cluster name.",
									Optional:    true,
								},
								"role_arn": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "IAM role ARN. If set then AWS IAM Authenticator assume a role to perform cluster operations instead of the default AWS credential provider chain.",
								},
							},
						},
					},
					"bearer_token": {
						Type:        schema.TypeString,
						Description: "Server requires Bearer authentication. The client will not attempt to use refresh tokens for an OAuth2 flow.",
						Optional:    true,
						Sensitive:   true,
					},
					"exec_provider_config": {
						Type:        schema.TypeList,
						Optional:    true,
						MaxItems:    1,
						Description: "Configuration for an exec provider used to call an external command to perform cluster authentication See: https://godoc.org/k8s.io/client-go/tools/clientcmd/api#ExecConfig.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"api_version": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "Preferred input version of the ExecInfo",
								},
								"args": {
									Type:        schema.TypeList,
									Optional:    true,
									Description: "Arguments to pass to the command when executing it",
									Sensitive:   true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
								"command": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "Command to execute",
								},
								"env": {
									Type:        schema.TypeMap,
									Optional:    true,
									Description: "Env defines additional environment variables to expose to the process. Passed as a map of strings",
									Sensitive:   true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
								"install_hint": {
									Type:        schema.TypeString,
									Description: "This text is shown to the user when the executable doesn't seem to be present",
									Optional:    true,
								},
							},
						},
					},
					"tls_client_config": {
						Type:        schema.TypeList,
						Description: "Settings to enable transport layer security when connecting to the cluster.",
						MaxItems:    1,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"ca_data": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "PEM-encoded bytes (typically read from a root certificates bundle).",
								},
								"cert_data": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "PEM-encoded bytes (typically read from a client certificate file).",
								},
								"insecure": {
									Type:        schema.TypeBool,
									Optional:    true,
									Description: "Whether server should be accessed without verifying the TLS certificate.",
								},
								"key_data": {
									Type:        schema.TypeString,
									Optional:    true,
									Sensitive:   true,
									Description: "PEM-encoded bytes (typically read from a client certificate key file).",
								},
								"server_name": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "Name to pass to the server for SNI and used in the client to check server certificates against. If empty, the hostname used to contact the server is used.",
								},
							},
						},
					},
					"username": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Username for servers that require Basic authentication.",
					},
					"password": {
						Type:        schema.TypeString,
						Description: "Password for servers that require Basic authentication.",
						Optional:    true,
						Sensitive:   true,
					},
				},
			},
		},
		"info": {
			Type:        schema.TypeList,
			Description: "Information about cluster cache and state.",
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"server_version": {
						Type:        schema.TypeString,
						Description: "Kubernetes version of the cluster.",
						Computed:    true,
					},
					"applications_count": {
						Type:        schema.TypeString,
						Description: "Number of applications managed by Argo CD on the cluster.",
						Computed:    true,
					},
					"connection_state": {
						Type:        schema.TypeList,
						Description: "Information about the connection to the cluster.",
						Computed:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"message": {
									Type:        schema.TypeString,
									Description: "Human readable information about the connection status.",
									Computed:    true,
								},
								"status": {
									Type:        schema.TypeString,
									Description: "Current status indicator for the connection.",
									Computed:    true,
								},
							},
						},
					},
				},
			},
		},
		"metadata": {
			Type:        schema.TypeList,
			Description: "Standard cluster secret's metadata. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata",
			Optional:    true,
			MaxItems:    2,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"annotations": {
						Type:         schema.TypeMap,
						Description:  "An unstructured key value map stored with the cluster secret that may be used to store arbitrary metadata. More info: http://kubernetes.io/docs/user-guide/annotations",
						Optional:     true,
						Elem:         &schema.Schema{Type: schema.TypeString},
						ValidateFunc: validateMetadataAnnotations,
					},
					"labels": {
						Type:         schema.TypeMap,
						Description:  "Map of string keys and values that can be used to organize and categorize (scope and select) the cluster secret. May match selectors of replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels",
						Optional:     true,
						Elem:         &schema.Schema{Type: schema.TypeString},
						ValidateFunc: validateMetadataLabels,
					},
				},
			},
		},
		"project": {
			Type:        schema.TypeString,
			Description: "Reference between project and cluster that allow you automatically to be added as item inside Destinations project entity. More info: https://argo-cd.readthedocs.io/en/stable/user-guide/projects/#project-scoped-repositories-and-clusters.",
			Optional:    true,
		},
	}
}

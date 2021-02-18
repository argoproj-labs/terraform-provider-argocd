package argocd

import (
	application "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Expand

func expandCluster(d *schema.ResourceData) (*application.Cluster, error) {
	var err error
	cluster := &application.Cluster{}
	if v, ok := d.GetOk("name"); ok {
		cluster.Name = v.(string)
	}
	if v, ok := d.GetOk("server"); ok {
		cluster.Server = v.(string)
	}
	if v, ok := d.GetOk("shard"); ok {
		cluster.Shard, err = convertStringToInt64Pointer(v.(string))
		if err != nil {
			return nil, err
		}
	}
	if v, ok := d.GetOk("namespaces"); ok {
		cluster.Namespaces = v.([]string)
	}
	if v, ok := d.GetOk("config"); ok {
		cluster.Config = expandClusterConfig(v.([]interface{})[0])
	}
	return cluster, err
}

func expandClusterConfig(config interface{}) (
	clusterConfig application.ClusterConfig) {
	c := config.(map[string]interface{})
	if _aws, ok := c["aws_auth_config"]; ok {
		aws := _aws.([]map[string]string)[0]
		for k, v := range aws {
			if k == "cluster_name" {
				clusterConfig.AWSAuthConfig.ClusterName = v
			}
			if k == "role_arn" {
				clusterConfig.AWSAuthConfig.RoleARN = v
			}
		}
	}
	if v, ok := c["bearer_token"]; ok {
		clusterConfig.BearerToken = v.(string)
	}
	if v, ok := c["username"]; ok {
		clusterConfig.Username = v.(string)
	}
	if v, ok := c["password"]; ok {
		clusterConfig.Password = v.(string)
	}
	if _tls, ok := c["tls_client_config"]; ok {
		clusterConfig.TLSClientConfig = application.TLSClientConfig{}
		tls := _tls.([]map[string]interface{})[0]
		for k, v := range tls {
			if k == "ca_data" {
				clusterConfig.TLSClientConfig.CAData = []byte(v.(string))
			}
			if k == "cert_data" {
				clusterConfig.TLSClientConfig.CertData = []byte(v.(string))
			}
			if k == "key_data" {
				clusterConfig.TLSClientConfig.KeyData = []byte(v.(string))
			}
			if k == "insecure" {
				clusterConfig.TLSClientConfig.Insecure = v.(bool)
			}
			if k == "server_name" {
				clusterConfig.TLSClientConfig.ServerName = v.(string)
			}
		}
	}
	if _epc, ok := c["exec_provider_config"]; ok {
		clusterConfig.ExecProviderConfig = &application.ExecProviderConfig{}
		epc := _epc.([]map[string]interface{})[0]
		for k, v := range epc {
			if k == "api_version" {
				clusterConfig.ExecProviderConfig.APIVersion = v.(string)
			}
			if k == "args" {
				clusterConfig.ExecProviderConfig.Args = v.([]string)
			}
			if k == "command" {
				clusterConfig.ExecProviderConfig.Command = v.(string)
			}
			if k == "install_hint" {
				clusterConfig.ExecProviderConfig.InstallHint = v.(string)
			}
			if k == "env" {
				clusterConfig.ExecProviderConfig.Env = v.(map[string]string)
			}
		}
	}
	return
}

// Flatten

func flattenCluster(cluster *application.Cluster, d *schema.ResourceData) error {
	r := map[string]interface{}{
		"name":       cluster.Name,
		"server":     cluster.Server,
		"shard":      *cluster.Shard,
		"namespaces": cluster.Namespaces,
		"info":       flattenClusterInfo(cluster.Info),
		"config":     flattenClusterConfig(cluster.Config),
	}
	for k, v := range r {
		if err := persistToState(k, v, d); err != nil {
			return err
		}
	}
	return nil
}

func flattenClusterInfo(info application.ClusterInfo) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"server_version":     info.ServerVersion,
			"applications_count": info.ApplicationsCount,
			"connection_state": []map[string]string{
				{
					"message": info.ConnectionState.Message,
					"status":  info.ConnectionState.Status,
				},
			},
		},
	}
}

func flattenClusterConfig(config application.ClusterConfig) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"aws_auth_config": []map[string]string{
				{
					"cluster_name": config.AWSAuthConfig.ClusterName,
					"role_arn":     config.AWSAuthConfig.RoleARN,
				},
			},
			"bearer_token":         config.BearerToken,
			"username":             config.Username,
			"password":             config.Password,
			"exec_provider_config": flattenClusterConfigExecProviderConfig(config.ExecProviderConfig),
			"tls_client_config":    flattenClusterConfigTLSClientConfig(config.TLSClientConfig),
		},
	}
}

func flattenClusterConfigTLSClientConfig(tls application.TLSClientConfig) []map[string]interface{} {
	return []map[string]interface{}{
		{},
	}
}

func flattenClusterConfigExecProviderConfig(epc *application.ExecProviderConfig) []map[string]interface{} {
	return []map[string]interface{}{
		{},
	}
}

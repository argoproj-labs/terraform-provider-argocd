package argocd

import (
	application "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	if ns, ok := d.GetOk("namespaces"); ok {
		for _, n := range ns.([]interface{}) {
			cluster.Namespaces = append(cluster.Namespaces, n.(string))
		}
	}
	if v, ok := d.GetOk("config"); ok {
		cluster.Config = expandClusterConfig(v.([]interface{})[0])
	}

	m := expandMetadata(d)
	cluster.Annotations = m.Annotations
	cluster.Labels = m.Labels

	if v, ok := d.GetOk("project"); ok {
		cluster.Project = v.(string)
	}

	return cluster, err
}

func expandClusterConfig(config interface{}) (
	clusterConfig application.ClusterConfig) {
	c := config.(map[string]interface{})
	if aws, ok := c["aws_auth_config"].([]interface{}); ok && len(aws) > 0 {
		clusterConfig.AWSAuthConfig = &application.AWSAuthConfig{}
		for k, v := range aws[0].(map[string]interface{}) {
			if k == "cluster_name" {
				clusterConfig.AWSAuthConfig.ClusterName = v.(string)
			}
			if k == "role_arn" {
				clusterConfig.AWSAuthConfig.RoleARN = v.(string)
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
	if tls, ok := c["tls_client_config"].([]interface{}); ok && len(tls) > 0 {
		clusterConfig.TLSClientConfig = application.TLSClientConfig{}
		for k, v := range tls[0].(map[string]interface{}) {
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
	if epc, ok := c["exec_provider_config"].([]interface{}); ok && len(epc) > 0 {
		clusterConfig.ExecProviderConfig = &application.ExecProviderConfig{}
		for k, v := range epc[0].(map[string]interface{}) {
			if k == "api_version" {
				clusterConfig.ExecProviderConfig.APIVersion = v.(string)
			}
			if k == "args" {
				argsI := v.([]interface{})
				for _, argI := range argsI {
					clusterConfig.ExecProviderConfig.Args = append(clusterConfig.ExecProviderConfig.Args, argI.(string))
				}
			}
			if k == "command" {
				clusterConfig.ExecProviderConfig.Command = v.(string)
			}
			if k == "install_hint" {
				clusterConfig.ExecProviderConfig.InstallHint = v.(string)
			}
			if k == "env" {
				clusterConfig.ExecProviderConfig.Env = make(map[string]string)
				envI := v.(map[string]interface{})
				for key, val := range envI {
					clusterConfig.ExecProviderConfig.Env[key] = val.(string)
				}
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
		"namespaces": cluster.Namespaces,
		"info":       flattenClusterInfo(cluster.Info),
		"config":     flattenClusterConfig(cluster.Config, d),
		"project":    cluster.Project,
	}
	if len(cluster.Annotations) != 0 || len(cluster.Labels) != 0 {
		// The generic flattenMetadata function can not be used since the Cluster
		// object does not actually have ObjectMeta, just label and annotation maps
		r["metadata"] = flattenClusterMetadata(cluster.Annotations, cluster.Labels)
	}
	if cluster.Shard != nil {
		r["shard"] = convertInt64PointerToString(cluster.Shard)
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
			"applications_count": convertInt64ToString(info.ApplicationsCount),
			"connection_state": []map[string]string{
				{
					"message": info.ConnectionState.Message,
					"status":  info.ConnectionState.Status,
				},
			},
		},
	}
}

func flattenClusterConfig(config application.ClusterConfig, d *schema.ResourceData) []map[string]interface{} {
	var scc application.ClusterConfig
	r := map[string]interface{}{
		"username":             config.Username,
		"exec_provider_config": flattenClusterConfigExecProviderConfig(config.ExecProviderConfig),
	}
	if stateClusterConfig, ok := d.GetOk("config"); ok {
		scc = expandClusterConfig(stateClusterConfig.([]interface{})[0])
		r["password"] = scc.Password
		r["bearer_token"] = scc.BearerToken
		r["tls_client_config"] = flattenClusterConfigTLSClientConfig(scc)
	}
	if config.AWSAuthConfig != nil {
		r["aws_auth_config"] = []map[string]string{
			{
				"cluster_name": config.AWSAuthConfig.ClusterName,
				"role_arn":     config.AWSAuthConfig.RoleARN,
			},
		}
	}
	return []map[string]interface{}{r}
}

func flattenClusterConfigTLSClientConfig(stateClusterConfig application.ClusterConfig) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"ca_data":     string(stateClusterConfig.CAData),
			"cert_data":   string(stateClusterConfig.CertData),
			"key_data":    string(stateClusterConfig.KeyData),
			"insecure":    stateClusterConfig.Insecure,
			"server_name": stateClusterConfig.ServerName,
		},
	}
}

func flattenClusterConfigExecProviderConfig(epc *application.ExecProviderConfig) (
	result []map[string]interface{}) {
	if epc != nil {
		result = []map[string]interface{}{
			{
				"api_version":  epc.APIVersion,
				"args":         epc.Args,
				"command":      epc.Command,
				"env":          epc.Env,
				"install_hint": epc.InstallHint,
			},
		}
	}
	return
}

func flattenClusterMetadata(annotations, labels map[string]string) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"annotations": annotations,
			"labels":      labels,
		},
	}
}

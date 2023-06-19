package argocd

import (
	"fmt"

	application "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandCluster(d *schema.ResourceData) (*application.Cluster, error) {
	cluster := &application.Cluster{}

	if v, ok := d.GetOk("name"); ok {
		cluster.Name = v.(string)
	}

	if v, ok := d.GetOk("server"); ok {
		cluster.Server = v.(string)
	}

	if v, ok := d.GetOk("shard"); ok {
		shard, err := convertStringToInt64Pointer(v.(string))
		if err != nil {
			return nil, err
		}

		cluster.Shard = shard
	}

	if ns, ok := d.GetOk("namespaces"); ok {
		for _, n := range ns.([]interface{}) {
			if n == nil {
				return nil, fmt.Errorf("namespaces: must contain non-empty strings")
			}

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

	return cluster, nil
}

func expandClusterConfig(config interface{}) application.ClusterConfig {
	clusterConfig := application.ClusterConfig{}

	c := config.(map[string]interface{})
	if aws, ok := c["aws_auth_config"].([]interface{}); ok && len(aws) > 0 {
		clusterConfig.AWSAuthConfig = &application.AWSAuthConfig{}

		for k, v := range aws[0].(map[string]interface{}) {
			switch k {
			case "cluster_name":
				clusterConfig.AWSAuthConfig.ClusterName = v.(string)
			case "role_arn":
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
			switch k {
			case "ca_data":
				clusterConfig.TLSClientConfig.CAData = []byte(v.(string))
			case "cert_data":
				clusterConfig.TLSClientConfig.CertData = []byte(v.(string))
			case "key_data":
				clusterConfig.TLSClientConfig.KeyData = []byte(v.(string))
			case "insecure":
				clusterConfig.TLSClientConfig.Insecure = v.(bool)
			case "server_name":
				clusterConfig.TLSClientConfig.ServerName = v.(string)
			}
		}
	}

	if epc, ok := c["exec_provider_config"].([]interface{}); ok && len(epc) > 0 {
		clusterConfig.ExecProviderConfig = &application.ExecProviderConfig{}

		for k, v := range epc[0].(map[string]interface{}) {
			switch k {
			case "api_version":
				clusterConfig.ExecProviderConfig.APIVersion = v.(string)
			case "args":
				argsI := v.([]interface{})
				for _, argI := range argsI {
					clusterConfig.ExecProviderConfig.Args = append(clusterConfig.ExecProviderConfig.Args, argI.(string))
				}
			case "command":
				clusterConfig.ExecProviderConfig.Command = v.(string)
			case "install_hint":
				clusterConfig.ExecProviderConfig.InstallHint = v.(string)
			case "env":
				clusterConfig.ExecProviderConfig.Env = make(map[string]string)

				envI := v.(map[string]interface{})
				for key, val := range envI {
					clusterConfig.ExecProviderConfig.Env[key] = val.(string)
				}
			}
		}
	}

	return clusterConfig
}

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
	r := map[string]interface{}{
		"username":             config.Username,
		"exec_provider_config": flattenClusterConfigExecProviderConfig(config.ExecProviderConfig, d),
		"tls_client_config":    flattenClusterConfigTLSClientConfig(config.TLSClientConfig, d),
	}

	if config.AWSAuthConfig != nil {
		r["aws_auth_config"] = []map[string]string{
			{
				"cluster_name": config.AWSAuthConfig.ClusterName,
				"role_arn":     config.AWSAuthConfig.RoleARN,
			},
		}
	}

	// ArgoCD API does not return these fields as they may contain
	// sensitive data. Thus, we can't track the state of these
	// attributes and load them from state instead.
	// See https://github.com/argoproj/argo-cd/blob/8840929187f4dd7b9d9fd908ea5085a006895507/server/cluster/cluster.go#L448-L466
	if bt, ok := d.GetOk("config.0.bearer_token"); ok {
		r["bearer_token"] = bt
	}

	if p, ok := d.GetOk("config.0.password"); ok {
		r["password"] = p
	}

	return []map[string]interface{}{r}
}

func flattenClusterConfigTLSClientConfig(tcc application.TLSClientConfig, d *schema.ResourceData) []map[string]interface{} {
	c := map[string]interface{}{
		"ca_data":     string(tcc.CAData),
		"cert_data":   string(tcc.CertData),
		"insecure":    tcc.Insecure,
		"server_name": tcc.ServerName,
	}

	// ArgoCD API does not return sensitive data. Thus, we can't track
	// the state of this attribute and load it from state instead.
	// See https://github.com/argoproj/argo-cd/blob/8840929187f4dd7b9d9fd908ea5085a006895507/server/cluster/cluster.go#L448-L466
	if kd, ok := d.GetOk("config.0.tls_client_config.0.key_data"); ok {
		c["key_data"] = kd
	}

	return []map[string]interface{}{c}
}

func flattenClusterConfigExecProviderConfig(epc *application.ExecProviderConfig, d *schema.ResourceData) []map[string]interface{} {
	if epc == nil {
		return nil
	}

	c := map[string]interface{}{
		"api_version":  epc.APIVersion,
		"command":      epc.Command,
		"install_hint": epc.InstallHint,
	}

	// ArgoCD API does not return these fields as they may contain
	// sensitive data. Thus, we can't track the state of these
	// attributes and load them from state instead.
	// See https://github.com/argoproj/argo-cd/blob/8840929187f4dd7b9d9fd908ea5085a006895507/server/cluster/cluster.go#L454-L461
	if args, ok := d.GetOk("config.0.exec_provider_config.0.args"); ok {
		c["args"] = args
	}

	if env, ok := d.GetOk("config.0.exec_provider_config.0.env"); ok {
		c["env"] = env
	}

	return []map[string]interface{}{c}
}

func flattenClusterMetadata(annotations, labels map[string]string) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"annotations": annotations,
			"labels":      labels,
		},
	}
}

package argocd

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func clusterSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Description: "Name of the cluster. If omitted, will use the server address",
			Optional:    true,
		},
		"server": {
			Type:        schema.TypeString,
			Description: "Server is the API server URL of the Kubernetes cluster",
			Optional:    true,
		},
		"shard": {
			Type:        schema.TypeString,
			Description: "Shard contains optional shard number. Calculated on the fly by the application controller if not specified.",
			Optional:    true,
		},
		"namespaces": {
			Type:        schema.TypeList,
			Description: "Holds list of namespaces which are accessible in that cluster. Cluster level resources would be ignored if namespace list is not empty.",
			Optional:    true,
			Elem:        schema.TypeString,
		},
		"info": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"server_version": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"applications_count": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"connection_state": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{},
						},
					},
				},
			},
		},
	}
}

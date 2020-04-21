package argocd

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

func projectSpecSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		MinItems:    1,
		MaxItems:    1,
		Description: "ArgoCD App project resource specs. Required attributes: destination, source_repos.",
		Required:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cluster_resource_whitelist": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"group": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"kind": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
				"description": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"destination": {
					Type:     schema.TypeSet,
					Required: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"server": {
								Type:     schema.TypeString,
								Required: true,
							},
							"namespace": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				"namespace_resource_blacklist": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"group": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"kind": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
				"orphaned_resources": {
					Type:     schema.TypeMap,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeBool},
					// TODO: add a validatefunc to ensure only warn is present
				},
				"role": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"description": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"groups": {
								Type:     schema.TypeSet,
								Set:      schema.HashString,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"name": {
								Type:     schema.TypeString,
								Required: true,
							},
							"jwt_token": {
								Type:     schema.TypeList,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"iat": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"exp": {
											Type:     schema.TypeString,
											Optional: true,
										},
									},
								},
							},
							"policies": {
								Type:     schema.TypeSet,
								Set:      schema.HashString,
								Required: true,
								// TODO: add a validatefunc
								Elem: &schema.Schema{Type: schema.TypeString},
							},
						},
					},
				},
				"source_repos": {
					Type:     schema.TypeList,
					Required: true,
					// TODO: add a validatefunc
					Elem: &schema.Schema{Type: schema.TypeString},
				},
				"sync_window": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"applications": {
								Type:     schema.TypeList,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"clusters": {
								Type:     schema.TypeList,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"duration": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"kind": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"manual_sync": {
								Type:     schema.TypeBool,
								Optional: true,
							},
							"namespaces": {
								Type:     schema.TypeList,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"schedule": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
			},
		},
	}
}

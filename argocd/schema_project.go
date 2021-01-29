package argocd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func projectSpecSchemaV0() *schema.Schema {
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
								Type:         schema.TypeString,
								ValidateFunc: validateGroupName,
								Optional:     true,
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
								Optional: true,
							},
							"namespace": {
								Type:     schema.TypeString,
								Required: true,
							},
							"name": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Name of the destination cluster which can be used instead of server.",
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
								Type:     schema.TypeList,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"name": {
								Type:         schema.TypeString,
								ValidateFunc: validateRoleName,
								Required:     true,
							},
							"policies": {
								Type:     schema.TypeList,
								Required: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
						},
					},
				},
				"source_repos": {
					Type:     schema.TypeList,
					Required: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
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
								Type:         schema.TypeString,
								ValidateFunc: validateSyncWindowDuration,
								Optional:     true,
							},
							"kind": {
								Type:         schema.TypeString,
								ValidateFunc: validateSyncWindowKind,
								Optional:     true,
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
								Type:         schema.TypeString,
								ValidateFunc: validateSyncWindowSchedule,
								Optional:     true,
							},
						},
					},
				},
			},
		},
	}
}

func projectSpecSchemaV1() *schema.Schema {
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
								Type:         schema.TypeString,
								ValidateFunc: validateGroupName,
								Optional:     true,
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
								Optional: true,
							},
							"namespace": {
								Type:     schema.TypeString,
								Required: true,
							},
							"name": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Name of the destination cluster which can be used instead of server.",
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
					Type:     schema.TypeSet,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"warn": {
								Type:     schema.TypeBool,
								Optional: true,
							},
							"ignore": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"group": {
											Type:         schema.TypeString,
											ValidateFunc: validateGroupName,
											Optional:     true,
										},
										"kind": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"name": {
											Type:     schema.TypeString,
											Optional: true,
										},
									},
								},
							},
						},
					},
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
								Type:     schema.TypeList,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"name": {
								Type:         schema.TypeString,
								ValidateFunc: validateRoleName,
								Required:     true,
							},
							"policies": {
								Type:     schema.TypeList,
								Required: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
						},
					},
				},
				"source_repos": {
					Type:     schema.TypeList,
					Required: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"signature_keys": {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
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
								Type:         schema.TypeString,
								ValidateFunc: validateSyncWindowDuration,
								Optional:     true,
							},
							"kind": {
								Type:         schema.TypeString,
								ValidateFunc: validateSyncWindowKind,
								Optional:     true,
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
								Type:         schema.TypeString,
								ValidateFunc: validateSyncWindowSchedule,
								Optional:     true,
							},
						},
					},
				},
			},
		},
	}
}

func resourceArgoCDProjectV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"metadata": metadataSchema("appprojects.argoproj.io"),
			"spec":     projectSpecSchemaV0(),
		},
	}
}

func resourceArgoCDProjectStateUpgradeV0(rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	orphanedResources := rawState["spec"].([]map[string]interface{})[0]["orphaned_resources"]

	switch orphanedResources.(type) {

	// <= v0.4.8
	case map[string]bool:
		warn := orphanedResources.(map[string]bool)["warn"]
		newOrphanedResources := schema.NewSet(
			schema.HashResource(&schema.Resource{
				Schema: map[string]*schema.Schema{
					"warn": {
						Type:     schema.TypeBool,
						Optional: true,
					},
					"ignore": {
						Type:     schema.TypeSet,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"group": {
									Type:         schema.TypeString,
									ValidateFunc: validateGroupName,
									Optional:     true,
								},
								"kind": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"name": {
									Type:     schema.TypeString,
									Optional: true,
								},
							},
						},
					},
				},
			}),
			[]interface{}{map[string]interface{}{"warn": warn}},
		)
		rawState["spec"].([]map[string]interface{})[0]["orphaned_resources"] = newOrphanedResources

	// >= v0.5.0 <= v1.1.0
	case *schema.Set:
	default:
		return nil, fmt.Errorf("error during state migration v0 to v1, unsupported type for 'orphaned_resources': %s", orphanedResources)
	}
	return rawState, nil
}

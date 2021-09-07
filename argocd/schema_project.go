package argocd

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				"cluster_resource_blacklist": {
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
				"namespace_resource_whitelist": {
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
				"cluster_resource_blacklist": {
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
				"namespace_resource_whitelist": {
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

func projectSpecSchemaV2() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		MinItems:    1,
		MaxItems:    1,
		Description: "ArgoCD App project resource specs. Required attributes: destination, source_repos.",
		Required:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cluster_resource_blacklist": {
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
				"namespace_resource_whitelist": {
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
					Type:     schema.TypeList,
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

func resourceArgoCDProjectV1() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"metadata": metadataSchema("appprojects.argoproj.io"),
			"spec":     projectSpecSchemaV1(),
		},
	}
}

func resourceArgoCDProjectStateUpgradeV0(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	_spec := rawState["spec"]
	if len(_spec.([]interface{})) > 0 {
		spec := _spec.([]interface{})[0]
		if orphanedResources, ok := spec.(map[string]interface{})["orphaned_resources"]; ok {
			switch orphanedResources.(type) {
			// <= v0.4.8 with nil orphaned_resources map
			case map[string]interface{}:
				warn := orphanedResources.(map[string]interface{})["warn"]
				newOrphanedResources := []interface{}{map[string]interface{}{"warn": warn}}
				rawState["spec"].([]interface{})[0].(map[string]interface{})["orphaned_resources"] = newOrphanedResources

			// <= v0.4.8 with non-nil orphaned_resources map
			case map[string]bool:
				warn := orphanedResources.(map[string]bool)["warn"]
				newOrphanedResources := []interface{}{map[string]bool{"warn": warn}}
				rawState["spec"].([]interface{})[0].(map[string]interface{})["orphaned_resources"] = newOrphanedResources

			// >= v0.5.0 <= v1.1.2 (should not happen)
			case *schema.Set:
				return nil, fmt.Errorf("error during state migration v0 to v1, unsupported type for 'orphaned_resources': %s", reflect.TypeOf(orphanedResources))
			default:
				return nil, fmt.Errorf("error during state migration v0 to v1, unsupported type for 'orphaned_resources': %s", reflect.TypeOf(orphanedResources))
			}
		}
	}
	return rawState, nil
}

func resourceArgoCDProjectStateUpgradeV1(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	_spec := rawState["spec"]
	if len(_spec.([]interface{})) > 0 {
		spec := _spec.([]interface{})[0]
		if orphanedResources, ok := spec.(map[string]interface{})["orphaned_resources"]; ok {
			switch orphanedResources.(type) {
			case []interface{}:
			// >= v0.5.0 <= v1.1.2
			case *schema.Set:
				rawState["spec"].([]interface{})[0].(map[string]interface{})["orphaned_resources"] = orphanedResources.(*schema.Set).List()
			default:
				return nil, fmt.Errorf("error during state migration v1 to v2, unsupported type for 'orphaned_resources': %s", reflect.TypeOf(orphanedResources))
			}
		}
	}
	return rawState, nil
}

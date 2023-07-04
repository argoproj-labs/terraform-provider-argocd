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
		Description: "ArgoCD AppProject spec.",
		Required:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cluster_resource_blacklist": {
					Type:        schema.TypeSet,
					Description: "Blacklisted cluster level resources.",
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"group": {
								Type:         schema.TypeString,
								Description:  "The Kubernetes resource Group to match for.",
								ValidateFunc: validateGroupName,
								Optional:     true,
							},
							"kind": {
								Type:        schema.TypeString,
								Description: "The Kubernetes resource Kind to match for.",
								Optional:    true,
							},
						},
					},
				},
				"cluster_resource_whitelist": {
					Type:        schema.TypeSet,
					Description: "Whitelisted cluster level resources.",
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"group": {
								Type:         schema.TypeString,
								Description:  "The Kubernetes resource Group to match for.",
								ValidateFunc: validateGroupName,
								Optional:     true,
							},
							"kind": {
								Type:        schema.TypeString,
								Description: "The Kubernetes resource Kind to match for.",
								Optional:    true,
							},
						},
					},
				},
				"description": {
					Type:        schema.TypeString,
					Description: "Project description.",
					Optional:    true,
				},
				"destination": {
					Type:        schema.TypeSet,
					Description: "Destinations available for deployment.",
					Required:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"server": {
								Type:        schema.TypeString,
								Description: "URL of the target cluster and must be set to the Kubernetes control plane API.",
								Optional:    true,
							},
							"namespace": {
								Type:        schema.TypeString,
								Description: "Target namespace for applications' resources.",
								Required:    true,
							},
							"name": {
								Type:        schema.TypeString,
								Description: "Name of the destination cluster which can be used instead of server.",
								Optional:    true,
							},
						},
					},
				},
				"namespace_resource_blacklist": {
					Type:        schema.TypeSet,
					Description: "Blacklisted namespace level resources.",
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"group": {
								Type:        schema.TypeString,
								Description: "The Kubernetes resource Group to match for.",
								Optional:    true,
							},
							"kind": {
								Type:        schema.TypeString,
								Description: "The Kubernetes resource Kind to match for.",
								Optional:    true,
							},
						},
					},
				},
				"namespace_resource_whitelist": {
					Type:        schema.TypeSet,
					Description: "Whitelisted namespace level resources.",
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"group": {
								Type:        schema.TypeString,
								Description: "The Kubernetes resource Group to match for.",
								Optional:    true,
							},
							"kind": {
								Type:        schema.TypeString,
								Description: "The Kubernetes resource Kind to match for.",
								Optional:    true,
							},
						},
					},
				},
				"orphaned_resources": {
					Type:        schema.TypeList,
					Description: "Settings specifying if controller should monitor orphaned resources of apps in this project.",
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"warn": {
								Type:        schema.TypeBool,
								Description: "Whether a warning condition should be created for apps which have orphaned resources.",
								Optional:    true,
							},
							"ignore": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"group": {
											Type:         schema.TypeString,
											Description:  "The Kubernetes resource Group to match for.",
											ValidateFunc: validateGroupName,
											Optional:     true,
										},
										"kind": {
											Type:        schema.TypeString,
											Description: "The Kubernetes resource Kind to match for.",
											Optional:    true,
										},
										"name": {
											Type:        schema.TypeString,
											Description: "The Kubernetes resource name to match for.",
											Optional:    true,
										},
									},
								},
							},
						},
					},
				},
				"role": {
					Type:        schema.TypeList,
					Description: "User defined RBAC roles associated with this project.",
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"description": {
								Type:        schema.TypeString,
								Description: "Description of the role.",
								Optional:    true,
							},
							"groups": {
								Type:        schema.TypeList,
								Description: "List of OIDC group claims bound to this role.",
								Optional:    true,
								Elem:        &schema.Schema{Type: schema.TypeString},
							},
							"name": {
								Type:         schema.TypeString,
								Description:  "Name of the role.",
								ValidateFunc: validateRoleName,
								Required:     true,
							},
							"policies": {
								Type:        schema.TypeList,
								Description: "List of casbin formatted strings that define access policies for the role in the project. For more information, see the [ArgoCD RBAC reference](https://argoproj.github.io/argo-cd/operator-manual/rbac/#rbac-permission-structure).",
								Required:    true,
								Elem:        &schema.Schema{Type: schema.TypeString},
							},
						},
					},
				},
				"source_repos": {
					Type:        schema.TypeList,
					Description: "List of repository URLs which can be used for deployment. Can be set to `[\"*\"]` to allow all configured repositories configured in ArgoCD.",
					Required:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"source_namespaces": {
					Type:        schema.TypeSet,
					Description: "List of namespaces that application resources are allowed to be created in.",
					Set:         schema.HashString,
					Optional:    true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"signature_keys": {
					Type:        schema.TypeList,
					Description: "List of PGP key IDs that commits in Git must be signed with in order to be allowed for sync.",
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"sync_window": {
					Type:        schema.TypeList,
					Description: "Settings controlling when syncs can be run for apps in this project.",
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"applications": {
								Type:        schema.TypeList,
								Description: "List of applications that the window will apply to.",
								Optional:    true,
								Elem:        &schema.Schema{Type: schema.TypeString},
							},
							"clusters": {
								Type:        schema.TypeList,
								Description: "List of clusters that the window will apply to.",
								Optional:    true,
								Elem:        &schema.Schema{Type: schema.TypeString},
							},
							"duration": {
								Type:         schema.TypeString,
								Description:  "Amount of time the sync window will be open.",
								ValidateFunc: validateSyncWindowDuration,
								Optional:     true,
							},
							"kind": {
								Type:         schema.TypeString,
								Description:  "Defines if the window allows or blocks syncs, allowed values are `allow` or `deny`.",
								ValidateFunc: validateSyncWindowKind,
								Optional:     true,
							},
							"manual_sync": {
								Type:        schema.TypeBool,
								Description: "Enables manual syncs when they would otherwise be blocked.",
								Optional:    true,
							},
							"namespaces": {
								Type:        schema.TypeList,
								Description: "List of namespaces that the window will apply to.",
								Optional:    true,
								Elem:        &schema.Schema{Type: schema.TypeString},
							},
							"schedule": {
								Type:         schema.TypeString,
								Description:  "Time the window will begin, specified in cron format.",
								ValidateFunc: validateSyncWindowSchedule,
								Optional:     true,
							},
							"timezone": {
								Type:         schema.TypeString,
								Description:  "Timezone that the schedule will be evaluated in.",
								ValidateFunc: validateSyncWindowTimezone,
								Optional:     true,
								Default:      "UTC",
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

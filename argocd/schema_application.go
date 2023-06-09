package argocd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func applicationSpecSchemaV0() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		MinItems:    1,
		MaxItems:    1,
		Description: "ArgoCD App application resource specs. Required attributes: destination, source.",
		Required:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"destination": {
					Type:     schema.TypeSet,
					Required: true,
					MinItems: 1,
					MaxItems: 1,
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
				"source": {
					Type:     schema.TypeList,
					Required: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"repo_url": {
								Type:     schema.TypeString,
								Required: true,
							},
							"path": {
								Type:     schema.TypeString,
								Optional: true,
								// TODO: add validator to test path is not absolute
							},
							"target_revision": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"chart": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"helm": {
								Type:     schema.TypeList,
								MaxItems: 1,
								MinItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"value_files": {
											Type:     schema.TypeList,
											Optional: true,
											Elem: &schema.Schema{
												Type: schema.TypeString,
											},
										},
										"values": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"parameter": {
											Type:     schema.TypeSet,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"name": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"value": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"force_string": {
														Type:        schema.TypeBool,
														Optional:    true,
														Description: "force_string determines whether to tell Helm to interpret booleans and numbers as strings",
													},
												},
											},
										},
										"release_name": {
											Type:        schema.TypeString,
											Description: "The Helm release name. If omitted it will use the application name",
											Optional:    true,
										},
									},
								},
							},
							"kustomize": {
								Type:     schema.TypeList,
								MaxItems: 1,
								MinItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"name_prefix": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"name_suffix": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"version": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"images": {
											Type:     schema.TypeSet,
											Optional: true,
											Elem: &schema.Schema{
												Type: schema.TypeString,
											},
										},
										"common_labels": {
											Type:         schema.TypeMap,
											Optional:     true,
											Elem:         &schema.Schema{Type: schema.TypeString},
											ValidateFunc: validateMetadataLabels,
										},
										"common_annotations": {
											Type:         schema.TypeMap,
											Optional:     true,
											Elem:         &schema.Schema{Type: schema.TypeString},
											ValidateFunc: validateMetadataAnnotations,
										},
									},
								},
							},
							"directory": {
								Type: schema.TypeList,
								DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
									// TODO: This can be removed once v2.6.7 is the oldest supported version
									// as the API returns `Zero` Directory objects now.

									// Avoid drift when recurse is explicitly set to false
									// Also ignore the directory node if both recurse & jsonnet are not set or ignored
									if k == "spec.0.source.0.directory.0.recurse" && oldValue == "" && newValue == "false" {
										return true
									}
									if k == "spec.0.source.0.directory.#" {
										_, hasRecurse := d.GetOk("spec.0.source.0.directory.0.recurse")
										_, hasJsonnet := d.GetOk("spec.0.source.0.directory.0.jsonnet")

										if !hasJsonnet && !hasRecurse {
											return true
										}
									}
									return false
								},
								MaxItems: 1,
								MinItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"recurse": {
											Type:     schema.TypeBool,
											Optional: true,
										},
										"jsonnet": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											MinItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"ext_var": {
														Type:     schema.TypeList,
														Optional: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"name": {
																	Type:     schema.TypeString,
																	Optional: true,
																},
																"value": {
																	Type:     schema.TypeString,
																	Optional: true,
																},
																"code": {
																	Type:     schema.TypeBool,
																	Optional: true,
																},
															},
														},
													},
													"tla": {
														Type:     schema.TypeSet,
														Optional: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"name": {
																	Type:     schema.TypeString,
																	Optional: true,
																},
																"value": {
																	Type:     schema.TypeString,
																	Optional: true,
																},
																"code": {
																	Type:     schema.TypeBool,
																	Optional: true,
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
							"plugin": {
								Type:     schema.TypeList,
								MaxItems: 1,
								MinItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"name": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"env": {
											Type:     schema.TypeSet,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"name": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"value": {
														Type:     schema.TypeString,
														Optional: true,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				"project": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The application project, defaults to 'default'",
					Default:     "default",
				},
				"sync_policy": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					MinItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"automated": {
								Type:     schema.TypeMap,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeBool},
							},
							"sync_options": {
								Type:     schema.TypeList,
								Optional: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
									// TODO: add a validator
								},
							},
							"retry": {
								Type:     schema.TypeList,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"limit": {
											Type:        schema.TypeString,
											Description: "Max number of allowed sync retries, as a string",
											Optional:    true,
										},
										"backoff": {
											Type:     schema.TypeMap,
											Optional: true,
											Elem:     &schema.Schema{Type: schema.TypeString},
										},
									},
								},
							},
						},
					},
				},
				"ignore_difference": {
					Type:     schema.TypeList,
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
							"name": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"namespace": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"json_pointers": {
								Type:     schema.TypeSet,
								Set:      schema.HashString,
								Optional: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
							"jq_path_expressions": {
								Type:     schema.TypeSet,
								Set:      schema.HashString,
								Optional: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
						},
					},
				},
				"info": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"value": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
				"revision_history_limit": {
					Type:     schema.TypeInt,
					Optional: true,
					Default:  10,
				},
			},
		},
	}
}

func applicationSpecSchemaV1() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		MinItems:    1,
		MaxItems:    1,
		Description: "ArgoCD App application resource specs. Required attributes: destination, source.",
		Required:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"destination": {
					Type:     schema.TypeSet,
					Required: true,
					MinItems: 1,
					MaxItems: 1,
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
				"source": {
					Type:     schema.TypeList,
					Required: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"repo_url": {
								Type:     schema.TypeString,
								Required: true,
							},
							"path": {
								Type:     schema.TypeString,
								Optional: true,
								// TODO: add validator to test path is not absolute
								Default: ".",
							},
							"target_revision": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"chart": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"helm": {
								Type:     schema.TypeList,
								MaxItems: 1,
								MinItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"value_files": {
											Type:     schema.TypeList,
											Optional: true,
											Elem: &schema.Schema{
												Type: schema.TypeString,
											},
										},
										"values": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"parameter": {
											Type:     schema.TypeSet,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"name": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"value": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"force_string": {
														Type:        schema.TypeBool,
														Optional:    true,
														Description: "force_string determines whether to tell Helm to interpret booleans and numbers as strings",
													},
												},
											},
										},
										"release_name": {
											Type:        schema.TypeString,
											Description: "The Helm release name. If omitted it will use the application name",
											Optional:    true,
										},
										"skip_crds": {
											Type:        schema.TypeBool,
											Description: "Helm installs custom resource definitions in the crds folder by default if they are not existing. If needed, it is possible to skip the CRD installation step with this flag",
											Optional:    true,
										},
									},
								},
							},
							"kustomize": {
								Type:     schema.TypeList,
								MaxItems: 1,
								MinItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"name_prefix": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"name_suffix": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"version": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"images": {
											Type:     schema.TypeSet,
											Optional: true,
											Elem: &schema.Schema{
												Type: schema.TypeString,
											},
										},
										"common_labels": {
											Type:         schema.TypeMap,
											Optional:     true,
											Elem:         &schema.Schema{Type: schema.TypeString},
											ValidateFunc: validateMetadataLabels,
										},
										"common_annotations": {
											Type:         schema.TypeMap,
											Optional:     true,
											Elem:         &schema.Schema{Type: schema.TypeString},
											ValidateFunc: validateMetadataAnnotations,
										},
									},
								},
							},
							"ksonnet": {
								Type:     schema.TypeList,
								MaxItems: 1,
								MinItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"environment": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"parameters": {
											Type:     schema.TypeSet,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"component": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"name": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"value": {
														Type:     schema.TypeString,
														Optional: true,
													},
												},
											},
										},
									},
								},
							},
							"directory": {
								Type: schema.TypeList,
								DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
									// Avoid drift when recurse is explicitly set to false
									// Also ignore the directory node if both recurse & jsonnet are not set or ignored
									if k == "spec.0.source.0.directory.0.recurse" && oldValue == "" && newValue == "false" {
										return true
									}
									if k == "spec.0.source.0.directory.#" {
										_, hasRecurse := d.GetOk("spec.0.source.0.directory.0.recurse")
										_, hasJsonnet := d.GetOk("spec.0.source.0.directory.0.jsonnet")

										if !hasJsonnet && !hasRecurse {
											return true
										}
									}
									return false
								},
								MaxItems: 1,
								MinItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"recurse": {
											Type:     schema.TypeBool,
											Optional: true,
										},
										"jsonnet": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											MinItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"ext_var": {
														Type:     schema.TypeList,
														Optional: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"name": {
																	Type:     schema.TypeString,
																	Optional: true,
																},
																"value": {
																	Type:     schema.TypeString,
																	Optional: true,
																},
																"code": {
																	Type:     schema.TypeBool,
																	Optional: true,
																},
															},
														},
													},
													"tla": {
														Type:     schema.TypeSet,
														Optional: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"name": {
																	Type:     schema.TypeString,
																	Optional: true,
																},
																"value": {
																	Type:     schema.TypeString,
																	Optional: true,
																},
																"code": {
																	Type:     schema.TypeBool,
																	Optional: true,
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
							"plugin": {
								Type:     schema.TypeList,
								MaxItems: 1,
								MinItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"name": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"env": {
											Type:     schema.TypeSet,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"name": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"value": {
														Type:     schema.TypeString,
														Optional: true,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				"project": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The application project, defaults to 'default'",
					Default:     "default",
				},
				"sync_policy": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					MinItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"automated": {
								Type:     schema.TypeMap,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeBool},
							},
							"sync_options": {
								Type:     schema.TypeList,
								Optional: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
									// TODO: add a validator
								},
							},
							"retry": {
								Type:     schema.TypeList,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"limit": {
											Type:        schema.TypeString,
											Description: "Max number of allowed sync retries, as a string",
											Optional:    true,
										},
										"backoff": {
											Type:     schema.TypeMap,
											Optional: true,
											Elem:     &schema.Schema{Type: schema.TypeString},
										},
									},
								},
							},
						},
					},
				},
				"ignore_difference": {
					Type:     schema.TypeList,
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
							"name": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"namespace": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"json_pointers": {
								Type:     schema.TypeSet,
								Set:      schema.HashString,
								Optional: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
							"jq_path_expressions": {
								Type:     schema.TypeSet,
								Set:      schema.HashString,
								Optional: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
						},
					},
				},
				"info": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"value": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
				"revision_history_limit": {
					Type:     schema.TypeInt,
					Optional: true,
					Default:  10,
				},
			},
		},
	}
}

func applicationSpecSchemaV2() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		MinItems:    1,
		MaxItems:    1,
		Description: "The application specification.",
		Required:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"destination": {
					Type:        schema.TypeSet,
					Description: "Reference to the Kubernetes server and namespace in which the application will be deployed.",
					Required:    true,
					MinItems:    1,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"server": {
								Type:        schema.TypeString,
								Description: "URL of the target cluster and must be set to the Kubernetes control plane API.",
								Optional:    true,
							},
							"namespace": {
								Type:        schema.TypeString,
								Description: "Target namespace for the application's resources. The namespace will only be set for namespace-scoped resources that have not set a value for .metadata.namespace.",
								Required:    true,
							},
							"name": {
								Type:        schema.TypeString,
								Description: "Name of the target cluster. Can be used instead of `server`.",
								Optional:    true,
							},
						},
					},
				},
				"source": {
					Type:        schema.TypeList,
					Description: "Location of the application's manifests or chart.",
					Required:    true,
					MinItems:    1,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"repo_url": {
								Type:        schema.TypeString,
								Description: "URL to the repository (Git or Helm) that contains the application manifests.",
								Required:    true,
							},
							"path": {
								Type:        schema.TypeString,
								Description: "Directory path within the repository. Only valid for applications sourced from Git.",
								Optional:    true,
								// TODO: add validator to test path is not absolute
								Default: ".",
							},
							"target_revision": {
								Type:        schema.TypeString,
								Description: "Revision of the source to sync the application to. In case of Git, this can be commit, tag, or branch. If omitted, will equal to HEAD. In case of Helm, this is a semver tag for the Chart's version.",
								Optional:    true,
							},
							"chart": {
								Type:        schema.TypeString,
								Description: "Helm chart name. Must be specified for applications sourced from a Helm repo.",
								Optional:    true,
							},
							"helm": {
								Type:        schema.TypeList,
								Description: "Helm specific options.",
								MaxItems:    1,
								MinItems:    1,
								Optional:    true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"value_files": {
											Type:        schema.TypeList,
											Description: "List of Helm value files to use when generating a template.",
											Optional:    true,
											Elem: &schema.Schema{
												Type: schema.TypeString,
											},
										},
										"values": {
											Type:        schema.TypeString,
											Description: "Helm values to be passed to helm template, typically defined as a block.",
											Optional:    true,
										},
										"parameter": {
											Type:        schema.TypeSet,
											Description: "Helm parameters which are passed to the helm template command upon manifest generation.",
											Optional:    true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"name": {
														Type:        schema.TypeString,
														Description: "Name of the Helm parameter.",
														Optional:    true,
													},
													"value": {
														Type:        schema.TypeString,
														Description: "Value of the Helm parameter.",
														Optional:    true,
													},
													"force_string": {
														Type:        schema.TypeBool,
														Optional:    true,
														Description: "Determines whether to tell Helm to interpret booleans and numbers as strings.",
													},
												},
											},
										},
										"release_name": {
											Type:        schema.TypeString,
											Description: "Helm release name. If omitted it will use the application name.",
											Optional:    true,
										},
										"skip_crds": {
											Type:        schema.TypeBool,
											Description: "Whether to skip custom resource definition installation step (Helm's [--skip-crds](https://helm.sh/docs/chart_best_practices/custom_resource_definitions/)).",
											Optional:    true,
										},
									},
								},
							},
							"kustomize": {
								Type:        schema.TypeList,
								Description: "Kustomize specific options.",
								MaxItems:    1,
								MinItems:    1,
								Optional:    true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"name_prefix": {
											Type:        schema.TypeString,
											Description: "Prefix appended to resources for Kustomize apps.",
											Optional:    true,
										},
										"name_suffix": {
											Type:        schema.TypeString,
											Description: "Suffix appended to resources for Kustomize apps.",
											Optional:    true,
										},
										"version": {
											Type:        schema.TypeString,
											Description: "Version of Kustomize to use for rendering manifests.",
											Optional:    true,
										},
										"images": {
											Type:        schema.TypeSet,
											Description: "List of Kustomize image override specifications.",
											Optional:    true,
											Elem: &schema.Schema{
												Type: schema.TypeString,
											},
										},
										"common_labels": {
											Type:         schema.TypeMap,
											Description:  "List of additional labels to add to rendered manifests.",
											Optional:     true,
											Elem:         &schema.Schema{Type: schema.TypeString},
											ValidateFunc: validateMetadataLabels,
										},
										"common_annotations": {
											Type:         schema.TypeMap,
											Description:  "List of additional annotations to add to rendered manifests.",
											Optional:     true,
											Elem:         &schema.Schema{Type: schema.TypeString},
											ValidateFunc: validateMetadataAnnotations,
										},
									},
								},
							},
							"directory": {
								Type:        schema.TypeList,
								Description: "Path/directory specific options.",
								DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
									// Avoid drift when recurse is explicitly set to false
									// Also ignore the directory node if both recurse & jsonnet are not set or ignored
									if k == "spec.0.source.0.directory.0.recurse" && oldValue == "" && newValue == "false" {
										return true
									}
									if k == "spec.0.source.0.directory.#" {
										_, hasRecurse := d.GetOk("spec.0.source.0.directory.0.recurse")
										_, hasJsonnet := d.GetOk("spec.0.source.0.directory.0.jsonnet")

										if !hasJsonnet && !hasRecurse {
											return true
										}
									}
									return false
								},
								MaxItems: 1,
								MinItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"recurse": {
											Type:        schema.TypeBool,
											Description: "Whether to scan a directory recursively for manifests.",
											Optional:    true,
										},
										"jsonnet": {
											Type:        schema.TypeList,
											Description: "Jsonnet specific options.",
											Optional:    true,
											MaxItems:    1,
											MinItems:    1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"ext_var": {
														Type:        schema.TypeList,
														Description: "List of Jsonnet External Variables.",
														Optional:    true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"name": {
																	Type:        schema.TypeString,
																	Description: "Name of Jsonnet variable.",
																	Optional:    true,
																},
																"value": {
																	Type:        schema.TypeString,
																	Description: "Value of Jsonnet variable.",
																	Optional:    true,
																},
																"code": {
																	Type:        schema.TypeBool,
																	Description: "Determines whether the variable should be evaluated as jsonnet code or treated as string.",
																	Optional:    true,
																},
															},
														},
													},
													"tla": {
														Type:        schema.TypeSet,
														Description: "List of Jsonnet Top-level Arguments",
														Optional:    true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"name": {
																	Type:        schema.TypeString,
																	Description: "Name of Jsonnet variable.",
																	Optional:    true,
																},
																"value": {
																	Type:        schema.TypeString,
																	Description: "Value of Jsonnet variable.",
																	Optional:    true,
																},
																"code": {
																	Type:        schema.TypeBool,
																	Description: "Determines whether the variable should be evaluated as jsonnet code or treated as string.",
																	Optional:    true,
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
							"plugin": {
								Type:        schema.TypeList,
								Description: "Config management plugin specific options.",
								MaxItems:    1,
								MinItems:    1,
								Optional:    true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"name": {
											Type:        schema.TypeString,
											Description: "Name of the plugin. Only set the plugin name if the plugin is defined in `argocd-cm`. If the plugin is defined as a sidecar, omit the name. The plugin will be automatically matched with the Application according to the plugin's discovery rules.",
											Optional:    true,
										},
										"env": {
											Type:        schema.TypeSet,
											Description: "Environment variables passed to the plugin.",
											Optional:    true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"name": {
														Type:        schema.TypeString,
														Description: "Name of the environment variable.",
														Optional:    true,
													},
													"value": {
														Type:        schema.TypeString,
														Description: "Value of the environment variable.",
														Optional:    true,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				"project": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The project the application belongs to. Defaults to `default`.",
					Default:     "default",
				},
				"sync_policy": {
					Type:        schema.TypeList,
					Description: "Controls when and how a sync will be performed.",
					Optional:    true,
					MaxItems:    1,
					MinItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"automated": {
								Type:        schema.TypeMap,
								Description: "Whether to automatically keep an application synced to the target revision.",
								Optional:    true,
								Elem:        &schema.Schema{Type: schema.TypeBool},
							},
							"sync_options": {
								Type:        schema.TypeList,
								Description: "List of sync options. More info: https://argo-cd.readthedocs.io/en/stable/user-guide/sync-options/.",
								Optional:    true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
									// TODO: add a validator
								},
							},
							"retry": {
								Type:        schema.TypeList,
								Description: "Controls failed sync retry behavior.",
								MaxItems:    1,
								Optional:    true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"limit": {
											Type:        schema.TypeString,
											Description: "Maximum number of attempts for retrying a failed sync. If set to 0, no retries will be performed.",
											Optional:    true,
										},
										"backoff": {
											Type:        schema.TypeMap,
											Description: "Controls how to backoff on subsequent retries of failed syncs.",
											Optional:    true,
											Elem:        &schema.Schema{Type: schema.TypeString},
										},
									},
								},
							},
						},
					},
				},
				"ignore_difference": {
					Type:        schema.TypeList,
					Description: "Resources and their fields which should be ignored during comparison. More info: https://argo-cd.readthedocs.io/en/stable/user-guide/diffing/#application-level-configuration.",
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
							"name": {
								Type:        schema.TypeString,
								Description: "The Kubernetes resource Name to match for.",
								Optional:    true,
							},
							"namespace": {
								Type:        schema.TypeString,
								Description: "The Kubernetes resource Namespace to match for.",
								Optional:    true,
							},
							"json_pointers": {
								Type:        schema.TypeSet,
								Description: "List of JSONPaths strings targeting the field(s) to ignore.",
								Set:         schema.HashString,
								Optional:    true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
							"jq_path_expressions": {
								Type:        schema.TypeSet,
								Description: "List of JQ path expression strings targeting the field(s) to ignore.",
								Set:         schema.HashString,
								Optional:    true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
						},
					},
				},
				"info": {
					Type:        schema.TypeSet,
					Description: "List of information (URLs, email addresses, and plain text) that relates to the application.",
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:        schema.TypeString,
								Description: "Name of the information.",
								Optional:    true,
							},
							"value": {
								Type:        schema.TypeString,
								Description: "Value of the information.",
								Optional:    true,
							},
						},
					},
				},
				"revision_history_limit": {
					Type:        schema.TypeInt,
					Description: "Limits the number of items kept in the application's revision history, which is used for informational purposes as well as for rollbacks to previous versions. This should only be changed in exceptional circumstances. Setting to zero will store no history. This will reduce storage used. Increasing will increase the space used to store the history, so we do not recommend increasing it. Default is 10.",
					Optional:    true,
					Default:     10,
				},
			},
		},
	}
}

func applicationSpecSchemaV3() *schema.Schema {
	// To support deploying applications to non-default namespaces (aka project
	// source namespaces), we need to do a state migration to ensure that the Id
	// on existing resources is updated to include the namespace.
	// For this to happen, we need to trigger a schema version upgrade on the
	// application resource however, the schema of the application `spec` has
	// changed from `v2`.
	return applicationSpecSchemaV2()
}

func applicationSpecSchemaV4(allOptional bool) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		MinItems:    1,
		MaxItems:    1,
		Description: "The application specification.",
		Optional:    allOptional,
		Required:    !allOptional,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"destination": {
					Type:        schema.TypeSet,
					Description: "Reference to the Kubernetes server and namespace in which the application will be deployed.",
					Optional:    allOptional,
					Required:    !allOptional,
					MinItems:    1,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"server": {
								Type:        schema.TypeString,
								Description: "URL of the target cluster and must be set to the Kubernetes control plane API.",
								Optional:    true,
							},
							"namespace": {
								Type:        schema.TypeString,
								Description: "Target namespace for the application's resources. The namespace will only be set for namespace-scoped resources that have not set a value for .metadata.namespace.",
								Optional:    true,
							},
							"name": {
								Type:        schema.TypeString,
								Description: "Name of the target cluster. Can be used instead of `server`.",
								Optional:    true,
							},
						},
					},
				},
				"source": {
					Type:        schema.TypeList,
					Description: "Location of the application's manifests or chart.",
					Optional:    allOptional,
					Required:    !allOptional,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"repo_url": {
								Type:        schema.TypeString,
								Description: "URL to the repository (Git or Helm) that contains the application manifests.",
								Optional:    allOptional,
								Required:    !allOptional,
							},
							"path": {
								Type:        schema.TypeString,
								Description: "Directory path within the repository. Only valid for applications sourced from Git.",
								Optional:    true,
								// TODO: add validator to test path is not absolute
								Default: ".",
							},
							"target_revision": {
								Type:        schema.TypeString,
								Description: "Revision of the source to sync the application to. In case of Git, this can be commit, tag, or branch. If omitted, will equal to HEAD. In case of Helm, this is a semver tag for the Chart's version.",
								Optional:    true,
							},
							"ref": {
								Type:        schema.TypeString,
								Description: "Reference to another `source` within defined sources. See associated documentation on [Helm value files from external Git repository](https://argo-cd.readthedocs.io/en/stable/user-guide/multiple_sources/#helm-value-files-from-external-git-repository) regarding combining `ref` with `path` and/or `chart`.",
								Optional:    true,
							},
							"chart": {
								Type:        schema.TypeString,
								Description: "Helm chart name. Must be specified for applications sourced from a Helm repo.",
								Optional:    true,
							},
							"helm": {
								Type:        schema.TypeList,
								Description: "Helm specific options.",
								MaxItems:    1,
								MinItems:    1,
								Optional:    true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"value_files": {
											Type:        schema.TypeList,
											Description: "List of Helm value files to use when generating a template.",
											Optional:    true,
											Elem: &schema.Schema{
												Type: schema.TypeString,
											},
										},
										"values": {
											Type:        schema.TypeString,
											Description: "Helm values to be passed to 'helm template', typically defined as a block.",
											Optional:    true,
										},
										"ignore_missing_value_files": {
											Type:        schema.TypeBool,
											Description: "Prevents 'helm template' from failing when `value_files` do not exist locally by not appending them to 'helm template --values'.",
											Optional:    true,
										},
										"parameter": {
											Type:        schema.TypeSet,
											Description: "Helm parameters which are passed to the helm template command upon manifest generation.",
											Optional:    true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"name": {
														Type:        schema.TypeString,
														Description: "Name of the Helm parameter.",
														Optional:    true,
													},
													"value": {
														Type:        schema.TypeString,
														Description: "Value of the Helm parameter.",
														Optional:    true,
													},
													"force_string": {
														Type:        schema.TypeBool,
														Optional:    true,
														Description: "Determines whether to tell Helm to interpret booleans and numbers as strings.",
													},
												},
											},
										},
										"file_parameter": {
											Type:        schema.TypeSet,
											Description: "File parameters for the helm template.",
											Optional:    true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"name": {
														Type:        schema.TypeString,
														Description: "Name of the Helm parameter.",
														Required:    true,
													},
													"path": {
														Type:        schema.TypeString,
														Description: "Path to the file containing the values for the Helm parameter.",
														Required:    true,
													},
												},
											},
										},
										"release_name": {
											Type:        schema.TypeString,
											Description: "Helm release name. If omitted it will use the application name.",
											Optional:    true,
										},
										"skip_crds": {
											Type:        schema.TypeBool,
											Description: "Whether to skip custom resource definition installation step (Helm's [--skip-crds](https://helm.sh/docs/chart_best_practices/custom_resource_definitions/)).",
											Optional:    true,
										},
										"pass_credentials": {
											Type:        schema.TypeBool,
											Description: "If true then adds '--pass-credentials' to Helm commands to pass credentials to all domains.",
											Optional:    true,
										},
									},
								},
							},
							"kustomize": {
								Type:        schema.TypeList,
								Description: "Kustomize specific options.",
								MaxItems:    1,
								MinItems:    1,
								Optional:    true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"name_prefix": {
											Type:        schema.TypeString,
											Description: "Prefix appended to resources for Kustomize apps.",
											Optional:    true,
										},
										"name_suffix": {
											Type:        schema.TypeString,
											Description: "Suffix appended to resources for Kustomize apps.",
											Optional:    true,
										},
										"version": {
											Type:        schema.TypeString,
											Description: "Version of Kustomize to use for rendering manifests.",
											Optional:    true,
										},
										"images": {
											Type:        schema.TypeSet,
											Description: "List of Kustomize image override specifications.",
											Optional:    true,
											Elem: &schema.Schema{
												Type: schema.TypeString,
											},
										},
										"common_labels": {
											Type:         schema.TypeMap,
											Description:  "List of additional labels to add to rendered manifests.",
											Optional:     true,
											Elem:         &schema.Schema{Type: schema.TypeString},
											ValidateFunc: validateMetadataLabels,
										},
										"common_annotations": {
											Type:         schema.TypeMap,
											Description:  "List of additional annotations to add to rendered manifests.",
											Optional:     true,
											Elem:         &schema.Schema{Type: schema.TypeString},
											ValidateFunc: validateMetadataAnnotations,
										},
									},
								},
							},
							"directory": {
								Type:        schema.TypeList,
								Description: "Path/directory specific options.",
								DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
									// Avoid drift when recurse is explicitly set to false
									// Also ignore the directory node if both recurse & jsonnet are not set or ignored
									if k == "spec.0.source.0.directory.0.recurse" && oldValue == "" && newValue == "false" {
										return true
									}
									if k == "spec.0.source.0.directory.#" {
										_, hasRecurse := d.GetOk("spec.0.source.0.directory.0.recurse")
										_, hasJsonnet := d.GetOk("spec.0.source.0.directory.0.jsonnet")

										if !hasJsonnet && !hasRecurse {
											return true
										}
									}
									return false
								},
								MaxItems: 1,
								MinItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"recurse": {
											Type:        schema.TypeBool,
											Description: "Whether to scan a directory recursively for manifests.",
											Optional:    true,
										},
										"jsonnet": {
											Type:        schema.TypeList,
											Description: "Jsonnet specific options.",
											Optional:    true,
											MaxItems:    1,
											MinItems:    1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"ext_var": {
														Type:        schema.TypeList,
														Description: "List of Jsonnet External Variables.",
														Optional:    true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"name": {
																	Type:        schema.TypeString,
																	Description: "Name of Jsonnet variable.",
																	Optional:    true,
																},
																"value": {
																	Type:        schema.TypeString,
																	Description: "Value of Jsonnet variable.",
																	Optional:    true,
																},
																"code": {
																	Type:        schema.TypeBool,
																	Description: "Determines whether the variable should be evaluated as jsonnet code or treated as string.",
																	Optional:    true,
																},
															},
														},
													},
													"tla": {
														Type:        schema.TypeSet,
														Description: "List of Jsonnet Top-level Arguments",
														Optional:    true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"name": {
																	Type:        schema.TypeString,
																	Description: "Name of Jsonnet variable.",
																	Optional:    true,
																},
																"value": {
																	Type:        schema.TypeString,
																	Description: "Value of Jsonnet variable.",
																	Optional:    true,
																},
																"code": {
																	Type:        schema.TypeBool,
																	Description: "Determines whether the variable should be evaluated as jsonnet code or treated as string.",
																	Optional:    true,
																},
															},
														},
													},
													"libs": {
														Type:        schema.TypeList,
														Description: "Additional library search dirs.",
														Optional:    true,
														Elem: &schema.Schema{
															Type: schema.TypeString,
														},
													},
												},
											},
										},
										"exclude": {
											Type:        schema.TypeString,
											Description: "Glob pattern to match paths against that should be explicitly excluded from being used during manifest generation. This takes precedence over the `include` field. To match multiple patterns, wrap the patterns in {} and separate them with commas. For example: '{config.yaml,env-use2/*}'",
											Optional:    true,
										},
										"include": {
											Type:        schema.TypeString,
											Description: "Glob pattern to match paths against that should be explicitly included during manifest generation. If this field is set, only matching manifests will be included. To match multiple patterns, wrap the patterns in {} and separate them with commas. For example: '{*.yml,*.yaml}'",
											Optional:    true,
										},
									},
								},
							},
							"plugin": {
								Type:        schema.TypeList,
								Description: "Config management plugin specific options.",
								MaxItems:    1,
								MinItems:    1,
								Optional:    true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"name": {
											Type:        schema.TypeString,
											Description: "Name of the plugin. Only set the plugin name if the plugin is defined in `argocd-cm`. If the plugin is defined as a sidecar, omit the name. The plugin will be automatically matched with the Application according to the plugin's discovery rules.",
											Optional:    true,
										},
										"env": {
											Type:        schema.TypeSet,
											Description: "Environment variables passed to the plugin.",
											Optional:    true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"name": {
														Type:        schema.TypeString,
														Description: "Name of the environment variable.",
														Optional:    true,
													},
													"value": {
														Type:        schema.TypeString,
														Description: "Value of the environment variable.",
														Optional:    true,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				"project": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The project the application belongs to. Defaults to `default`.",
					Default:     "default",
				},
				"sync_policy": {
					Type:        schema.TypeList,
					Description: "Controls when and how a sync will be performed.",
					Optional:    true,
					MaxItems:    1,
					MinItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"automated": {
								Type:        schema.TypeSet,
								Description: "Whether to automatically keep an application synced to the target revision.",
								MaxItems:    1,
								Optional:    true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"prune": {
											Type:        schema.TypeBool,
											Description: "Whether to delete resources from the cluster that are not found in the sources anymore as part of automated sync.",
											Optional:    true,
										},
										"self_heal": {
											Type:        schema.TypeBool,
											Description: "Whether to revert resources back to their desired state upon modification in the cluster.",
											Optional:    true,
										},
										"allow_empty": {
											Type:        schema.TypeBool,
											Description: "Allows apps have zero live resources.",
											Optional:    true,
										},
									},
								},
							},
							"sync_options": {
								Type:        schema.TypeList,
								Description: "List of sync options. More info: https://argo-cd.readthedocs.io/en/stable/user-guide/sync-options/.",
								Optional:    true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
									// TODO: add a validator
								},
							},
							"retry": {
								Type:        schema.TypeList,
								Description: "Controls failed sync retry behavior.",
								MaxItems:    1,
								Optional:    true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"limit": {
											Type:        schema.TypeString,
											Description: "Maximum number of attempts for retrying a failed sync. If set to 0, no retries will be performed.",
											Optional:    true,
										},
										"backoff": {
											Type:        schema.TypeSet,
											MaxItems:    1,
											Description: "Controls how to backoff on subsequent retries of failed syncs.",
											Optional:    true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"duration": {
														Type:        schema.TypeString,
														Description: "Duration is the amount to back off. Default unit is seconds, but could also be a duration (e.g. `2m`, `1h`), as a string.",
														Optional:    true,
													},
													"factor": {
														Type:        schema.TypeString,
														Description: "Factor to multiply the base duration after each failed retry.",
														Optional:    true,
													},
													"max_duration": {
														Type:        schema.TypeString,
														Description: "Maximum amount of time allowed for the backoff strategy. Default unit is seconds, but could also be a duration (e.g. `2m`, `1h`), as a string.",
														Optional:    true,
													},
												},
											},
										},
									},
								},
							},
							"managed_namespace_metadata": {
								Type:        schema.TypeList,
								MaxItems:    1,
								Description: "Controls metadata in the given namespace (if `CreateNamespace=true`).",
								Optional:    true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"annotations": {
											Type:         schema.TypeMap,
											Description:  "Annotations to apply to the namespace.",
											Optional:     true,
											Elem:         &schema.Schema{Type: schema.TypeString},
											ValidateFunc: validateMetadataAnnotations,
										},
										"labels": {
											Type:         schema.TypeMap,
											Description:  "Labels to apply to the namespace.",
											Optional:     true,
											Elem:         &schema.Schema{Type: schema.TypeString},
											ValidateFunc: validateMetadataLabels,
										},
									},
								},
							},
						},
					},
				},
				"ignore_difference": {
					Type:        schema.TypeList,
					Description: "Resources and their fields which should be ignored during comparison. More info: https://argo-cd.readthedocs.io/en/stable/user-guide/diffing/#application-level-configuration.",
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
							"name": {
								Type:        schema.TypeString,
								Description: "The Kubernetes resource Name to match for.",
								Optional:    true,
							},
							"namespace": {
								Type:        schema.TypeString,
								Description: "The Kubernetes resource Namespace to match for.",
								Optional:    true,
							},
							"json_pointers": {
								Type:        schema.TypeSet,
								Description: "List of JSONPaths strings targeting the field(s) to ignore.",
								Set:         schema.HashString,
								Optional:    true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
							"jq_path_expressions": {
								Type:        schema.TypeSet,
								Description: "List of JQ path expression strings targeting the field(s) to ignore.",
								Set:         schema.HashString,
								Optional:    true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
						},
					},
				},
				"info": {
					Type:        schema.TypeSet,
					Description: "List of information (URLs, email addresses, and plain text) that relates to the application.",
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:        schema.TypeString,
								Description: "Name of the information.",
								Optional:    true,
							},
							"value": {
								Type:        schema.TypeString,
								Description: "Value of the information.",
								Optional:    true,
							},
						},
					},
				},
				"revision_history_limit": {
					Type:        schema.TypeInt,
					Description: "Limits the number of items kept in the application's revision history, which is used for informational purposes as well as for rollbacks to previous versions. This should only be changed in exceptional circumstances. Setting to zero will store no history. This will reduce storage used. Increasing will increase the space used to store the history, so we do not recommend increasing it. Default is 10.",
					Optional:    true,
					Default:     10,
				},
			},
		},
	}
}

func resourceArgoCDApplicationV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"metadata": metadataSchema("appprojects.argoproj.io"),
			"spec":     applicationSpecSchemaV0(),
		},
	}
}

func resourceArgoCDApplicationV1() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"metadata": metadataSchema("appprojects.argoproj.io"),
			"spec":     applicationSpecSchemaV1(),
		},
	}
}

func resourceArgoCDApplicationV2() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"metadata": metadataSchema("appprojects.argoproj.io"),
			"spec":     applicationSpecSchemaV2(),
		},
	}
}

func resourceArgoCDApplicationV3() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"metadata": metadataSchema("appprojects.argoproj.io"),
			"spec":     applicationSpecSchemaV3(),
		},
	}
}

func applicationStatusSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Status information for the application. **Note**: this is not guaranteed to be up to date immediately after creating/updating an application unless `wait=true`.",
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"conditions": {
					Type:        schema.TypeList,
					Description: "List of currently observed application conditions.",
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"message": {
								Type:        schema.TypeString,
								Description: "Human-readable message indicating details about condition.",
								Computed:    true,
							},
							"last_transition_time": {
								Type:        schema.TypeString,
								Description: "The time the condition was last observed.",
								Computed:    true,
							},
							"type": {
								Type:        schema.TypeString,
								Description: "Application condition type.",
								Computed:    true,
							},
						},
					},
				},
				"health": {
					Type:        schema.TypeList,
					Description: "Application's current health status.",
					Computed:    true,
					Elem:        resourceApplicationHealthStatus(),
				},
				"operation_state": {
					Type:        schema.TypeList,
					Description: "Information about any ongoing operations, such as a sync.",
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"finished_at": {
								Type:        schema.TypeString,
								Description: "Time of operation completion.",
								Computed:    true,
							},
							"message": {
								Type:        schema.TypeString,
								Description: "Any pertinent messages when attempting to perform operation (typically errors).",
								Computed:    true,
							},
							"phase": {
								Type:        schema.TypeString,
								Description: "The current phase of the operation.",
								Computed:    true,
							},
							"retry_count": {
								Type:        schema.TypeString,
								Description: "Count of operation retries.",
								Computed:    true,
							},
							"started_at": {
								Type:        schema.TypeString,
								Description: "Time of operation start.",
								Computed:    true,
							},
						},
					},
				},
				"reconciled_at": {
					Type:        schema.TypeString,
					Description: "When the application state was reconciled using the latest git version.",
					Computed:    true,
				},
				"resources": {
					Type:        schema.TypeList,
					Description: "List of Kubernetes resources managed by this application.",
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"group": {
								Type:        schema.TypeString,
								Description: "The Kubernetes resource Group.",
								Computed:    true,
							},
							"health": {
								Type:        schema.TypeList,
								Description: "Resource health status.",
								Computed:    true,
								Elem:        resourceApplicationHealthStatus(),
							},
							"kind": {
								Type:        schema.TypeString,
								Description: "The Kubernetes resource Kind.",
								Computed:    true,
							},
							"hook": {
								Type:        schema.TypeBool,
								Description: "Indicates whether or not this resource has a hook annotation.",
								Computed:    true,
							},
							"name": {
								Type:        schema.TypeString,
								Description: "The Kubernetes resource Name.",
								Computed:    true,
							},
							"namespace": {
								Type:        schema.TypeString,
								Description: "The Kubernetes resource Namespace.",
								Computed:    true,
							},
							"requires_pruning": {
								Type:        schema.TypeBool,
								Description: "Indicates if the resources requires pruning or not.",
								Computed:    true,
							},
							"status": {
								Type:        schema.TypeString,
								Description: "Resource sync status.",
								Computed:    true,
							},
							"sync_wave": {
								Type:        schema.TypeString,
								Description: "Sync wave.",
								Computed:    true,
							},
							"version": {
								Type:        schema.TypeString,
								Description: "The Kubernetes resource Version.",
								Computed:    true,
							},
						},
					},
				},
				"summary": {
					Type:        schema.TypeList,
					Description: "List of URLs and container images used by this application.",
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"external_urls": {
								Type:        schema.TypeList,
								Description: "All external URLs of application child resources.",
								Computed:    true,
								Elem:        schema.TypeString,
							},
							"images": {
								Type:        schema.TypeList,
								Description: "All images of application child resources.",
								Computed:    true,
								Elem:        schema.TypeString,
							},
						},
					},
				},
				"sync": {
					Type:        schema.TypeList,
					Description: "Application's current sync status",
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"revision": {
								Type:        schema.TypeString,
								Description: "Information about the revision the comparison has been performed to.",
								Computed:    true,
							},
							"revisions": {
								Type:        schema.TypeList,
								Description: "Information about the revision(s) the comparison has been performed to.",
								Computed:    true,
								Elem:        schema.TypeString,
							},
							"status": {
								Type:        schema.TypeString,
								Description: "Sync state of the comparison.",
								Computed:    true,
							},
						},
					},
				},
			},
		},
	}
}

func resourceApplicationHealthStatus() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"message": {
				Type:        schema.TypeString,
				Description: "Human-readable informational message describing the health status.",
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: "Status code of the application or resource.",
				Computed:    true,
			},
		},
	}
}

func resourceArgoCDApplicationStateUpgradeV0(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	_spec, ok := rawState["spec"].([]interface{})
	if !ok || len(_spec) == 0 {
		return rawState, nil
	}

	spec := _spec[0].(map[string]interface{})

	_source, ok := spec["source"].([]interface{})
	if !ok || len(_source) == 0 {
		return rawState, nil
	}

	source := _source[0].(map[string]interface{})

	_helm, ok := source["helm"].([]interface{})
	if !ok || len(_helm) == 0 {
		return rawState, nil
	}

	helm := _helm[0].(map[string]interface{})
	helm["skip_crds"] = false

	return rawState, nil
}

func resourceArgoCDApplicationStateUpgradeV1(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	_spec, ok := rawState["spec"].([]interface{})
	if !ok || len(_spec) == 0 {
		return rawState, nil
	}

	spec := _spec[0].(map[string]interface{})

	_source, ok := spec["source"].([]interface{})
	if !ok || len(_source) == 0 {
		return rawState, nil
	}

	source := _source[0].(map[string]interface{})

	_ksonnet, ok := source["ksonnet"].([]interface{})
	if !ok || len(_ksonnet) == 0 {
		return rawState, nil
	}

	return nil, fmt.Errorf("error during state migration v1 to v2, 'ksonnet' support has been removed")
}

func resourceArgoCDApplicationStateUpgradeV2(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	_metadata, ok := rawState["metadata"].([]interface{})
	if !ok || len(_metadata) == 0 {
		return nil, fmt.Errorf("failed to read metadata during state migration v2 to v3")
	}

	metadata := _metadata[0].(map[string]interface{})
	rawState["id"] = fmt.Sprintf("%s:%s", metadata["name"].(string), metadata["namespace"].(string))

	return rawState, nil
}

func resourceArgoCDApplicationStateUpgradeV3(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	_spec, ok := rawState["spec"].([]interface{})
	if !ok || len(_spec) == 0 {
		return rawState, nil
	}

	spec := _spec[0].(map[string]interface{})

	_syncPolicy, ok := spec["sync_policy"].([]interface{})
	if !ok || len(_syncPolicy) == 0 {
		return rawState, nil
	}

	syncPolicy := _syncPolicy[0].(map[string]interface{})

	automated, ok := syncPolicy["automated"].(map[string]interface{})
	if ok {
		updated := make(map[string]interface{}, 0)
		for k, v := range automated {
			updated[k] = v
		}

		syncPolicy["automated"] = []map[string]interface{}{updated}
	}

	_retry, ok := syncPolicy["retry"].([]interface{})
	if !ok || len(_retry) == 0 {
		return rawState, nil
	}

	retry := _retry[0].(map[string]interface{})

	if backoff, ok := retry["backoff"].(map[string]interface{}); ok {
		updated := make(map[string]interface{}, 0)
		for k, v := range backoff {
			updated[k] = v
		}

		retry["backoff"] = []map[string]interface{}{updated}
	}

	return rawState, nil
}

package argocd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func applicationSpecSchema() *schema.Schema {
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
								// TODO: ArgoCD API ApplicationQuery does not return Directory attributes, investigate?
								// TODO: this provokes perpetual TF state drift as spec.0.source.0.directory cannot be read
								// TODO: the Directory attributes are to be used with care until a fix is made upstream
								DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
									return true
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
				},
			},
		},
	}
}

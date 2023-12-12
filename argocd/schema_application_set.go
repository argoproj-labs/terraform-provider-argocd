package argocd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const generatorSchemaLevel = 3

func applicationSetSpecSchemaV0() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		MinItems:    1,
		MaxItems:    1,
		Description: "ArgoCD application set resource spec.",
		Required:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"ignore_application_differences": {
					Type:        schema.TypeList,
					Description: "Application Set [ignoreApplicationDifferences](https://argo-cd.readthedocs.io/en/stable/operator-manual/applicationset/Controlling-Resource-Modification/#ignore-certain-changes-to-applications).",
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"json_pointers": {
								Type:        schema.TypeSet,
								Description: "Json pointers to ignore differences",
								Optional:    true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
							"jq_path_expressions": {
								Type:        schema.TypeSet,
								Description: "jq path to ignore differences",
								Optional:    true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
							"name": {
								Type:        schema.TypeString,
								Description: "name",
								Optional:    true,
							},
						},
					},
				},
				"generator": applicationSetGeneratorSchemaV0(),
				"go_template": {
					Type:        schema.TypeBool,
					Description: "Enable use of [Go Text Template](https://pkg.go.dev/text/template).",
					Optional:    true,
				},
				"strategy": {
					Type:        schema.TypeList,
					Description: "[Progressive Sync](https://argo-cd.readthedocs.io/en/stable/operator-manual/applicationset/Progressive-Syncs/) strategy",
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"type": {
								Type:        schema.TypeString,
								Description: "Type of progressive sync.",
								Required:    true,
							},
							"rolling_sync": {
								Type:        schema.TypeList,
								Description: "Update strategy allowing you to group Applications by labels present on the generated Application resources. When the ApplicationSet changes, the changes will be applied to each group of Application resources sequentially.",
								Optional:    true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"step": {
											Type:        schema.TypeList,
											Description: "Configuration used to define which applications to include in each stage of the rolling sync. All Applications in each group must become Healthy before the ApplicationSet controller will proceed to update the next group of Applications.",
											MinItems:    1,
											Required:    true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"match_expressions": matchExpressionsSchema(),
													"max_update": {
														Type:         schema.TypeString,
														Description:  "Maximum number of simultaneous Application updates in a group. Supports both integer and percentage string values (rounds down, but floored at 1 Application for >0%). Default is 100%, unbounded.",
														ValidateFunc: validateIntOrStringPercentage,
														Optional:     true,
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
				"sync_policy": {
					Type:        schema.TypeList,
					Description: "Application Set [sync policy](https://argo-cd.readthedocs.io/en/stable/operator-manual/applicationset/Controlling-Resource-Modification/).",
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"preserve_resources_on_deletion": {
								Type:        schema.TypeBool,
								Description: "Label selector used to narrow the scope of targeted clusters.",
								Optional:    true,
							},
							"applications_sync": {
								Type:        schema.TypeString,
								Description: "Represents the policy applied on the generated applications. Possible values are create-only, create-update, create-delete, and sync.",
								Optional:    true,
							},
						},
					},
				},
				"template": {
					Type:        schema.TypeList,
					Description: "Application set template. The template fields of the ApplicationSet spec are used to generate Argo CD Application resources.",
					Required:    true,
					MinItems:    1,
					MaxItems:    1,
					Elem:        applicationSetTemplateResource(false),
				},
			},
		},
	}
}

func applicationSetGeneratorSchemaV0() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Application set generators. Generators are responsible for generating parameters, which are then rendered into the template: fields of the ApplicationSet resource.",
		Required:    true,
		MinItems:    1,
		Elem:        generatorResourceV0(generatorSchemaLevel),
	}
}

func generatorResourceV0(level int) *schema.Resource {
	if level > 1 {
		return &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cluster_decision_resource": applicationSetClusterDecisionResourceGeneratorSchemaV0(),
				"clusters":                  applicationSetClustersGeneratorSchemaV0(),
				"git":                       applicationSetGitGeneratorSchemaV0(),
				"list":                      applicationSetListGeneratorSchemaV0(),
				"matrix":                    applicationSetMatrixGeneratorSchemaV0(level),
				"merge":                     applicationSetMergeGeneratorSchemaV0(level),
				"pull_request":              applicationSetPullRequestGeneratorSchemaV0(),
				"scm_provider":              applicationSetSCMProviderGeneratorSchemaV0(),
				"selector": {
					Type:        schema.TypeList,
					Description: "The Selector allows to post-filter based on generated values using the kubernetes common labelSelector format.",
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: labelSelectorSchema(),
					},
				},
			},
		}
	}

	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cluster_decision_resource": applicationSetClusterDecisionResourceGeneratorSchemaV0(),
			"clusters":                  applicationSetClustersGeneratorSchemaV0(),
			"git":                       applicationSetGitGeneratorSchemaV0(),
			"list":                      applicationSetListGeneratorSchemaV0(),
			"pull_request":              applicationSetPullRequestGeneratorSchemaV0(),
			"scm_provider":              applicationSetSCMProviderGeneratorSchemaV0(),
			"selector": {
				Type:        schema.TypeList,
				Description: "The Selector allows to post-filter based on generated values using the kubernetes common labelSelector format.",
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: labelSelectorSchema(),
				},
			},
		},
	}
}

func applicationSetClustersGeneratorSchemaV0() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "The [cluster generator](https://argo-cd.readthedocs.io/en/stable/operator-manual/applicationset/Generators-Cluster/) produces parameters based on the list of items found within the cluster secret.",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"selector": {
					Type:        schema.TypeList,
					Description: "Label selector used to narrow the scope of targeted clusters.",
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: labelSelectorSchema(),
					},
				},
				"template": {
					Type:        schema.TypeList,
					Description: "Generator template. Used to override the values of the spec-level template.",
					Optional:    true,
					MaxItems:    1,
					Elem:        applicationSetTemplateResource(true),
				},
				"values": {
					Type:        schema.TypeMap,
					Description: "Arbitrary string key-value pairs to pass to the template via the values field of the cluster generator.",
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"enabled": {
					Type:        schema.TypeBool,
					Description: "Boolean value defaulting to `true` to indicate that this block has been added thereby allowing all other attributes to be optional.",
					Required:    true,
					DefaultFunc: func() (interface{}, error) { return true, nil },
				},
			},
		},
	}
}

func applicationSetClusterDecisionResourceGeneratorSchemaV0() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "The [cluster decision resource](https://argo-cd.readthedocs.io/en/stable/operator-manual/applicationset/Generators-Cluster-Decision-Resource/) generates a list of Argo CD clusters.",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"config_map_ref": {
					Type:        schema.TypeString,
					Description: "ConfigMap with the duck type definitions needed to retrieve the data this includes apiVersion(group/version), kind, matchKey and validation settings.",
					Required:    true,
				},
				"name": {
					Type:        schema.TypeString,
					Description: "Resource name of the kind, group and version, defined in the `config_map_ref`.",
					Optional:    true,
				},
				"label_selector": {
					Type:        schema.TypeList,
					Description: "Label selector used to find the resource defined in the `config_map_ref`. Alternative to `name`.",
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: labelSelectorSchema(),
					},
				},
				"requeue_after_seconds": {
					Type:        schema.TypeString,
					Description: "How often to check for changes (in seconds). Default: 3min.",
					Optional:    true,
				},
				"template": {
					Type:        schema.TypeList,
					Description: "Generator template. Used to override the values of the spec-level template.",
					Optional:    true,
					MaxItems:    1,
					Elem:        applicationSetTemplateResource(true),
				},
				"values": {
					Type:        schema.TypeMap,
					Description: "Arbitrary string key-value pairs which are passed directly as parameters to the template.",
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			},
		},
	}
}

func applicationSetGitGeneratorSchemaV0() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "[Git generators](https://argo-cd.readthedocs.io/en/stable/operator-manual/applicationset/Generators-Git/) generates parameters using either the directory structure of a specified Git repository (directory generator), or, using the contents of JSON/YAML files found within a specified repository (file generator). ",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"directory": {
					Type:        schema.TypeList,
					Description: "List of directories in the source repository to use when template the Application..",
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"path": {
								Type:        schema.TypeString,
								Description: "Path in the repository.",
								Required:    true,
							},
							"exclude": {
								Type:        schema.TypeBool,
								Description: "Flag indicating whether or not the directory should be excluded when templating.",
								Optional:    true,
								Default:     false,
							},
						},
					},
				},
				"file": {
					Type:        schema.TypeList,
					Description: "List of files in the source repository to use when template the Application.",
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"path": {
								Type:        schema.TypeString,
								Description: "Path to the file in the repository.",
								Required:    true,
							},
						},
					},
				},
				"repo_url": {
					Type:        schema.TypeString,
					Description: "URL to the repository to use.",
					Required:    true,
				},
				"revision": {
					Type:        schema.TypeString,
					Description: "Revision of the source repository to use.",
					Optional:    true,
				},
				"path_param_prefix": {
					Type:        schema.TypeString,
					Description: "Prefix for all path-related parameter names.",
					Optional:    true,
				},
				"template": {
					Type:        schema.TypeList,
					Description: "Generator template. Used to override the values of the spec-level template.",
					Optional:    true,
					MaxItems:    1,
					Elem:        applicationSetTemplateResource(true),
				},
			},
		},
	}
}

func applicationSetListGeneratorSchemaV0() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "[List generators](https://argo-cd.readthedocs.io/en/stable/operator-manual/applicationset/Generators-List/) generate parameters based on an arbitrary list of key/value pairs (as long as the values are string values).",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"elements": {
					Type:        schema.TypeList,
					Description: "List of key/value pairs to pass as parameters into the template",
					Required:    true,
					Elem: &schema.Schema{
						Type: schema.TypeMap,
						Elem: &schema.Schema{Type: schema.TypeString},
					},
				},
				"template": {
					Type:        schema.TypeList,
					Description: "Generator template. Used to override the values of the spec-level template.",
					Optional:    true,
					MaxItems:    1,
					Elem:        applicationSetTemplateResource(true),
				},
			},
		},
	}
}

func applicationSetMatrixGeneratorSchemaV0(level int) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "[Matrix generators](https://argo-cd.readthedocs.io/en/stable/operator-manual/applicationset/Generators-Matrix/) combine the parameters generated by two child generators, iterating through every combination of each generator's generated parameters. Take note of the [restrictions](https://argo-cd.readthedocs.io/en/stable/operator-manual/applicationset/Generators-Matrix/#restrictions) regarding their usage - particularly regarding nesting matrix generators.",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"generator": {
					Type:        schema.TypeList,
					Description: "Child generator. Generators are responsible for generating parameters, which are then combined by the parent matrix generator into the template fields of the ApplicationSet resource.",
					Required:    true,
					MinItems:    2,
					MaxItems:    2,
					Elem:        generatorResourceV0(level - 1),
				},
				"template": {
					Type:        schema.TypeList,
					Description: "Generator template. Used to override the values of the spec-level template.",
					Optional:    true,
					MaxItems:    1,
					Elem:        applicationSetTemplateResource(true),
				},
			},
		},
	}
}

func applicationSetMergeGeneratorSchemaV0(level int) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "[Merge generators](https://argo-cd.readthedocs.io/en/stable/operator-manual/applicationset/Generators-Merge/) combine parameters produced by the base (first) generator with matching parameter sets produced by subsequent generators. Take note of the [restrictions](https://argo-cd.readthedocs.io/en/stable/operator-manual/applicationset/Generators-Merge/#restrictions) regarding their usage - particularly regarding nesting merge generators.",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"merge_keys": {
					Type:        schema.TypeList,
					Description: "Keys to merge into resulting parameter set.",
					Required:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"generator": {
					Type:        schema.TypeList,
					Description: "Child generator. Generators are responsible for generating parameters, which are then combined by the parent merge generator.",
					Required:    true,
					MinItems:    2,
					Elem:        generatorResourceV0(level - 1),
				},
				"template": {
					Type:        schema.TypeList,
					Description: "Generator template. Used to override the values of the spec-level template.",
					Optional:    true,
					MaxItems:    1,
					Elem:        applicationSetTemplateResource(true),
				},
			},
		},
	}
}

func applicationSetSCMProviderGeneratorSchemaV0() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "[SCM Provider generators](https://argo-cd.readthedocs.io/en/stable/operator-manual/applicationset/Generators-SCM-Provider/) uses the API of an SCMaaS provider to automatically discover repositories within an organization.",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"azure_devops": {
					Type:        schema.TypeList,
					Description: "Uses the Azure DevOps API to look up eligible repositories based on a team project within an Azure DevOps organization.",
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"all_branches": {
								Type:        schema.TypeBool,
								Description: "Scan all branches instead of just the default branch.",
								Optional:    true,
							},
							"access_token_ref": {
								Type:        schema.TypeList,
								Description: "The Personal Access Token (PAT) to use when connecting.",
								Optional:    true,
								MaxItems:    1,
								Elem:        secretRefResource(),
							},
							"api": {
								Type:        schema.TypeString,
								Description: "The URL to Azure DevOps. Defaults to https://dev.azure.com.",
								Optional:    true,
							},
							"organization": {
								Type:        schema.TypeString,
								Description: "Azure Devops organization. E.g. \"my-organization\".",
								Required:    true,
							},
							"team_project": {
								Type:        schema.TypeString,
								Description: "Azure Devops team project. E.g. \"my-team\".",
								Required:    true,
							},
						},
					},
				},
				"bitbucket_cloud": {
					Type:        schema.TypeList,
					Description: "Uses the Bitbucket API V2 to scan a workspace in bitbucket.org.",
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"all_branches": {
								Type:        schema.TypeBool,
								Description: "Scan all branches instead of just the default branch.",
								Optional:    true,
							},
							"app_password_ref": {
								Type:        schema.TypeList,
								Description: "The app password to use for the user. See: https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/.",
								Optional:    true,
								MaxItems:    1,
								Elem:        secretRefResource(),
							},
							"owner": {
								Type:        schema.TypeString,
								Description: "Bitbucket workspace to scan.",
								Required:    true,
							},
							"user": {
								Type:        schema.TypeString,
								Description: "Bitbucket user to use when authenticating. Should have a \"member\" role to be able to read all repositories and branches.",
								Required:    true,
							},
						},
					},
				},
				"bitbucket_server": {
					Type:        schema.TypeList,
					Description: "Use the Bitbucket Server API (1.0) to scan repos in a project.",
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"all_branches": {
								Type:        schema.TypeBool,
								Description: "Scan all branches instead of just the default branch.",
								Optional:    true,
							},
							"api": {
								Type:        schema.TypeString,
								Description: "The Bitbucket REST API URL to talk to e.g. https://bitbucket.org/rest.",
								Required:    true,
							},
							"basic_auth": {
								Type:        schema.TypeList,
								Description: "Credentials for Basic auth.",
								Optional:    true,
								MaxItems:    1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"username": {
											Type:        schema.TypeString,
											Description: "Username for Basic auth.",
											Optional:    true,
										},
										"password_ref": {
											Type:        schema.TypeList,
											Description: "Password (or personal access token) reference.",
											Optional:    true,
											MaxItems:    1,
											Elem:        secretRefResource(),
										},
									},
								},
							},
							"project": {
								Type:        schema.TypeString,
								Description: "Project to scan.",
								Required:    true,
							},
						},
					},
				},
				"clone_protocol": {
					Type:        schema.TypeString,
					Description: "Which protocol to use for the SCM URL. Default is provider-specific but ssh if possible. Not all providers necessarily support all protocols.",
					Optional:    true,
				},
				"filter": {
					Type:        schema.TypeList,
					Description: "Filters for which repos should be considered.",
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"branch_match": {
								Type:        schema.TypeString,
								Description: "A regex which must match the branch name.",
								Optional:    true,
							},
							"label_match": {
								Type:        schema.TypeString,
								Description: "A regex which must match at least one label.",
								Optional:    true,
							},
							"paths_do_not_exist": {
								Type:        schema.TypeList,
								Description: "An array of paths, all of which must not exist.",
								Optional:    true,
								Elem:        &schema.Schema{Type: schema.TypeString},
							},
							"paths_exist": {
								Type:        schema.TypeList,
								Description: "An array of paths, all of which must exist.",
								Optional:    true,
								Elem:        &schema.Schema{Type: schema.TypeString},
							},
							"repository_match": {
								Type:        schema.TypeString,
								Description: "A regex for repo names.",
								Optional:    true,
							},
						},
					},
				},
				"gitea": {
					Type:        schema.TypeList,
					Description: "Gitea mode uses the Gitea API to scan organizations in your instance.",
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"all_branches": {
								Type:        schema.TypeBool,
								Description: "Scan all branches instead of just the default branch.",
								Optional:    true,
							},
							"api": {
								Type:        schema.TypeString,
								Description: "The Gitea URL to talk to. For example https://gitea.mydomain.com/.",
								Optional:    true,
							},
							"insecure": {
								Type:        schema.TypeBool,
								Description: "Allow self-signed TLS / Certificates.",
								Optional:    true,
							},
							"owner": {
								Type:        schema.TypeString,
								Description: "Gitea organization or user to scan.",
								Required:    true,
							},
							"token_ref": {
								Type:        schema.TypeList,
								Description: "Authentication token reference.",
								Optional:    true,
								MaxItems:    1,
								Elem:        secretRefResource(),
							},
						},
					},
				},
				"github": {
					Type:        schema.TypeList,
					Description: "Uses the GitHub API to scan an organization in either github.com or GitHub Enterprise.",
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"all_branches": {
								Type:        schema.TypeBool,
								Description: "If true, scan every branch of every repository. If false, scan only the default branch.",
								Optional:    true,
							},
							"api": {
								Type:        schema.TypeString,
								Description: "The GitHub API URL to talk to. Default https://api.github.com/.",
								Optional:    true,
							},
							"app_secret_name": {
								Type:        schema.TypeString,
								Description: "Reference to a GitHub App repo-creds secret. Uses a GitHub App to access the API instead of a PAT.",
								Optional:    true,
							},
							"organization": {
								Type:        schema.TypeString,
								Description: "GitHub org to scan.",
								Required:    true,
							},
							"token_ref": {
								Type:        schema.TypeList,
								Description: "Authentication token reference.",
								Optional:    true,
								MaxItems:    1,
								Elem:        secretRefResource(),
							},
						},
					},
				},
				"gitlab": {
					Type:        schema.TypeList,
					Description: "Uses the GitLab API to scan and organization in either gitlab.com or self-hosted GitLab.",
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"all_branches": {
								Type:        schema.TypeBool,
								Description: "If true, scan every branch of every repository. If false, scan only the default branch.",
								Optional:    true,
							},
							"api": {
								Type:        schema.TypeString,
								Description: "The Gitlab API URL to talk to.",
								Optional:    true,
							},
							"group": {
								Type:        schema.TypeString,
								Description: "Gitlab group to scan. You can use either the project id (recommended) or the full namespaced path.",
								Required:    true,
							},
							"include_subgroups": {
								Type:        schema.TypeBool,
								Description: "Recurse through subgroups (true) or scan only the base group (false). Defaults to `false`.",
								Optional:    true,
							},
							"token_ref": {
								Type:        schema.TypeList,
								Description: "Authentication token reference.",
								Optional:    true,
								MaxItems:    1,
								Elem:        secretRefResource(),
							},
						},
					},
				},
				"requeue_after_seconds": {
					Type:        schema.TypeString,
					Description: "How often to check for changes (in seconds). Default: 3min.",
					Optional:    true,
				},
				"template": {
					Type:        schema.TypeList,
					Description: "Generator template. Used to override the values of the spec-level template.",
					Optional:    true,
					MaxItems:    1,
					Elem:        applicationSetTemplateResource(true),
				},
			},
		},
	}
}

func applicationSetPullRequestGeneratorSchemaV0() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "[Pull Request generators](https://argo-cd.readthedocs.io/en/stable/operator-manual/applicationset/Generators-Pull-Request/) uses the API of an SCMaaS provider to automatically discover open pull requests within a repository.",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"bitbucket_server": {
					Type:        schema.TypeList,
					Description: "Fetch pull requests from a repo hosted on a Bitbucket Server.",
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"api": {
								Type:        schema.TypeString,
								Description: "The Bitbucket REST API URL to talk to e.g. https://bitbucket.org/rest.",
								Required:    true,
							},
							"basic_auth": {
								Type:        schema.TypeList,
								Description: "Credentials for Basic auth.",
								Optional:    true,
								MaxItems:    1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"username": {
											Type:        schema.TypeString,
											Description: "Username for Basic auth.",
											Optional:    true,
										},
										"password_ref": {
											Type:        schema.TypeList,
											Description: "Password (or personal access token) reference.",
											Optional:    true,
											MaxItems:    1,
											Elem:        secretRefResource(),
										},
									},
								},
							},
							"project": {
								Type:        schema.TypeString,
								Description: "Project to scan.",
								Required:    true,
							},
							"repo": {
								Type:        schema.TypeString,
								Description: "Repo name to scan.",
								Required:    true,
							},
						},
					},
				},
				"filter": {
					Type:        schema.TypeList,
					Description: "Filters allow selecting which pull requests to generate for.",
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"branch_match": {
								Type:        schema.TypeString,
								Description: "A regex which must match the branch name.",
								Optional:    true,
							},
						},
					},
				},
				"gitea": {
					Type:        schema.TypeList,
					Description: "Specify the repository from which to fetch the Gitea Pull requests.",
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"api": {
								Type:        schema.TypeString,
								Description: "The Gitea API URL to talk to.",
								Required:    true,
							},
							"insecure": {
								Type:        schema.TypeBool,
								Description: "Allow insecure tls, for self-signed certificates; default: false.",
								Optional:    true,
							},
							"owner": {
								Type:        schema.TypeString,
								Description: "Gitea org or user to scan.",
								Required:    true,
							},
							"repo": {
								Type:        schema.TypeString,
								Description: "Gitea repo name to scan.",
								Required:    true,
							},
							"token_ref": {
								Type:        schema.TypeList,
								Description: "Authentication token reference.",
								Optional:    true,
								MaxItems:    1,
								Elem:        secretRefResource(),
							},
						},
					},
				},
				"github": {
					Type:        schema.TypeList,
					Description: "Specify the repository from which to fetch the GitHub Pull requests.",
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"api": {
								Type:        schema.TypeString,
								Description: "The GitHub API URL to talk to. Default https://api.github.com/.",
								Optional:    true,
							},
							"app_secret_name": {
								Type:        schema.TypeString,
								Description: "Reference to a GitHub App repo-creds secret with permission to access pull requests.",
								Optional:    true,
							},
							"labels": {
								Type:        schema.TypeList,
								Description: "Labels is used to filter the PRs that you want to target.",
								Optional:    true,
								Elem:        &schema.Schema{Type: schema.TypeString},
							},
							"owner": {
								Type:        schema.TypeString,
								Description: "GitHub org or user to scan.",
								Required:    true,
							},
							"repo": {
								Type:        schema.TypeString,
								Description: "GitHub repo name to scan.",
								Required:    true,
							},
							"token_ref": {
								Type:        schema.TypeList,
								Description: "Authentication token reference.",
								Optional:    true,
								MaxItems:    1,
								Elem:        secretRefResource(),
							},
						},
					},
				},
				"gitlab": {
					Type:        schema.TypeList,
					Description: "Specify the project from which to fetch the GitLab merge requests.",
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"api": {
								Type:        schema.TypeString,
								Description: "The GitLab API URL to talk to. If blank, uses https://gitlab.com/.",
								Optional:    true,
							},
							"labels": {
								Type:        schema.TypeList,
								Description: "Labels is used to filter the PRs that you want to target.",
								Optional:    true,
								Elem:        &schema.Schema{Type: schema.TypeString},
							},
							"project": {
								Type:        schema.TypeString,
								Description: "GitLab project to scan.",
								Required:    true,
							},
							"pull_request_state": {
								Type:        schema.TypeString,
								Description: "additional MRs filter to get only those with a certain state. Default:  \"\" (all states).",
								Optional:    true,
							},
							"token_ref": {
								Type:        schema.TypeList,
								Description: "Authentication token reference.",
								Optional:    true,
								MaxItems:    1,
								Elem:        secretRefResource(),
							},
						},
					},
				},
				"requeue_after_seconds": {
					Type:        schema.TypeString,
					Description: "How often to check for changes (in seconds). Default: 30min.",
					Optional:    true,
				},
				"template": {
					Type:        schema.TypeList,
					Description: "Generator template. Used to override the values of the spec-level template.",
					Optional:    true,
					MaxItems:    1,
					Elem:        applicationSetTemplateResource(true),
				},
			},
		},
	}
}

func applicationSetTemplateResource(allOptional bool) *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"metadata": {
				Type:        schema.TypeList,
				Description: "Kubernetes object metadata for templated Application.",
				Optional:    allOptional,
				Required:    !allOptional,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"annotations": {
							Type:        schema.TypeMap,
							Description: "An unstructured key value map that may be used to store arbitrary metadata for the resulting Application.",
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"labels": {
							Type:        schema.TypeMap,
							Description: "Map of string keys and values that can be used to organize and categorize (scope and select) the resulting Application.",
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"name": {
							Type:        schema.TypeString,
							Description: "Name of the resulting Application",
							Optional:    allOptional,
							Required:    !allOptional,
						},
						"namespace": {
							Type:        schema.TypeString,
							Description: "Namespace of the resulting Application",
							Optional:    true,
						},
						"finalizers": {
							Type:        schema.TypeList,
							Description: "List of finalizers to apply to the resulting Application.",
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"spec": applicationSpecSchemaV4(allOptional),
		},
	}
}

func secretRefResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key": {
				Type:        schema.TypeString,
				Description: "Key containing information in Kubernetes `Secret`.",
				Required:    true,
			},
			"secret_name": {
				Type:        schema.TypeString,
				Description: "Name of Kubernetes `Secret`.",
				Required:    true,
			},
		},
	}
}

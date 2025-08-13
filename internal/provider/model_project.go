package provider

import (
	"github.com/argoproj-labs/terraform-provider-argocd/internal/validators"
	"github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type projectModel struct {
	ID       types.String       `tfsdk:"id"`
	Metadata []objectMeta       `tfsdk:"metadata"`
	Spec     []projectSpecModel `tfsdk:"spec"`
}

type projectSpecModel struct {
	ClusterResourceBlacklist   []groupKindModel                 `tfsdk:"cluster_resource_blacklist"`
	ClusterResourceWhitelist   []groupKindModel                 `tfsdk:"cluster_resource_whitelist"`
	Description                types.String                     `tfsdk:"description"`
	Destination                []destinationModel               `tfsdk:"destination"`
	DestinationServiceAccount  []destinationServiceAccountModel `tfsdk:"destination_service_account"`
	NamespaceResourceBlacklist []groupKindModel                 `tfsdk:"namespace_resource_blacklist"`
	NamespaceResourceWhitelist []groupKindModel                 `tfsdk:"namespace_resource_whitelist"`
	OrphanedResources          []orphanedResourcesModel         `tfsdk:"orphaned_resources"`
	Role                       []projectRoleModel               `tfsdk:"role"`
	SourceRepos                []types.String                   `tfsdk:"source_repos"`
	SourceNamespaces           []types.String                   `tfsdk:"source_namespaces"`
	SignatureKeys              []types.String                   `tfsdk:"signature_keys"`
	SyncWindow                 []syncWindowModel                `tfsdk:"sync_window"`
}

type groupKindModel struct {
	Group types.String `tfsdk:"group"`
	Kind  types.String `tfsdk:"kind"`
}

type destinationModel struct {
	Server    types.String `tfsdk:"server"`
	Namespace types.String `tfsdk:"namespace"`
	Name      types.String `tfsdk:"name"`
}

type destinationServiceAccountModel struct {
	DefaultServiceAccount types.String `tfsdk:"default_service_account"`
	Namespace             types.String `tfsdk:"namespace"`
	Server                types.String `tfsdk:"server"`
}

type orphanedResourcesModel struct {
	Warn   types.Bool                     `tfsdk:"warn"`
	Ignore []orphanedResourcesIgnoreModel `tfsdk:"ignore"`
}

type orphanedResourcesIgnoreModel struct {
	Group types.String `tfsdk:"group"`
	Kind  types.String `tfsdk:"kind"`
	Name  types.String `tfsdk:"name"`
}

type projectRoleModel struct {
	Description types.String    `tfsdk:"description"`
	Groups      []types.String  `tfsdk:"groups"`
	Name        types.String    `tfsdk:"name"`
	Policies    []types.String  `tfsdk:"policies"`
	JwtTokens   []jwtTokenModel `tfsdk:"jwt_tokens"`
}

type jwtTokenModel struct {
	ID  types.String `tfsdk:"id"`
	Iat types.Int64  `tfsdk:"iat"`
	Exp types.Int64  `tfsdk:"exp"`
}

type syncWindowModel struct {
	Applications []types.String `tfsdk:"applications"`
	Clusters     []types.String `tfsdk:"clusters"`
	Duration     types.String   `tfsdk:"duration"`
	Kind         types.String   `tfsdk:"kind"`
	ManualSync   types.Bool     `tfsdk:"manual_sync"`
	Namespaces   []types.String `tfsdk:"namespaces"`
	Schedule     types.String   `tfsdk:"schedule"`
	Timezone     types.String   `tfsdk:"timezone"`
}

func projectSchemaBlocks() map[string]schema.Block {
	return map[string]schema.Block{
		"metadata": objectMetaSchemaListBlock("appproject", false),
		"spec": schema.ListNestedBlock{
			Description: "ArgoCD AppProject spec.",
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
			},
			NestedObject: schema.NestedBlockObject{
				Attributes: projectSpecSchemaAttributesOnly(),
				Blocks:     projectSpecSchemaBlocks(),
			},
		},
	}
}

func projectSpecSchemaBlocks() map[string]schema.Block {
	return map[string]schema.Block{
		"destination": schema.SetNestedBlock{
			Description: "Destinations available for deployment.",
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"server": schema.StringAttribute{
						Description: "URL of the target cluster and must be set to the Kubernetes control plane API.",
						Optional:    true,
					},
					"namespace": schema.StringAttribute{
						Description: "Target namespace for applications' resources.",
						Required:    true,
					},
					"name": schema.StringAttribute{
						Description: "Name of the destination cluster which can be used instead of server.",
						Optional:    true,
					},
				},
			},
		},
		"cluster_resource_blacklist": schema.SetNestedBlock{
			Description: "Blacklisted cluster level resources.",
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"group": schema.StringAttribute{
						Description: "The Kubernetes resource Group to match for.",
						Optional:    true,
						Validators: []validator.String{
							validators.GroupNameValidator(),
						},
					},
					"kind": schema.StringAttribute{
						Description: "The Kubernetes resource Kind to match for.",
						Optional:    true,
					},
				},
			},
		},
		"cluster_resource_whitelist": schema.SetNestedBlock{
			Description: "Whitelisted cluster level resources.",
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"group": schema.StringAttribute{
						Description: "The Kubernetes resource Group to match for.",
						Optional:    true,
						Validators: []validator.String{
							validators.GroupNameValidator(),
						},
					},
					"kind": schema.StringAttribute{
						Description: "The Kubernetes resource Kind to match for.",
						Optional:    true,
					},
				},
			},
		},
		"namespace_resource_blacklist": schema.SetNestedBlock{
			Description: "Blacklisted namespace level resources.",
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"group": schema.StringAttribute{
						Description: "The Kubernetes resource Group to match for.",
						Optional:    true,
						Validators: []validator.String{
							validators.GroupNameValidator(),
						},
					},
					"kind": schema.StringAttribute{
						Description: "The Kubernetes resource Kind to match for.",
						Optional:    true,
					},
				},
			},
		},
		"namespace_resource_whitelist": schema.SetNestedBlock{
			Description: "Whitelisted namespace level resources.",
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"group": schema.StringAttribute{
						Description: "The Kubernetes resource Group to match for.",
						Optional:    true,
						Validators: []validator.String{
							validators.GroupNameValidator(),
						},
					},
					"kind": schema.StringAttribute{
						Description: "The Kubernetes resource Kind to match for.",
						Optional:    true,
					},
				},
			},
		},
		"orphaned_resources": schema.SetNestedBlock{
			Description: "Configuration for orphaned resources tracking.",
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"warn": schema.BoolAttribute{
						Description: "Warn about orphaned resources.",
						Optional:    true,
					},
				},
				Blocks: map[string]schema.Block{
					"ignore": schema.SetNestedBlock{
						Description: "List of resources to ignore during orphaned resources detection.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"group": schema.StringAttribute{
									Description: "The Kubernetes resource Group to match for.",
									Optional:    true,
									Validators: []validator.String{
										validators.GroupNameValidator(),
									},
								},
								"kind": schema.StringAttribute{
									Description: "The Kubernetes resource Kind to match for.",
									Optional:    true,
								},
								"name": schema.StringAttribute{
									Description: "The Kubernetes resource Name to match for.",
									Optional:    true,
								},
							},
						},
					},
				},
			},
		},
		"role": schema.SetNestedBlock{
			Description: "Project roles.",
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Description: "The name of the role.",
						Required:    true,
					},
					"description": schema.StringAttribute{
						Description: "The role description.",
						Optional:    true,
					},
					"policies": schema.SetAttribute{
						Description: "The list of policies associated with the role.",
						Required:    true,
						ElementType: types.StringType,
					},
					"groups": schema.SetAttribute{
						Description: "The list of groups associated with the role.",
						Optional:    true,
						ElementType: types.StringType,
					},
					"jwt_tokens": schema.SetNestedAttribute{
						Description: "List of JWT tokens issued for this role.",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"iat": schema.Int64Attribute{
									Description: "Token issued at (timestamp).",
									Required:    true,
								},
								"id": schema.StringAttribute{
									Description: "Token identifier.",
									Optional:    true,
								},
								"exp": schema.Int64Attribute{
									Description: "Token expiration (timestamp).",
									Optional:    true,
								},
							},
						},
					},
				},
			},
		},
		"destination_service_account": schema.SetNestedBlock{
			Description: "Service accounts to be impersonated for the application sync operation for each destination.",
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"default_service_account": schema.StringAttribute{
						Description: "Used for impersonation during the sync operation",
						Required:    true,
					},
					"namespace": schema.StringAttribute{
						Description: "Specifies the target namespace for the application's resources.",
						Optional:    true,
					},
					"server": schema.StringAttribute{
						Description: "Specifies the URL of the target cluster's Kubernetes control plane API.",
						Optional:    true,
					},
				},
			},
		},
		"sync_window": schema.SetNestedBlock{
			Description: "Controls when sync operations are allowed for the project.",
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"kind": schema.StringAttribute{
						Description: "Defines if the window allows or blocks syncs.",
						Optional:    true,
						Validators: []validator.String{
							validators.SyncWindowKindValidator(),
						},
					},
					"applications": schema.SetAttribute{
						Description: "The list of applications assigned to this sync window.",
						Optional:    true,
						ElementType: types.StringType,
					},
					"namespaces": schema.SetAttribute{
						Description: "The list of namespaces assigned to this sync window.",
						Optional:    true,
						ElementType: types.StringType,
					},
					"clusters": schema.SetAttribute{
						Description: "The list of clusters assigned to this sync window.",
						Optional:    true,
						ElementType: types.StringType,
					},
					"manual_sync": schema.BoolAttribute{
						Description: "Defines if sync will be blocked if the sync window is active.",
						Optional:    true,
					},
					"schedule": schema.StringAttribute{
						Description: "Time the window will begin, specified in cron format.",
						Optional:    true,
						Validators: []validator.String{
							validators.SyncWindowScheduleValidator(),
						},
					},
					"duration": schema.StringAttribute{
						Description: "The duration of the sync window.",
						Optional:    true,
						Validators: []validator.String{
							validators.DurationValidator(),
						},
					},
					"timezone": schema.StringAttribute{
						Description: "Timezone that the schedule will be evaluated in.",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString("UTC"),
						Validators: []validator.String{
							validators.SyncWindowTimezoneValidator(),
						},
					},
				},
			},
		},
	}
}

func projectSpecSchemaAttributesOnly() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"description": schema.StringAttribute{
			Description: "Project description.",
			Optional:    true,
		},
		"source_repos": schema.SetAttribute{
			Description: "List of repositories from which applications may be created.",
			Optional:    true,
			ElementType: types.StringType,
		},
		"source_namespaces": schema.SetAttribute{
			Description: "List of source namespaces for applications.",
			Optional:    true,
			ElementType: types.StringType,
		},
		"signature_keys": schema.SetAttribute{
			Description: "Signature keys for verifying the integrity of applications.",
			Optional:    true,
			ElementType: types.StringType,
		},
	}
}

func newProject(project *v1alpha1.AppProject) *projectModel {
	p := &projectModel{
		Metadata: []objectMeta{newObjectMeta(project.ObjectMeta)},
		Spec:     []projectSpecModel{newProjectSpec(&project.Spec)},
	}
	return p
}

func newProjectSpec(spec *v1alpha1.AppProjectSpec) projectSpecModel {
	ps := projectSpecModel{
		Description: types.StringValue(spec.Description),
	}

	// Convert source repos
	if len(spec.SourceRepos) > 0 {
		ps.SourceRepos = make([]types.String, len(spec.SourceRepos))
		for i, repo := range spec.SourceRepos {
			ps.SourceRepos[i] = types.StringValue(repo)
		}
	}

	// Convert signature keys
	if len(spec.SignatureKeys) > 0 {
		ps.SignatureKeys = make([]types.String, len(spec.SignatureKeys))
		for i, key := range spec.SignatureKeys {
			ps.SignatureKeys[i] = types.StringValue(key.KeyID)
		}
	}

	// Convert source namespaces
	if len(spec.SourceNamespaces) > 0 {
		ps.SourceNamespaces = make([]types.String, len(spec.SourceNamespaces))
		for i, ns := range spec.SourceNamespaces {
			ps.SourceNamespaces[i] = types.StringValue(ns)
		}
	}

	// Convert cluster resource blacklist
	if len(spec.ClusterResourceBlacklist) > 0 {
		ps.ClusterResourceBlacklist = make([]groupKindModel, len(spec.ClusterResourceBlacklist))
		for i, gk := range spec.ClusterResourceBlacklist {
			ps.ClusterResourceBlacklist[i] = groupKindModel{
				Group: types.StringValue(gk.Group),
				Kind:  types.StringValue(gk.Kind),
			}
		}
	}

	// Convert cluster resource whitelist
	if len(spec.ClusterResourceWhitelist) > 0 {
		ps.ClusterResourceWhitelist = make([]groupKindModel, len(spec.ClusterResourceWhitelist))
		for i, gk := range spec.ClusterResourceWhitelist {
			ps.ClusterResourceWhitelist[i] = groupKindModel{
				Group: types.StringValue(gk.Group),
				Kind:  types.StringValue(gk.Kind),
			}
		}
	}

	// Convert namespace resource blacklist
	if len(spec.NamespaceResourceBlacklist) > 0 {
		ps.NamespaceResourceBlacklist = make([]groupKindModel, len(spec.NamespaceResourceBlacklist))
		for i, gk := range spec.NamespaceResourceBlacklist {
			ps.NamespaceResourceBlacklist[i] = groupKindModel{
				Group: types.StringValue(gk.Group),
				Kind:  types.StringValue(gk.Kind),
			}
		}
	}

	// Convert namespace resource whitelist
	if len(spec.NamespaceResourceWhitelist) > 0 {
		ps.NamespaceResourceWhitelist = make([]groupKindModel, len(spec.NamespaceResourceWhitelist))
		for i, gk := range spec.NamespaceResourceWhitelist {
			ps.NamespaceResourceWhitelist[i] = groupKindModel{
				Group: types.StringValue(gk.Group),
				Kind:  types.StringValue(gk.Kind),
			}
		}
	}

	// Convert destinations
	if len(spec.Destinations) > 0 {
		ps.Destination = make([]destinationModel, len(spec.Destinations))
		for i, dest := range spec.Destinations {
			d := destinationModel{
				Namespace: types.StringValue(dest.Namespace),
			}
			if dest.Server != "" {
				d.Server = types.StringValue(dest.Server)
			} else {
				d.Server = types.StringNull()
			}
			if dest.Name != "" {
				d.Name = types.StringValue(dest.Name)
			} else {
				d.Name = types.StringNull()
			}
			ps.Destination[i] = d
		}
	}

	// Convert destination service accounts
	if len(spec.DestinationServiceAccounts) > 0 {
		ps.DestinationServiceAccount = make([]destinationServiceAccountModel, len(spec.DestinationServiceAccounts))
		for i, dsa := range spec.DestinationServiceAccounts {
			ps.DestinationServiceAccount[i] = destinationServiceAccountModel{
				DefaultServiceAccount: types.StringValue(dsa.DefaultServiceAccount),
				Namespace:             types.StringValue(dsa.Namespace),
				Server:                types.StringValue(dsa.Server),
			}
		}
	}

	// Convert orphaned resources
	if spec.OrphanedResources != nil {
		or := orphanedResourcesModel{
			Warn: types.BoolPointerValue(spec.OrphanedResources.Warn),
		}
		if len(spec.OrphanedResources.Ignore) > 0 {
			or.Ignore = make([]orphanedResourcesIgnoreModel, len(spec.OrphanedResources.Ignore))
			for i, ignore := range spec.OrphanedResources.Ignore {
				or.Ignore[i] = orphanedResourcesIgnoreModel{
					Group: types.StringValue(ignore.Group),
					Kind:  types.StringValue(ignore.Kind),
					Name:  types.StringValue(ignore.Name),
				}
			}
		}
		ps.OrphanedResources = []orphanedResourcesModel{or}
	}

	// Convert roles
	if len(spec.Roles) > 0 {
		ps.Role = make([]projectRoleModel, len(spec.Roles))
		for i, role := range spec.Roles {
			pr := projectRoleModel{
				Name: types.StringValue(role.Name),
			}

			// Handle description
			if role.Description != "" {
				pr.Description = types.StringValue(role.Description)
			} else {
				pr.Description = types.StringNull()
			}

			// Handle policies
			if len(role.Policies) > 0 {
				pr.Policies = make([]types.String, len(role.Policies))
				for j, policy := range role.Policies {
					pr.Policies[j] = types.StringValue(policy)
				}
			}

			// Handle groups
			if len(role.Groups) > 0 {
				pr.Groups = make([]types.String, len(role.Groups))
				for j, group := range role.Groups {
					pr.Groups[j] = types.StringValue(group)
				}
			}

			// JWT tokens are not managed by the project resource - they are managed by argocd_project_token resources
			// So we explicitly set them to nil to avoid conflicts and ensure they don't appear in state
			pr.JwtTokens = nil

			ps.Role[i] = pr
		}
	}

	// Convert sync windows
	if len(spec.SyncWindows) > 0 {
		ps.SyncWindow = make([]syncWindowModel, len(spec.SyncWindows))
		for i, sw := range spec.SyncWindows {
			swm := syncWindowModel{
				Duration:   types.StringValue(sw.Duration),
				Kind:       types.StringValue(sw.Kind),
				ManualSync: types.BoolValue(sw.ManualSync),
				Schedule:   types.StringValue(sw.Schedule),
				Timezone:   types.StringValue("UTC"), // Default
			}

			if sw.TimeZone != "" {
				swm.Timezone = types.StringValue(sw.TimeZone)
			}

			if len(sw.Applications) > 0 {
				swm.Applications = make([]types.String, len(sw.Applications))
				for j, app := range sw.Applications {
					swm.Applications[j] = types.StringValue(app)
				}
			}

			if len(sw.Clusters) > 0 {
				swm.Clusters = make([]types.String, len(sw.Clusters))
				for j, cluster := range sw.Clusters {
					swm.Clusters[j] = types.StringValue(cluster)
				}
			}

			if len(sw.Namespaces) > 0 {
				swm.Namespaces = make([]types.String, len(sw.Namespaces))
				for j, ns := range sw.Namespaces {
					swm.Namespaces[j] = types.StringValue(ns)
				}
			}

			ps.SyncWindow[i] = swm
		}
	}

	return ps
}

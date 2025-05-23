package provider

import (
	"github.com/argoproj-labs/terraform-provider-argocd/internal/utils"
	"github.com/argoproj-labs/terraform-provider-argocd/internal/validators"
	"github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type applicationModel struct {
	ID       types.String       `tfsdk:"id"`
	Metadata objectMeta         `tfsdk:"metadata"`
	Spec     *applicationSpec   `tfsdk:"spec"`
	Status   *applicationStatus `tfsdk:"status"`
}

type applicationSpec struct {
	Destination          applicationDestination                 `tfsdk:"destination"`
	IgnoreDifferences    []applicationResourceIgnoreDifferences `tfsdk:"ignore_differences"`
	Infos                []applicationInfo                      `tfsdk:"infos"`
	Project              types.String                           `tfsdk:"project"`
	RevisionHistoryLimit types.Int64                            `tfsdk:"revision_history_limit"`
	Sources              []applicationSource                    `tfsdk:"sources"`
	SyncPolicy           *applicationSyncPolicy                 `tfsdk:"sync_policy"`
}

func applicationSpecSchemaAttribute(allOptional, computed bool) schema.Attribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "The application specification.",
		Computed:            computed,
		Required:            !computed,
		Attributes: map[string]schema.Attribute{
			"destination":        applicationDestinationSchemaAttribute(computed),
			"ignore_differences": applicationResourceIgnoreDifferencesSchemaAttribute(computed),
			"infos":              applicationInfoSchemaAttribute(computed),
			"project": schema.StringAttribute{
				Computed:            computed,
				Optional:            !computed,
				MarkdownDescription: "The project the application belongs to. Defaults to `default`.",
				Default:             stringdefault.StaticString("default"),
			},
			"revision_history_limit": schema.Int64Attribute{
				MarkdownDescription: "Limits the number of items kept in the application's revision history, which is used for informational purposes as well as for rollbacks to previous versions. This should only be changed in exceptional circumstances. Setting to zero will store no history. This will reduce storage used. Increasing will increase the space used to store the history, so we do not recommend increasing it. Default is 10.",
				Computed:            computed,
				Optional:            !computed,
			},
			"sources":     applicationSourcesSchemaAttribute(allOptional, computed),
			"sync_policy": applicationSyncPolicySchemaAttribute(computed),
		},
	}
}

func newApplicationSpec(as v1alpha1.ApplicationSpec) *applicationSpec {
	m := &applicationSpec{
		Destination:          newApplicationDestination(as.Destination),
		IgnoreDifferences:    newApplicationResourceIgnoreDifferences(as.IgnoreDifferences),
		Infos:                newApplicationInfos(as.Info),
		Project:              types.StringValue(as.Project),
		RevisionHistoryLimit: utils.OptionalInt64(as.RevisionHistoryLimit),
		SyncPolicy:           newApplicationSyncPolicy(as.SyncPolicy),
	}

	if as.Source != nil {
		m.Sources = append(m.Sources, newApplicationSource(*as.Source))
	}

	for _, v := range as.Sources {
		m.Sources = append(m.Sources, newApplicationSource(v))
	}

	return m
}

type applicationDestination struct {
	Server    types.String `tfsdk:"server"`
	Namespace types.String `tfsdk:"namespace"`
	Name      types.String `tfsdk:"name"`
}

func applicationDestinationSchemaAttribute(computed bool) schema.Attribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Reference to the Kubernetes server and namespace in which the application will be deployed.",
		Computed:            computed,
		Required:            !computed,
		Attributes: map[string]schema.Attribute{
			"server": schema.StringAttribute{
				MarkdownDescription: "URL of the target cluster and must be set to the Kubernetes control plane API.",
				Computed:            computed,
				Optional:            !computed,
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "Target namespace for the application's resources. The namespace will only be set for namespace-scoped resources that have not set a value for .metadata.namespace.",
				Computed:            computed,
				Optional:            !computed,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the target cluster. Can be used instead of `server`.",
				Computed:            computed,
				Optional:            !computed,
			},
		},
	}
}

func newApplicationDestination(ad v1alpha1.ApplicationDestination) applicationDestination {
	return applicationDestination{
		Name:      types.StringValue(ad.Name),
		Namespace: types.StringValue(ad.Namespace),
		Server:    types.StringValue(ad.Server),
	}
}

type applicationResourceIgnoreDifferences struct {
	Group             types.String   `tfsdk:"group"`
	Kind              types.String   `tfsdk:"kind"`
	Name              types.String   `tfsdk:"name"`
	Namespace         types.String   `tfsdk:"namespace"`
	JsonPointers      []types.String `tfsdk:"json_pointers"`
	JQPathExpressions []types.String `tfsdk:"jq_path_expressions"`
}

func applicationResourceIgnoreDifferencesSchemaAttribute(computed bool) schema.Attribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Resources and their fields which should be ignored during comparison. More info: https://argo-cd.readthedocs.io/en/stable/user-guide/diffing/#application-level-configuration.",
		Computed:            computed,
		Optional:            !computed,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"group": schema.StringAttribute{
					MarkdownDescription: "The Kubernetes resource Group to match for.",
					Computed:            computed,
					Optional:            !computed,
				},
				"kind": schema.StringAttribute{
					MarkdownDescription: "The Kubernetes resource Kind to match for.",
					Computed:            computed,
					Optional:            !computed,
				},
				"name": schema.StringAttribute{
					MarkdownDescription: "The Kubernetes resource Name to match for.",
					Computed:            computed,
					Optional:            !computed,
				},
				"namespace": schema.StringAttribute{
					MarkdownDescription: "The Kubernetes resource Namespace to match for.",
					Computed:            computed,
					Optional:            !computed,
				},
				"json_pointers": schema.SetAttribute{
					MarkdownDescription: "List of JSONPaths strings targeting the field(s) to ignore.",
					Computed:            computed,
					Optional:            !computed,
					ElementType:         types.StringType,
				},
				"jq_path_expressions": schema.SetAttribute{
					MarkdownDescription: "List of JQ path expression strings targeting the field(s) to ignore.",
					Computed:            computed,
					Optional:            !computed,
					ElementType:         types.StringType,
				},
			},
		},
	}
}

func newApplicationResourceIgnoreDifferences(diffs []v1alpha1.ResourceIgnoreDifferences) []applicationResourceIgnoreDifferences {
	if diffs == nil {
		return nil
	}

	ds := make([]applicationResourceIgnoreDifferences, len(diffs))
	for i, v := range diffs {
		ds[i] = applicationResourceIgnoreDifferences{
			Group:             types.StringValue(v.Group),
			Kind:              types.StringValue(v.Kind),
			Name:              types.StringValue(v.Name),
			Namespace:         types.StringValue(v.Namespace),
			JsonPointers:      pie.Map(v.JSONPointers, types.StringValue),
			JQPathExpressions: pie.Map(v.JQPathExpressions, types.StringValue),
		}
	}

	return ds
}

type applicationInfo struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func applicationInfoSchemaAttribute(computed bool) schema.Attribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "List of information (URLs, email addresses, and plain text) that relates to the application.",
		Computed:            computed,
		Optional:            !computed,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: "Name of the information.",
					Computed:            computed,
					Optional:            !computed,
				},
				"value": schema.StringAttribute{
					MarkdownDescription: "Value of the information.",
					Computed:            computed,
					Optional:            !computed,
				},
			},
		},
	}
}

func newApplicationInfos(infos []v1alpha1.Info) []applicationInfo {
	if infos == nil {
		return nil
	}

	is := make([]applicationInfo, len(infos))
	for i, v := range infos {
		is[i] = applicationInfo{
			Name:  types.StringValue(v.Name),
			Value: types.StringValue(v.Value),
		}
	}

	return is
}

type applicationSource struct {
	Chart          types.String                `tfsdk:"chart"`
	Directory      *applicationSourceDirectory `tfsdk:"directory"`
	Helm           *applicationSourceHelm      `tfsdk:"helm"`
	Kustomize      *applicationSourceKustomize `tfsdk:"kustomize"`
	Name           types.String                `tfsdk:"name"`
	Path           types.String                `tfsdk:"path"`
	Plugin         *applicationSourcePlugin    `tfsdk:"plugin"`
	Ref            types.String                `tfsdk:"ref"`
	RepoURL        types.String                `tfsdk:"repo_url"`
	TargetRevision types.String                `tfsdk:"target_revision"`
}

func applicationSourcesSchemaAttribute(allOptional, computed bool) schema.Attribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Location of the application's manifests or chart.",
		Computed:            computed,
		Required:            !computed,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"chart": schema.StringAttribute{
					MarkdownDescription: "Helm chart name. Must be specified for applications sourced from a Helm repo.",
					Computed:            computed,
					Optional:            !computed,
				},
				"directory": applicationSourceDirectorySchemaAttribute(computed),
				"helm":      applicationSourceHelmSchemaAttribute(computed),
				"kustomize": applicationSourceKustomizeSchemaAttribute(computed),
				"name": schema.StringAttribute{
					MarkdownDescription: "Name is used to refer to a source and is displayed in the UI. It is supported in multi-source Applications since version 2.14",
					Computed:            computed,
					Optional:            !computed,
				},
				"path": schema.StringAttribute{
					MarkdownDescription: "Directory path within the repository. Only valid for applications sourced from Git.",
					Computed:            computed,
					Optional:            !computed,
					Default:             stringdefault.StaticString("."),
				},
				"plugin": applicationSourcePluginSchemaAttribute(computed),
				"ref": schema.StringAttribute{
					MarkdownDescription: "Reference to another `source` within defined sources. See associated documentation on [Helm value files from external Git repository](https://argo-cd.readthedocs.io/en/stable/user-guide/multiple_sources/#helm-value-files-from-external-git-repository) regarding combining `ref` with `path` and/or `chart`.",
					Computed:            computed,
					Optional:            !computed,
				},
				"repo_url": schema.StringAttribute{
					MarkdownDescription: "URL to the repository (Git or Helm) that contains the application manifests.",
					Optional:            allOptional && !computed,
					Required:            !allOptional && !computed,
					Computed:            computed,
				},
				"target_revision": schema.StringAttribute{
					MarkdownDescription: "Revision of the source to sync the application to. In case of Git, this can be commit, tag, or branch. If omitted, will equal to HEAD. In case of Helm, this is a semver tag for the Chart's version.",
					Computed:            computed,
					Optional:            !computed,
				},
			},
		},
	}
}

func newApplicationSource(as v1alpha1.ApplicationSource) applicationSource {
	return applicationSource{
		Chart:          types.StringValue(as.Chart),
		Directory:      newApplicationSourceDirectory(as.Directory),
		Helm:           newApplicationSourceHelm(as.Helm),
		Kustomize:      newApplicationSourceKustomize(as.Kustomize),
		Name:           types.StringValue(as.Name),
		Path:           types.StringValue(as.Path),
		Plugin:         newApplicationSourcePlugin(as.Plugin),
		Ref:            types.StringValue(as.Ref),
		RepoURL:        types.StringValue(as.RepoURL),
		TargetRevision: types.StringValue(as.TargetRevision),
	}
}

type applicationSourceDirectory struct {
	Exclude types.String             `tfsdk:"exclude"`
	Jsonnet applicationSourceJsonnet `tfsdk:"jsonnet"`
	Include types.String             `tfsdk:"include"`
	Recurse types.Bool               `tfsdk:"recurse"`
}

func applicationSourceDirectorySchemaAttribute(computed bool) schema.Attribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Path/directory specific options.",
		Computed:            computed,
		Optional:            !computed,
		// TODO: This isn't used at present but we will need to migrate it if we
		// migrate the ArgoCD application resource.
		//
		// DiffSuppressFunc: func(k,
		// oldValue, newValue string, d *schema.ResourceData) bool {
		//  // Avoid drift when recurse is explicitly set to false
		//  // Also ignore the directory node if both recurse & jsonnet are not set or ignored
		//  if k == "spec.0.source.0.directory.0.recurse" && oldValue == "" && newValue == "false" {
		//      return true
		//  }
		//  if k == "spec.0.source.0.directory.#" {
		//      _, hasRecurse := d.GetOk("spec.0.source.0.directory.0.recurse")
		//      _, hasJsonnet := d.GetOk("spec.0.source.0.directory.0.jsonnet")

		// 		if !hasJsonnet && !hasRecurse {
		// 			return true
		// 		}
		// 	}
		// 	return false
		// },
		Attributes: map[string]schema.Attribute{
			"exclude": schema.StringAttribute{
				MarkdownDescription: "Glob pattern to match paths against that should be explicitly excluded from being used during manifest generation. This takes precedence over the `include` field. To match multiple patterns, wrap the patterns in {} and separate them with commas. For example: '{config.yaml,env-use2/*}'",
				Computed:            computed,
				Optional:            !computed,
			},
			"include": schema.StringAttribute{
				MarkdownDescription: "Glob pattern to match paths against that should be explicitly included during manifest generation. If this field is set, only matching manifests will be included. To match multiple patterns, wrap the patterns in {} and separate them with commas. For example: '{*.yml,*.yaml}'",
				Computed:            computed,
				Optional:            !computed,
			},
			"jsonnet": applicationSourceJsonnetSchemaAttribute(computed),
			"recurse": schema.BoolAttribute{
				MarkdownDescription: "Whether to scan a directory recursively for manifests.",
				Computed:            computed,
				Optional:            !computed,
			},
		},
	}
}

func newApplicationSourceDirectory(ad *v1alpha1.ApplicationSourceDirectory) *applicationSourceDirectory {
	if ad == nil {
		return nil
	}

	return &applicationSourceDirectory{
		Exclude: types.StringValue(ad.Exclude),
		Jsonnet: newApplicationSourceJsonnet(ad.Jsonnet),
		Include: types.StringValue(ad.Include),
		Recurse: types.BoolValue(ad.Recurse),
	}
}

type applicationSourceJsonnet struct {
	ExtVars []applicationJsonnetVar `tfsdk:"ext_vars"`
	Libs    []types.String          `tfsdk:"libs"`
	TLAs    []applicationJsonnetVar `tfsdk:"tlas"`
}

func applicationSourceJsonnetSchemaAttribute(computed bool) schema.Attribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Jsonnet specific options.",
		Computed:            computed,
		Optional:            !computed,
		Attributes: map[string]schema.Attribute{
			"ext_vars": schema.ListNestedAttribute{
				MarkdownDescription: "List of Jsonnet External Variables.",
				Computed:            computed,
				Optional:            !computed,
				NestedObject:        applicationJsonnetVarSchemaNestedAttributeObject(computed),
			},
			"libs": schema.ListAttribute{
				MarkdownDescription: "Additional library search dirs.",
				Computed:            computed,
				Optional:            !computed,
				ElementType:         types.StringType,
			},
			"tlas": schema.ListNestedAttribute{
				MarkdownDescription: "List of Jsonnet Top-level Arguments",
				Computed:            computed,
				Optional:            !computed,
				NestedObject:        applicationJsonnetVarSchemaNestedAttributeObject(computed),
			},
		},
	}
}

func newApplicationSourceJsonnet(asj v1alpha1.ApplicationSourceJsonnet) applicationSourceJsonnet {
	return applicationSourceJsonnet{
		ExtVars: newApplicationJsonnetVars(asj.ExtVars),
		Libs:    pie.Map(asj.Libs, types.StringValue),
		TLAs:    newApplicationJsonnetVars(asj.TLAs),
	}
}

type applicationJsonnetVar struct {
	Code  types.Bool   `tfsdk:"code"`
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func applicationJsonnetVarSchemaNestedAttributeObject(computed bool) schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of Jsonnet variable.",
				Computed:            computed,
				Optional:            !computed,
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "Value of Jsonnet variable.",
				Computed:            computed,
				Optional:            !computed,
			},
			"code": schema.BoolAttribute{
				MarkdownDescription: "Determines whether the variable should be evaluated as jsonnet code or treated as string.",
				Computed:            computed,
				Optional:            !computed,
			},
		},
	}
}

func newApplicationJsonnetVars(jvs []v1alpha1.JsonnetVar) []applicationJsonnetVar {
	if jvs == nil {
		return nil
	}

	vs := make([]applicationJsonnetVar, len(jvs))
	for i, v := range jvs {
		vs[i] = applicationJsonnetVar{
			Code:  types.BoolValue(v.Code),
			Name:  types.StringValue(v.Name),
			Value: types.StringValue(v.Value),
		}
	}

	return vs
}

type applicationSourceHelm struct {
	FileParameters          []applicationHelmFileParameter `tfsdk:"file_parameters"`
	IgnoreMissingValueFiles types.Bool                     `tfsdk:"ignore_missing_value_files"`
	Parameters              []applicationHelmParameter     `tfsdk:"parameters"`
	PassCredentials         types.Bool                     `tfsdk:"pass_credentials"`
	ReleaseName             types.String                   `tfsdk:"release_name"`
	SkipCRDs                types.Bool                     `tfsdk:"skip_crds"`
	ValueFiles              []types.String                 `tfsdk:"value_files"`
	Values                  types.String                   `tfsdk:"values"`
}

func applicationSourceHelmSchemaAttribute(computed bool) schema.Attribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Helm specific options.",
		Computed:            computed,
		Optional:            !computed,
		Attributes: map[string]schema.Attribute{
			"file_parameters": applicationHelmFileParameterSchemaAttribute(computed),
			"ignore_missing_value_files": schema.BoolAttribute{
				MarkdownDescription: "Prevents 'helm template' from failing when `value_files` do not exist locally by not appending them to 'helm template --values'.",
				Computed:            computed,
				Optional:            !computed,
			},
			"parameters": applicationHelmParameterSchemaAttribute(computed),
			"release_name": schema.StringAttribute{
				MarkdownDescription: "Helm release name. If omitted it will use the application name.",
				Computed:            computed,
				Optional:            !computed,
			},
			"skip_crds": schema.BoolAttribute{
				MarkdownDescription: "Whether to skip custom resource definition installation step (Helm's [--skip-crds](https://helm.sh/docs/chart_best_practices/custom_resource_definitions/)).",
				Computed:            computed,
				Optional:            !computed,
			},
			"pass_credentials": schema.BoolAttribute{
				MarkdownDescription: "If true then adds '--pass-credentials' to Helm commands to pass credentials to all domains.",
				Computed:            computed,
				Optional:            !computed,
			},
			"values": schema.StringAttribute{
				MarkdownDescription: "Helm values to be passed to 'helm template', typically defined as a Attribute.",
				Computed:            computed,
				Optional:            !computed,
			},
			"value_files": schema.ListAttribute{
				MarkdownDescription: "List of Helm value files to use when generating a template.",
				Computed:            computed,
				Optional:            !computed,
				ElementType:         types.StringType,
			},
		},
	}
}

func newApplicationSourceHelm(ash *v1alpha1.ApplicationSourceHelm) *applicationSourceHelm {
	if ash == nil {
		return nil
	}

	return &applicationSourceHelm{
		FileParameters:          newApplicationSourceHelmFileParameters(ash.FileParameters),
		IgnoreMissingValueFiles: types.BoolValue(ash.IgnoreMissingValueFiles),
		Parameters:              newApplicationSourceHelmParameters(ash.Parameters),
		PassCredentials:         types.BoolValue(ash.PassCredentials),
		ReleaseName:             types.StringValue(ash.ReleaseName),
		SkipCRDs:                types.BoolValue(ash.SkipCrds),
		ValueFiles:              pie.Map(ash.ValueFiles, types.StringValue),
		Values:                  types.StringValue(ash.Values),
	}
}

type applicationHelmFileParameter struct {
	Name types.String `tfsdk:"name"`
	Path types.String `tfsdk:"path"`
}

func applicationHelmFileParameterSchemaAttribute(computed bool) schema.Attribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "File parameters for the helm template.",
		Computed:            computed,
		Optional:            !computed,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: "Name of the Helm parameters.",
					Required:            !computed,
					Computed:            computed,
				},
				"path": schema.StringAttribute{
					MarkdownDescription: "Path to the file containing the values for the Helm parameters.",
					Required:            !computed,
					Computed:            computed,
				},
			},
		},
	}
}

func newApplicationSourceHelmFileParameters(hfps []v1alpha1.HelmFileParameter) []applicationHelmFileParameter {
	if hfps == nil {
		return nil
	}

	fps := make([]applicationHelmFileParameter, len(hfps))
	for i, v := range hfps {
		fps[i] = applicationHelmFileParameter{
			Name: types.StringValue(v.Name),
			Path: types.StringValue(v.Path),
		}
	}

	return fps
}

type applicationHelmParameter struct {
	ForceString types.Bool   `tfsdk:"force_string"`
	Name        types.String `tfsdk:"name"`
	Value       types.String `tfsdk:"value"`
}

func applicationHelmParameterSchemaAttribute(computed bool) schema.Attribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Helm parameters which are passed to the helm template command upon manifest generation.",
		Computed:            computed,
		Optional:            !computed,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: "Name of the Helm parameters.",
					Computed:            computed,
					Optional:            !computed,
				},
				"value": schema.StringAttribute{
					MarkdownDescription: "Value of the Helm parameters.",
					Computed:            computed,
					Optional:            !computed,
				},
				"force_string": schema.BoolAttribute{
					MarkdownDescription: "Determines whether to tell Helm to interpret booleans and numbers as strings.",
					Computed:            computed,
					Optional:            !computed,
				},
			},
		},
	}
}

func newApplicationSourceHelmParameters(hps []v1alpha1.HelmParameter) []applicationHelmParameter {
	if hps == nil {
		return nil
	}

	ps := make([]applicationHelmParameter, len(hps))
	for i, v := range hps {
		ps[i] = applicationHelmParameter{
			ForceString: types.BoolValue(v.ForceString),
			Name:        types.StringValue(v.Name),
			Value:       types.StringValue(v.Value),
		}
	}

	return ps
}

type applicationSourceKustomize struct {
	CommonAnnotations map[string]types.String `tfsdk:"common_annotations"`
	CommonLabels      map[string]types.String `tfsdk:"common_labels"`
	Images            []types.String          `tfsdk:"images"`
	NamePrefix        types.String            `tfsdk:"name_prefix"`
	NameSuffix        types.String            `tfsdk:"name_suffix"`
	Version           types.String            `tfsdk:"version"`
}

func applicationSourceKustomizeSchemaAttribute(computed bool) schema.Attribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Kustomize specific options.",
		Computed:            computed,
		Optional:            !computed,
		Attributes: map[string]schema.Attribute{
			"name_prefix": schema.StringAttribute{
				MarkdownDescription: "Prefix appended to resources for Kustomize apps.",
				Computed:            computed,
				Optional:            !computed,
			},
			"name_suffix": schema.StringAttribute{
				MarkdownDescription: "Suffix appended to resources for Kustomize apps.",
				Computed:            computed,
				Optional:            !computed,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "Version of Kustomize to use for rendering manifests.",
				Computed:            computed,
				Optional:            !computed,
			},
			"images": schema.SetAttribute{
				MarkdownDescription: "List of Kustomize image override specifications.",
				Computed:            computed,
				Optional:            !computed,
				ElementType:         types.StringType,
			},
			"common_labels": schema.MapAttribute{
				MarkdownDescription: "List of additional labels to add to rendered manifests.",
				Computed:            computed,
				Optional:            !computed,
				ElementType:         types.StringType,
				Validators: []validator.Map{
					validators.MetadataLabels(),
				},
			},
			"common_annotations": schema.MapAttribute{
				MarkdownDescription: "List of additional annotations to add to rendered manifests.",
				Computed:            computed,
				Optional:            !computed,
				ElementType:         types.StringType,
				Validators: []validator.Map{
					validators.MetadataAnnotations(),
				},
			},
		},
	}
}

func newApplicationSourceKustomize(ask *v1alpha1.ApplicationSourceKustomize) *applicationSourceKustomize {
	if ask == nil {
		return nil
	}

	k := &applicationSourceKustomize{
		CommonAnnotations: utils.MapMap(ask.CommonAnnotations, types.StringValue),
		CommonLabels:      utils.MapMap(ask.CommonLabels, types.StringValue),
		NamePrefix:        types.StringValue(ask.NamePrefix),
		NameSuffix:        types.StringValue(ask.NameSuffix),
		Version:           types.StringValue(ask.Version),
	}

	if ask.Images != nil {
		k.Images = make([]basetypes.StringValue, len(ask.Images))
		for i, v := range ask.Images {
			k.Images[i] = types.StringValue(string(v))
		}
	}

	return k
}

type applicationSourcePlugin struct {
	Env        []applicationEnvEntry              `tfsdk:"env"`
	Name       types.String                       `tfsdk:"name"`
	Parameters []applicationSourcePluginParameter `tfsdk:"parameters"`
}

func applicationSourcePluginSchemaAttribute(computed bool) schema.Attribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Config management plugin specific options.",
		Computed:            computed,
		Optional:            !computed,
		Attributes: map[string]schema.Attribute{
			"env": applicationEnvEntriesSchemaAttribute(computed),
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the plugin. Only set the plugin name if the plugin is defined in `argocd-cm`. If the plugin is defined as a sidecar, omit the name. The plugin will be automatically matched with the Application according to the plugin's discovery rules.",
				Computed:            computed,
				Optional:            !computed,
			},
			"parameters": applicationSourcePluginParametersSchemaAttribute(computed),
		},
	}
}

func newApplicationSourcePlugin(asp *v1alpha1.ApplicationSourcePlugin) *applicationSourcePlugin {
	if asp == nil {
		return nil
	}

	return &applicationSourcePlugin{
		Env:        newApplicationEnvEntries(asp.Env),
		Name:       types.StringValue(asp.Name),
		Parameters: newApplicationSourcePluginParameters(asp.Parameters),
	}
}

type applicationEnvEntry struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func applicationEnvEntriesSchemaAttribute(computed bool) schema.Attribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Environment variables passed to the plugin.",
		Computed:            computed,
		Optional:            !computed,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: "Name of the environment variable.",
					Computed:            computed,
					Optional:            !computed,
				},
				"value": schema.StringAttribute{
					MarkdownDescription: "Value of the environment variable.",
					Computed:            computed,
					Optional:            !computed,
				},
			},
		},
	}
}

func newApplicationEnvEntries(ees []*v1alpha1.EnvEntry) []applicationEnvEntry {
	if ees == nil {
		return nil
	}

	var es []applicationEnvEntry

	for _, v := range ees {
		if v == nil {
			continue
		}

		es = append(es, applicationEnvEntry{
			Name:  types.StringValue(v.Name),
			Value: types.StringValue(v.Value),
		})
	}

	return es
}

type applicationSourcePluginParameter struct {
	Array  []types.String          `tfsdk:"array"`
	Map    map[string]types.String `tfsdk:"map"`
	Name   types.String            `tfsdk:"name"`
	String types.String            `tfsdk:"string"`
}

func applicationSourcePluginParametersSchemaAttribute(computed bool) schema.Attribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Parameters to supply to config management plugin.",
		Computed:            computed,
		Optional:            !computed,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"array": schema.ListAttribute{
					MarkdownDescription: "Value of an array type parameters.",
					Computed:            computed,
					Optional:            !computed,
					ElementType:         types.StringType,
				},
				"name": schema.StringAttribute{
					MarkdownDescription: "Name identifying a parameters.",
					Computed:            computed,
					Optional:            !computed,
				},
				"map": schema.MapAttribute{
					MarkdownDescription: "Value of a map type parameters.",
					Computed:            computed,
					Optional:            !computed,
					ElementType:         types.StringType,
				},
				"string": schema.StringAttribute{
					MarkdownDescription: "Value of a string type parameters.",
					Computed:            computed,
					Optional:            !computed,
				},
			},
		},
	}
}

func newApplicationSourcePluginParameters(aspps v1alpha1.ApplicationSourcePluginParameters) []applicationSourcePluginParameter {
	if aspps == nil {
		return nil
	}

	pps := make([]applicationSourcePluginParameter, len(aspps))

	for i, v := range aspps {
		pps[i] = applicationSourcePluginParameter{
			Array:  pie.Map(v.Array, types.StringValue),
			Map:    utils.MapMap(v.Map, types.StringValue),
			Name:   types.StringValue(v.Name),
			String: utils.OptionalString(v.String_),
		}
	}

	return pps
}

type applicationSyncPolicy struct {
	Automated   *applicationSyncPolicyAutomated `tfsdk:"automated"`
	Retry       *applicationRetryStrategy       `tfsdk:"retry"`
	SyncOptions []types.String                  `tfsdk:"sync_options"`
}

func applicationSyncPolicySchemaAttribute(computed bool) schema.Attribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Controls when and how a sync will be performed.",
		Computed:            computed,
		Optional:            !computed,
		Attributes: map[string]schema.Attribute{
			"automated": applicationSyncPolicyAutomatedSchemaAttribute(computed),
			"retry":     applicationRetryStrategySchemaAttribute(computed),
			"sync_options": schema.SetAttribute{
				MarkdownDescription: "List of sync options. More info: https://argo-cd.readthedocs.io/en/stable/user-guide/sync-options/.",
				Computed:            computed,
				Optional:            !computed,
				ElementType:         types.StringType,
			},
		},
	}
}

func newApplicationSyncPolicy(sp *v1alpha1.SyncPolicy) *applicationSyncPolicy {
	if sp == nil {
		return nil
	}

	return &applicationSyncPolicy{
		Automated:   newApplicationSyncPolicyAutomated(sp.Automated),
		Retry:       newApplicationRetryStrategy(sp.Retry),
		SyncOptions: pie.Map(sp.SyncOptions, types.StringValue),
	}
}

type applicationSyncPolicyAutomated struct {
	AllowEmpty types.Bool `tfsdk:"allow_empty"`
	Prune      types.Bool `tfsdk:"prune"`
	SelfHeal   types.Bool `tfsdk:"self_heal"`
}

func applicationSyncPolicyAutomatedSchemaAttribute(computed bool) schema.Attribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Whether to automatically keep an application synced to the target revision.",
		Computed:            computed,
		Optional:            !computed,
		Attributes: map[string]schema.Attribute{
			"allow_empty": schema.BoolAttribute{
				MarkdownDescription: "Allows apps have zero live resources.",
				Computed:            computed,
				Optional:            !computed,
			},
			"prune": schema.BoolAttribute{
				MarkdownDescription: "Whether to delete resources from the cluster that are not found in the sources anymore as part of automated sync.",
				Computed:            computed,
				Optional:            !computed,
			},
			"self_heal": schema.BoolAttribute{
				MarkdownDescription: "Whether to revert resources back to their desired state upon modification in the cluster.",
				Computed:            computed,
				Optional:            !computed,
			},
		},
	}
}

func newApplicationSyncPolicyAutomated(spa *v1alpha1.SyncPolicyAutomated) *applicationSyncPolicyAutomated {
	if spa == nil {
		return nil
	}

	return &applicationSyncPolicyAutomated{
		AllowEmpty: types.BoolValue(spa.AllowEmpty),
		Prune:      types.BoolValue(spa.Prune),
		SelfHeal:   types.BoolValue(spa.SelfHeal),
	}
}

type applicationRetryStrategy struct {
	Limit   types.Int64         `tfsdk:"limit"`
	Backoff *applicationBackoff `tfsdk:"backoff"`
}

func applicationRetryStrategySchemaAttribute(computed bool) schema.Attribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Controls failed sync retry behavior.",
		Computed:            computed,
		Optional:            !computed,
		Attributes: map[string]schema.Attribute{
			"backoff": applicationBackoffSchemaAttribute(computed),
			"limit": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of attempts for retrying a failed sync. If set to 0, no retries will be performed.",
				Computed:            computed,
				Optional:            !computed,
			},
		},
	}
}

func newApplicationRetryStrategy(rs *v1alpha1.RetryStrategy) *applicationRetryStrategy {
	if rs == nil {
		return nil
	}

	return &applicationRetryStrategy{
		Backoff: newApplicationBackoff(rs.Backoff),
		Limit:   types.Int64Value(rs.Limit),
	}
}

type applicationBackoff struct {
	Duration    types.String `tfsdk:"duration"`
	Factor      types.Int64  `tfsdk:"factor"`
	MaxDuration types.String `tfsdk:"max_duration"`
}

func applicationBackoffSchemaAttribute(computed bool) schema.Attribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Controls how to backoff on subsequent retries of failed syncs.",
		Computed:            computed,
		Optional:            !computed,
		Attributes: map[string]schema.Attribute{
			"duration": schema.StringAttribute{
				MarkdownDescription: "Duration is the amount to back off. Default unit is seconds, but could also be a duration (e.g. `2m`, `1h`), as a string.",
				Computed:            computed,
				Optional:            !computed,
			},
			"factor": schema.Int64Attribute{
				MarkdownDescription: "Factor to multiply the base duration after each failed retry.",
				Computed:            computed,
				Optional:            !computed,
			},
			"max_duration": schema.StringAttribute{
				MarkdownDescription: "Maximum amount of time allowed for the backoff strategy. Default unit is seconds, but could also be a duration (e.g. `2m`, `1h`), as a string.",
				Computed:            computed,
				Optional:            !computed,
			},
		},
	}
}

func newApplicationBackoff(b *v1alpha1.Backoff) *applicationBackoff {
	if b == nil {
		return nil
	}

	return &applicationBackoff{
		Duration:    types.StringValue(b.Duration),
		Factor:      utils.OptionalInt64(b.Factor),
		MaxDuration: types.StringValue(b.MaxDuration),
	}
}

type applicationStatus struct {
	Conditions     []applicationCondition      `tfsdk:"conditions"`
	Health         applicationHealthStatus     `tfsdk:"health"`
	OperationState *applicationOperationState  `tfsdk:"operation_state"`
	ReconciledAt   types.String                `tfsdk:"reconciled_at"`
	Resources      []applicationResourceStatus `tfsdk:"resources"`
	Summary        applicationSummary          `tfsdk:"summary"`
	Sync           applicationSyncStatus       `tfsdk:"sync"`
}

func applicationStatusSchemaAttribute() schema.Attribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Status information for the application.",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"conditions":      applicationConditionSchemaAttribute(),
			"health":          applicationHealthStatusSchemaAttribute(),
			"operation_state": applicationOperationStateSchemaAttribute(),
			"reconciled_at": schema.StringAttribute{
				MarkdownDescription: "When the application state was reconciled using the latest git version.",
				Computed:            true,
			},
			"resources": applicationResourceStatusSchemaAttribute(),
			"summary":   applicationSummarySchemaAttribute(),
			"sync":      applicationSyncStatusSchemaAttribute(),
		},
	}
}

func newApplicationStatus(as v1alpha1.ApplicationStatus) *applicationStatus {
	return &applicationStatus{
		Conditions:     newApplicationConditions(as.Conditions),
		Health:         *newApplicationHealthStatus(&as.Health),
		OperationState: newApplicationOperationState(as.OperationState),
		ReconciledAt:   types.StringValue(as.ReconciledAt.String()),
		Resources:      newApplicationResourceStatuses(as.Resources),
		Summary:        newApplicationSummary(as.Summary),
		Sync:           newApplicationSyncStatus(as.Sync),
	}
}

type applicationCondition struct {
	Message            types.String `tfsdk:"message"`
	LastTransitionTime types.String `tfsdk:"last_transition_time"`
	Type               types.String `tfsdk:"type"`
}

func applicationConditionSchemaAttribute() schema.Attribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "List of currently observed application conditions.",
		Computed:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"message": schema.StringAttribute{
					MarkdownDescription: "Human-readable message indicating details about condition.",
					Computed:            true,
				},
				"last_transition_time": schema.StringAttribute{
					MarkdownDescription: "The time the condition was last observed.",
					Computed:            true,
				},
				"type": schema.StringAttribute{
					MarkdownDescription: "Application condition type.",
					Computed:            true,
				},
			},
		},
	}
}

func newApplicationConditions(acs []v1alpha1.ApplicationCondition) []applicationCondition {
	if acs == nil {
		return nil
	}

	cs := make([]applicationCondition, len(acs))

	for i, v := range acs {
		cs[i] = applicationCondition{
			LastTransitionTime: utils.OptionalTimeString(v.LastTransitionTime),
			Message:            types.StringValue(v.Message),
			Type:               types.StringValue(v.Type),
		}
	}

	return cs
}

type applicationHealthStatus struct {
	Message types.String `tfsdk:"message"`
	Status  types.String `tfsdk:"status"`
}

func applicationHealthStatusSchemaAttribute() schema.Attribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Application's current health status.",
		Computed:            true,
		Attributes:          applicationHealthStatusSchemaAttributes(),
	}
}

func applicationHealthStatusSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"message": schema.StringAttribute{
			MarkdownDescription: "Human-readable informational message describing the health status.",
			Computed:            true,
		},
		"status": schema.StringAttribute{
			MarkdownDescription: "Status code of the application or resource.",
			Computed:            true,
		},
	}
}

func newApplicationHealthStatus(hs *v1alpha1.HealthStatus) *applicationHealthStatus {
	if hs == nil {
		return nil
	}

	return &applicationHealthStatus{
		Message: types.StringValue(hs.Message),
		Status:  types.StringValue(string(hs.Status)),
	}
}

type applicationOperationState struct {
	FinishedAt types.String `tfsdk:"finished_at"`
	Message    types.String `tfsdk:"message"`
	Phase      types.String `tfsdk:"phase"`
	RetryCount types.Int64  `tfsdk:"retry_count"`
	StartedAt  types.String `tfsdk:"started_at"`
}

func applicationOperationStateSchemaAttribute() schema.Attribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Information about any ongoing operations, such as a sync.",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"finished_at": schema.StringAttribute{
				MarkdownDescription: "Time of operation completion.",
				Computed:            true,
			},
			"message": schema.StringAttribute{
				MarkdownDescription: "Any pertinent messages when attempting to perform operation (typically errors).",
				Computed:            true,
			},
			"phase": schema.StringAttribute{
				MarkdownDescription: "The current phase of the operation.",
				Computed:            true,
			},
			"retry_count": schema.Int64Attribute{
				MarkdownDescription: "Count of operation retries.",
				Computed:            true,
			},
			"started_at": schema.StringAttribute{
				MarkdownDescription: "Time of operation start.",
				Computed:            true,
			},
		},
	}
}

func newApplicationOperationState(os *v1alpha1.OperationState) *applicationOperationState {
	if os == nil {
		return nil
	}

	return &applicationOperationState{
		FinishedAt: utils.OptionalTimeString(os.FinishedAt),
		Message:    types.StringValue(os.Message),
		Phase:      types.StringValue(string(os.Phase)),
		RetryCount: types.Int64Value(os.RetryCount),
		StartedAt:  types.StringValue(os.StartedAt.String()),
	}
}

type applicationResourceStatus struct {
	Group           types.String             `tfsdk:"group"`
	Health          *applicationHealthStatus `tfsdk:"health"`
	Hook            types.Bool               `tfsdk:"hook"`
	Kind            types.String             `tfsdk:"kind"`
	Name            types.String             `tfsdk:"name"`
	Namespace       types.String             `tfsdk:"namespace"`
	RequiresPruning types.Bool               `tfsdk:"requires_pruning"`
	Status          types.String             `tfsdk:"status"`
	SyncWave        types.Int64              `tfsdk:"sync_wave"`
	Version         types.String             `tfsdk:"version"`
}

func applicationResourceStatusSchemaAttribute() schema.Attribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "List of Kubernetes resources managed by this application.",
		Computed:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"group": schema.StringAttribute{
					MarkdownDescription: "The Kubernetes resource Group.",
					Computed:            true,
				},
				"health": schema.SingleNestedAttribute{
					MarkdownDescription: "Resource health status.",
					Computed:            true,
					Attributes:          applicationHealthStatusSchemaAttributes(),
				},
				"kind": schema.StringAttribute{
					MarkdownDescription: "The Kubernetes resource Kind.",
					Computed:            true,
				},
				"hook": schema.BoolAttribute{
					MarkdownDescription: "Indicates whether or not this resource has a hook annotation.",
					Computed:            true,
				},
				"name": schema.StringAttribute{
					MarkdownDescription: "The Kubernetes resource Name.",
					Computed:            true,
				},
				"namespace": schema.StringAttribute{
					MarkdownDescription: "The Kubernetes resource Namespace.",
					Computed:            true,
				},
				"requires_pruning": schema.BoolAttribute{
					MarkdownDescription: "Indicates if the resources requires pruning or not.",
					Computed:            true,
				},
				"status": schema.StringAttribute{
					MarkdownDescription: "Resource sync status.",
					Computed:            true,
				},
				"sync_wave": schema.Int64Attribute{
					MarkdownDescription: "Sync wave.",
					Computed:            true,
				},
				"version": schema.StringAttribute{
					MarkdownDescription: "The Kubernetes resource Version.",
					Computed:            true,
				},
			},
		},
	}
}

func newApplicationResourceStatuses(rss []v1alpha1.ResourceStatus) []applicationResourceStatus {
	if rss == nil {
		return nil
	}

	rs := make([]applicationResourceStatus, len(rss))

	for i, v := range rss {
		rs[i] = applicationResourceStatus{
			Group:           types.StringValue(v.Group),
			Health:          newApplicationHealthStatus(v.Health),
			Hook:            types.BoolValue(v.Hook),
			Kind:            types.StringValue(v.Kind),
			Name:            types.StringValue(v.Name),
			Namespace:       types.StringValue(v.Namespace),
			RequiresPruning: types.BoolValue(v.RequiresPruning),
			Status:          types.StringValue(string(v.Status)),
			SyncWave:        types.Int64Value(v.SyncWave),
			Version:         types.StringValue(v.Version),
		}
	}

	return rs
}

type applicationSummary struct {
	ExternalURLs []types.String `tfsdk:"external_urls"`
	Images       []types.String `tfsdk:"images"`
}

func applicationSummarySchemaAttribute() schema.Attribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "List of URLs and container images used by this application.",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"external_urls": schema.ListAttribute{
				MarkdownDescription: "All external URLs of application child resources.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"images": schema.ListAttribute{
				MarkdownDescription: "All images of application child resources.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func newApplicationSummary(as v1alpha1.ApplicationSummary) applicationSummary {
	return applicationSummary{
		ExternalURLs: pie.Map(as.ExternalURLs, types.StringValue),
		Images:       pie.Map(as.Images, types.StringValue),
	}
}

type applicationSyncStatus struct {
	Revisions []types.String `tfsdk:"revisions"`
	Status    types.String   `tfsdk:"status"`
}

func applicationSyncStatusSchemaAttribute() schema.Attribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Application's current sync status",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"revisions": schema.ListAttribute{
				MarkdownDescription: "Information about the revision(s) the comparison has been performed to.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Sync state of the comparison.",
				Computed:            true,
			},
		},
	}
}

func newApplicationSyncStatus(ss v1alpha1.SyncStatus) applicationSyncStatus {
	ass := applicationSyncStatus{
		Status: types.StringValue(string(ss.Status)),
	}

	if ss.Revision != "" {
		ass.Revisions = append(ass.Revisions, types.StringValue(ss.Revision))
	}

	if len(ss.Revisions) > 0 {
		ass.Revisions = append(ass.Revisions, pie.Map(ss.Revisions, types.StringValue)...)
	}

	return ass
}

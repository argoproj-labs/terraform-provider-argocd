package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/argoproj-labs/terraform-provider-argocd/internal/diagnostics"
	"github.com/argoproj-labs/terraform-provider-argocd/internal/features"
	argocdSync "github.com/argoproj-labs/terraform-provider-argocd/internal/sync"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/project"
	"github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &projectResource{}

func NewProjectResource() resource.Resource {
	return &projectResource{}
}

type projectResource struct {
	si *ServerInterface
}

func (r *projectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *projectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages [projects](https://argo-cd.readthedocs.io/en/stable/user-guide/projects/) within ArgoCD.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Project identifier",
				Computed:    true,
			},
		},
		Blocks: projectSchemaBlocks(),
	}
}

func (r *projectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	si, ok := req.ProviderData.(*ServerInterface)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *ServerInterface, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.si = si
}

func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data projectModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	// Initialize API clients
	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	// Check for errors before proceeding
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that spec list is not empty
	if len(data.Spec) == 0 {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"spec block is required but not provided",
		)

		return
	}

	// Convert model to ArgoCD project
	objectMeta, spec, diags := expandProject(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectName := objectMeta.Name

	// Check feature support
	if !r.si.IsFeatureSupported(features.ProjectSourceNamespaces) && len(data.Spec[0].SourceNamespaces) > 0 {
		resp.Diagnostics.Append(diagnostics.FeatureNotSupported(features.ProjectSourceNamespaces)...)
		return
	}

	if !r.si.IsFeatureSupported(features.ProjectDestinationServiceAccounts) && len(data.Spec[0].DestinationServiceAccount) > 0 {
		resp.Diagnostics.Append(diagnostics.FeatureNotSupported(features.ProjectDestinationServiceAccounts)...)
		return
	}

	// Get or create project mutex safely
	projectMutex := argocdSync.GetProjectMutex(projectName)
	projectMutex.Lock()
	defer projectMutex.Unlock()

	// Check if project already exists
	p, err := r.si.ProjectClient.Get(ctx, &project.ProjectQuery{
		Name: projectName,
	})
	if err != nil && !strings.Contains(err.Error(), "NotFound") {
		resp.Diagnostics.Append(diagnostics.ArgoCDAPIError("get", "project", projectName, err)...)
		return
	} else if p != nil {
		switch p.DeletionTimestamp {
		case nil:
		default:
			// Pre-existing project is still in Kubernetes soft deletion queue
			if p.DeletionGracePeriodSeconds != nil {
				time.Sleep(time.Duration(*p.DeletionGracePeriodSeconds) * time.Second)
			}
		}
	}

	// Create project
	p, err = r.si.ProjectClient.Create(ctx, &project.ProjectCreateRequest{
		Project: &v1alpha1.AppProject{
			ObjectMeta: objectMeta,
			Spec:       spec,
		},
		Upsert: false,
	})

	if err != nil {
		resp.Diagnostics.Append(diagnostics.ArgoCDAPIError("create", "project", projectName, err)...)
		return
	} else if p == nil {
		resp.Diagnostics.AddError(
			"Project Creation Failed",
			fmt.Sprintf("project %s could not be created: unknown reason", projectName),
		)

		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created project %s", projectName))

	// Parse response and store state
	projectData := newProject(p)
	projectData.ID = types.StringValue(projectName)
	resp.Diagnostics.Append(resp.State.Set(ctx, projectData)...)
}

func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data projectModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	// Check for errors before proceeding
	if resp.Diagnostics.HasError() {
		return
	}

	projectName := data.Metadata[0].Name.ValueString()

	// Get or create project mutex safely
	projectMutex := argocdSync.GetProjectMutex(projectName)
	projectMutex.RLock()
	defer projectMutex.RUnlock()

	r.readUnsafe(ctx, data, projectName, resp)
}

func (r *projectResource) readUnsafe(ctx context.Context, data projectModel, projectName string, resp *resource.ReadResponse) {
	// Initialize API clients
	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	// Check for errors before proceeding
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that metadata list is not empty
	if len(data.Metadata) == 0 {
		resp.Diagnostics.AddError(
			"Invalid State",
			"metadata block is missing from state",
		)

		return
	}

	p, err := r.si.ProjectClient.Get(ctx, &project.ProjectQuery{
		Name: projectName,
	})

	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.Append(diagnostics.ArgoCDAPIError("read", "project", projectName, err)...)

		return
	}

	// Save updated data into Terraform state
	projectData := newProject(p)
	projectData.ID = types.StringValue(projectName)
	resp.Diagnostics.Append(resp.State.Set(ctx, projectData)...)
}

func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data projectModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	// Initialize API clients
	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	// Check for errors before proceeding
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that spec list is not empty
	if len(data.Spec) == 0 {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"spec block is required but not provided",
		)

		return
	}

	// Convert model to ArgoCD project
	objectMeta, spec, diags := expandProject(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectName := objectMeta.Name

	// Check feature support
	if !r.si.IsFeatureSupported(features.ProjectSourceNamespaces) && len(data.Spec[0].SourceNamespaces) > 0 {
		resp.Diagnostics.Append(diagnostics.FeatureNotSupported(features.ProjectSourceNamespaces)...)
		return
	}

	if !r.si.IsFeatureSupported(features.ProjectDestinationServiceAccounts) && len(data.Spec[0].DestinationServiceAccount) > 0 {
		resp.Diagnostics.Append(diagnostics.FeatureNotSupported(features.ProjectDestinationServiceAccounts)...)
		return
	}

	// Get or create project mutex safely
	projectMutex := argocdSync.GetProjectMutex(projectName)
	projectMutex.Lock()
	defer projectMutex.Unlock()

	// Get current project
	p, err := r.si.ProjectClient.Get(ctx, &project.ProjectQuery{
		Name: projectName,
	})
	if err != nil {
		resp.Diagnostics.Append(diagnostics.ArgoCDAPIError("get", "project", projectName, err)...)
		return
	}

	// Preserve preexisting JWTs for managed roles
	roles := expandProjectRoles(ctx, data.Spec[0].Role)
	for _, r := range roles {
		var pr *v1alpha1.ProjectRole

		var i int

		pr, i, err = p.GetRoleByName(r.Name)
		if err != nil {
			// i == -1 means the role does not exist and was recently added
			if i != -1 {
				resp.Diagnostics.AddError(
					"Project Role Retrieval Failed",
					fmt.Sprintf("project role %s could not be retrieved: %s", r.Name, err.Error()),
				)

				return
			}
		} else {
			// Only preserve preexisting JWTs for managed roles if we found an existing matching project
			spec.Roles[i].JWTTokens = pr.JWTTokens
		}
	}

	// Update project
	projectRequest := &project.ProjectUpdateRequest{
		Project: &v1alpha1.AppProject{
			ObjectMeta: objectMeta,
			Spec:       spec,
		},
	}

	// Kubernetes API requires providing the up-to-date correct ResourceVersion for updates
	projectRequest.Project.ResourceVersion = p.ResourceVersion

	_, err = r.si.ProjectClient.Update(ctx, projectRequest)
	if err != nil {
		resp.Diagnostics.Append(diagnostics.ArgoCDAPIError("update", "project", projectName, err)...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated project %s", projectName))

	// Read updated resource
	readReq := resource.ReadRequest{State: req.State}
	readResp := resource.ReadResponse{State: resp.State, Diagnostics: resp.Diagnostics}

	var updatedData projectModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(readReq.State.Get(ctx, &updatedData)...)

	r.readUnsafe(ctx, updatedData, projectName, &readResp)
	resp.State = readResp.State
	resp.Diagnostics = readResp.Diagnostics
}

func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data projectModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	// Initialize API clients
	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	// Check for errors before proceeding
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that metadata list is not empty
	if len(data.Metadata) == 0 {
		resp.Diagnostics.AddError(
			"Invalid State",
			"metadata block is missing from state",
		)

		return
	}

	projectName := data.Metadata[0].Name.ValueString()

	// Get or create project mutex safely
	projectMutex := argocdSync.GetProjectMutex(projectName)
	projectMutex.Lock()
	defer projectMutex.Unlock()

	_, err := r.si.ProjectClient.Delete(ctx, &project.ProjectQuery{Name: projectName})

	if err != nil && !strings.Contains(err.Error(), "NotFound") {
		resp.Diagnostics.Append(diagnostics.ArgoCDAPIError("delete", "project", projectName, err)...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted project %s", projectName))
}

func (r *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Initialize API clients
	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Try to get the project from ArgoCD to verify it exists
	p, err := r.si.ProjectClient.Get(ctx, &project.ProjectQuery{
		Name: req.ID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			resp.Diagnostics.AddError(
				"Cannot import non-existent remote object",
				fmt.Sprintf("Project %s does not exist in ArgoCD", req.ID),
			)

			return
		}

		resp.Diagnostics.Append(diagnostics.ArgoCDAPIError("get", "project", req.ID, err)...)

		return
	}

	// If project exists, populate the state with the full project data
	projectData := newProject(p)
	projectData.ID = types.StringValue(req.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, projectData)...)
}

// expandProject converts the Terraform model to ArgoCD API types
func expandProject(ctx context.Context, data *projectModel) (metav1.ObjectMeta, v1alpha1.AppProjectSpec, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Validate that metadata list is not empty
	if len(data.Metadata) == 0 {
		diags.AddError(
			"Invalid Configuration",
			"metadata block is required but not provided",
		)

		return metav1.ObjectMeta{}, v1alpha1.AppProjectSpec{}, diags
	}

	// Validate that spec list is not empty
	if len(data.Spec) == 0 {
		diags.AddError(
			"Invalid Configuration",
			"spec block is required but not provided",
		)

		return metav1.ObjectMeta{}, v1alpha1.AppProjectSpec{}, diags
	}

	objectMeta := metav1.ObjectMeta{
		Name:      data.Metadata[0].Name.ValueString(),
		Namespace: data.Metadata[0].Namespace.ValueString(),
	}

	if len(data.Metadata[0].Labels) > 0 {
		labels := make(map[string]string)
		for k, v := range data.Metadata[0].Labels {
			labels[k] = v.ValueString()
		}

		objectMeta.Labels = labels
	}

	if len(data.Metadata[0].Annotations) > 0 {
		annotations := make(map[string]string)
		for k, v := range data.Metadata[0].Annotations {
			annotations[k] = v.ValueString()
		}

		objectMeta.Annotations = annotations
	}

	spec := v1alpha1.AppProjectSpec{}

	if !data.Spec[0].Description.IsNull() {
		spec.Description = data.Spec[0].Description.ValueString()
	}

	// Convert source repos
	for _, repo := range data.Spec[0].SourceRepos {
		spec.SourceRepos = append(spec.SourceRepos, repo.ValueString())
	}

	// Convert signature keys
	for _, key := range data.Spec[0].SignatureKeys {
		spec.SignatureKeys = append(spec.SignatureKeys, v1alpha1.SignatureKey{KeyID: key.ValueString()})
	}

	// Convert source namespaces
	for _, ns := range data.Spec[0].SourceNamespaces {
		spec.SourceNamespaces = append(spec.SourceNamespaces, ns.ValueString())
	}

	// Convert destinations
	for _, dest := range data.Spec[0].Destination {
		d := v1alpha1.ApplicationDestination{
			Namespace: dest.Namespace.ValueString(),
		}
		if !dest.Server.IsNull() {
			d.Server = dest.Server.ValueString()
		}

		if !dest.Name.IsNull() {
			d.Name = dest.Name.ValueString()
		}

		spec.Destinations = append(spec.Destinations, d)
	}

	// Convert destination service accounts
	for _, dsa := range data.Spec[0].DestinationServiceAccount {
		d := v1alpha1.ApplicationDestinationServiceAccount{
			DefaultServiceAccount: dsa.DefaultServiceAccount.ValueString(),
			Server:                dsa.Server.ValueString(),
		}
		if !dsa.Namespace.IsNull() {
			d.Namespace = dsa.Namespace.ValueString()
		}

		spec.DestinationServiceAccounts = append(spec.DestinationServiceAccounts, d)
	}

	// Convert cluster resource blacklist
	for _, gk := range data.Spec[0].ClusterResourceBlacklist {
		spec.ClusterResourceBlacklist = append(spec.ClusterResourceBlacklist, metav1.GroupKind{
			Group: gk.Group.ValueString(),
			Kind:  gk.Kind.ValueString(),
		})
	}

	// Convert cluster resource whitelist
	for _, gk := range data.Spec[0].ClusterResourceWhitelist {
		spec.ClusterResourceWhitelist = append(spec.ClusterResourceWhitelist, metav1.GroupKind{
			Group: gk.Group.ValueString(),
			Kind:  gk.Kind.ValueString(),
		})
	}

	// Convert namespace resource blacklist
	for _, gk := range data.Spec[0].NamespaceResourceBlacklist {
		spec.NamespaceResourceBlacklist = append(spec.NamespaceResourceBlacklist, metav1.GroupKind{
			Group: gk.Group.ValueString(),
			Kind:  gk.Kind.ValueString(),
		})
	}

	// Convert namespace resource whitelist
	for _, gk := range data.Spec[0].NamespaceResourceWhitelist {
		spec.NamespaceResourceWhitelist = append(spec.NamespaceResourceWhitelist, metav1.GroupKind{
			Group: gk.Group.ValueString(),
			Kind:  gk.Kind.ValueString(),
		})
	}

	// Convert orphaned resources
	if len(data.Spec[0].OrphanedResources) > 0 {
		or := data.Spec[0].OrphanedResources[0]
		spec.OrphanedResources = &v1alpha1.OrphanedResourcesMonitorSettings{}

		if !or.Warn.IsNull() {
			spec.OrphanedResources.Warn = or.Warn.ValueBoolPointer()
		}

		for _, ignore := range or.Ignore {
			i := v1alpha1.OrphanedResourceKey{
				Group: ignore.Group.ValueString(),
				Kind:  ignore.Kind.ValueString(),
			}
			if !ignore.Name.IsNull() {
				i.Name = ignore.Name.ValueString()
			}

			spec.OrphanedResources.Ignore = append(spec.OrphanedResources.Ignore, i)
		}
	}

	// Convert roles
	spec.Roles = expandProjectRoles(ctx, data.Spec[0].Role)

	// Convert sync windows
	for _, sw := range data.Spec[0].SyncWindow {
		window := v1alpha1.SyncWindow{}
		if !sw.Duration.IsNull() {
			window.Duration = sw.Duration.ValueString()
		}

		if !sw.Kind.IsNull() {
			window.Kind = sw.Kind.ValueString()
		}

		if !sw.ManualSync.IsNull() {
			window.ManualSync = sw.ManualSync.ValueBool()
		}

		if !sw.Schedule.IsNull() {
			window.Schedule = sw.Schedule.ValueString()
		}

		if !sw.Timezone.IsNull() {
			window.TimeZone = sw.Timezone.ValueString()
		}

		for _, app := range sw.Applications {
			window.Applications = append(window.Applications, app.ValueString())
		}

		for _, cluster := range sw.Clusters {
			window.Clusters = append(window.Clusters, cluster.ValueString())
		}

		for _, ns := range sw.Namespaces {
			window.Namespaces = append(window.Namespaces, ns.ValueString())
		}

		spec.SyncWindows = append(spec.SyncWindows, &window)
	}

	return objectMeta, spec, diags
}

// expandProjectRoles converts project role models to ArgoCD API types
func expandProjectRoles(_ context.Context, roles []projectRoleModel) []v1alpha1.ProjectRole {
	var result []v1alpha1.ProjectRole

	for _, role := range roles {
		pr := v1alpha1.ProjectRole{
			Name: role.Name.ValueString(),
		}

		if !role.Description.IsNull() {
			pr.Description = role.Description.ValueString()
		}

		for _, policy := range role.Policies {
			pr.Policies = append(pr.Policies, policy.ValueString())
		}

		for _, group := range role.Groups {
			pr.Groups = append(pr.Groups, group.ValueString())
		}

		result = append(result, pr)
	}

	return result
}

package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/argoproj-labs/terraform-provider-argocd/internal/diagnostics"
	"github.com/argoproj-labs/terraform-provider-argocd/internal/sync"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/repocreds"
	"github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &repositoryCredentialsResource{}
var _ resource.ResourceWithImportState = &repositoryCredentialsResource{}

func NewRepositoryCredentialsResource() resource.Resource {
	return &repositoryCredentialsResource{}
}

// repositoryCredentialsResource defines the resource implementation.
type repositoryCredentialsResource struct {
	si *ServerInterface
}

func (r *repositoryCredentialsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository_credentials"
}

func (r *repositoryCredentialsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages [repository credentials](https://argo-cd.readthedocs.io/en/stable/user-guide/private-repositories/#credentials) within ArgoCD.\n\n" +
			"**Note**: due to restrictions in the ArgoCD API the provider is unable to track drift in this resource to fields other than `username`. I.e. the " +
			"provider is unable to detect changes to repository credentials that are made outside of Terraform (e.g. manual updates to the underlying Kubernetes " +
			"Secrets).",
		Attributes: repositoryCredentialsSchemaAttributes(),
	}
}

func (r *repositoryCredentialsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *repositoryCredentialsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data repositoryCredentialsModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	// Initialize API clients
	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	// Check for errors before proceeding
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to API model
	creds, err := data.toAPIModel()
	if err != nil {
		resp.Diagnostics.AddError("Failed to convert repository credentials model", err.Error())
		return
	}

	// Create repository credentials
	sync.RepositoryCredentialsMutex.Lock()
	createdCreds, err := r.si.RepoCredsClient.CreateRepositoryCredentials(
		ctx,
		&repocreds.RepoCredsCreateRequest{
			Creds:  creds,
			Upsert: false,
		},
	)
	sync.RepositoryCredentialsMutex.Unlock()

	if err != nil {
		resp.Diagnostics.Append(diagnostics.ArgoCDAPIError("create", "repository credentials", creds.URL, err)...)
		return
	}

	// Set the ID from the created credentials
	data.ID = types.StringValue(createdCreds.URL)

	// Update the model with the created credentials data
	result := data // Start with the original data to preserve all fields
	result.ID = types.StringValue(createdCreds.URL)
	result.URL = types.StringValue(createdCreds.URL)

	// Only update fields that are returned by the API
	if createdCreds.Username != "" {
		result.Username = types.StringValue(createdCreds.Username)
	}

    result.UseAzureWorkloadIdentity =  types.BoolValue(createdCreds.UseAzureWorkloadIdentity)
	result.EnableOCI = types.BoolValue(createdCreds.EnableOCI)

	// Update computed fields if available
	if createdCreds.TLSClientCertData != "" {
		result.TLSClientCertData = types.StringValue(createdCreds.TLSClientCertData)
	}

	if createdCreds.GitHubAppEnterpriseBaseURL != "" {
		result.GitHubAppEnterpriseBaseURL = types.StringValue(createdCreds.GitHubAppEnterpriseBaseURL)
	}

	// GitHub App ID conversion
	if createdCreds.GithubAppId > 0 {
		result.GitHubAppID = types.StringValue(strconv.FormatInt(createdCreds.GithubAppId, 10))
	}

	// GitHub App Installation ID conversion
	if createdCreds.GithubAppInstallationId > 0 {
		result.GitHubAppInstallationID = types.StringValue(strconv.FormatInt(createdCreds.GithubAppInstallationId, 10))
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)

	tflog.Trace(ctx, fmt.Sprintf("created repository credentials %s", result.ID.ValueString()))
}

func (r *repositoryCredentialsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data repositoryCredentialsModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	// Initialize API clients
	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	// Check for errors before proceeding
	if resp.Diagnostics.HasError() {
		return
	}

	// Read repository credentials from API
	creds, diags := r.readRepositoryCredentials(ctx, data.ID.ValueString())
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If credentials were not found, remove from state
	if creds == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update the model with the read credentials data
	result := data // Start with the original data to preserve all fields
	result.ID = types.StringValue(creds.URL)
	result.URL = types.StringValue(creds.URL)

	// Only update fields that are returned by the API
	if creds.Username != "" {
		result.Username = types.StringValue(creds.Username)
	}

    result.UseAzureWorkloadIdentity =  types.BoolValue(creds.UseAzureWorkloadIdentity)
	result.EnableOCI = types.BoolValue(creds.EnableOCI)

	// Update computed fields if available
	if creds.TLSClientCertData != "" {
		result.TLSClientCertData = types.StringValue(creds.TLSClientCertData)
	}

	if creds.GitHubAppEnterpriseBaseURL != "" {
		result.GitHubAppEnterpriseBaseURL = types.StringValue(creds.GitHubAppEnterpriseBaseURL)
	}

	// GitHub App ID conversion
	if creds.GithubAppId > 0 {
		result.GitHubAppID = types.StringValue(strconv.FormatInt(creds.GithubAppId, 10))
	}

	// GitHub App Installation ID conversion
	if creds.GithubAppInstallationId > 0 {
		result.GitHubAppInstallationID = types.StringValue(strconv.FormatInt(creds.GithubAppInstallationId, 10))
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *repositoryCredentialsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data repositoryCredentialsModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	// Initialize API clients
	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	// Check for errors before proceeding
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to API model
	creds, err := data.toAPIModel()
	if err != nil {
		resp.Diagnostics.AddError("Failed to convert repository credentials model", err.Error())
		return
	}

	// Update repository credentials
	sync.RepositoryCredentialsMutex.Lock()
	updatedCreds, err := r.si.RepoCredsClient.UpdateRepositoryCredentials(
		ctx,
		&repocreds.RepoCredsUpdateRequest{Creds: creds},
	)
	sync.RepositoryCredentialsMutex.Unlock()

	if err != nil {
		resp.Diagnostics.Append(diagnostics.ArgoCDAPIError("update", "repository credentials", creds.URL, err)...)
		return
	}

	// Set the ID from the updated credentials
	data.ID = types.StringValue(updatedCreds.URL)

	// Update the model with the updated credentials data
	result := data // Start with the original data to preserve all fields
	result.ID = types.StringValue(updatedCreds.URL)
	result.URL = types.StringValue(updatedCreds.URL)

	// Only update fields that are returned by the API
	if updatedCreds.Username != "" {
		result.Username = types.StringValue(updatedCreds.Username)
	}

    result.UseAzureWorkloadIdentity =  types.BoolValue(updatedCreds.UseAzureWorkloadIdentity)
	result.EnableOCI = types.BoolValue(updatedCreds.EnableOCI)

	// Update computed fields if available
	if updatedCreds.TLSClientCertData != "" {
		result.TLSClientCertData = types.StringValue(updatedCreds.TLSClientCertData)
	}

	if updatedCreds.GitHubAppEnterpriseBaseURL != "" {
		result.GitHubAppEnterpriseBaseURL = types.StringValue(updatedCreds.GitHubAppEnterpriseBaseURL)
	}

	// GitHub App ID conversion
	if updatedCreds.GithubAppId > 0 {
		result.GitHubAppID = types.StringValue(strconv.FormatInt(updatedCreds.GithubAppId, 10))
	}

	// GitHub App Installation ID conversion
	if updatedCreds.GithubAppInstallationId > 0 {
		result.GitHubAppInstallationID = types.StringValue(strconv.FormatInt(updatedCreds.GithubAppInstallationId, 10))
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)

	tflog.Trace(ctx, fmt.Sprintf("updated repository credentials %s", result.ID.ValueString()))
}

func (r *repositoryCredentialsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data repositoryCredentialsModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	// Initialize API clients
	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	// Check for errors before proceeding
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete repository credentials
	sync.RepositoryCredentialsMutex.Lock()
	_, err := r.si.RepoCredsClient.DeleteRepositoryCredentials(
		ctx,
		&repocreds.RepoCredsDeleteRequest{Url: data.ID.ValueString()},
	)
	sync.RepositoryCredentialsMutex.Unlock()

	if err != nil {
		if !strings.Contains(err.Error(), "NotFound") {
			resp.Diagnostics.Append(diagnostics.ArgoCDAPIError("delete", "repository credentials", data.ID.ValueString(), err)...)
			return
		}
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted repository credentials %s", data.ID.ValueString()))
}

func (r *repositoryCredentialsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *repositoryCredentialsResource) readRepositoryCredentials(ctx context.Context, url string) (*v1alpha1.RepoCreds, diag.Diagnostics) {
	var diags diag.Diagnostics

	sync.RepositoryCredentialsMutex.RLock()
	defer sync.RepositoryCredentialsMutex.RUnlock()

	credsList, err := r.si.RepoCredsClient.ListRepositoryCredentials(ctx, &repocreds.RepoCredsQuery{
		Url: url,
	})

	if err != nil {
		diags.Append(diagnostics.ArgoCDAPIError("read", "repository credentials", url, err)...)
		return nil, diags
	}

	if credsList == nil || len(credsList.Items) == 0 {
		// Repository credentials have been deleted out-of-band
		return nil, diags
	}

	// Find the specific credentials by URL
	for _, creds := range credsList.Items {
		if creds.URL == url {
			return &creds, diags
		}
	}

	// Credentials not found
	return nil, diags
}

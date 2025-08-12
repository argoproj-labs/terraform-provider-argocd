package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/argoproj-labs/terraform-provider-argocd/internal/diagnostics"
	"github.com/argoproj-labs/terraform-provider-argocd/internal/sync"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/repository"
	"github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &repositoryResource{}
var _ resource.ResourceWithImportState = &repositoryResource{}

func NewRepositoryResource() resource.Resource {
	return &repositoryResource{}
}

// repositoryResource defines the resource implementation.
type repositoryResource struct {
	si *ServerInterface
}

func (r *repositoryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository"
}

func (r *repositoryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages [repositories](https://argo-cd.readthedocs.io/en/stable/operator-manual/declarative-setup/#repositories) within ArgoCD.",
		Attributes:          repositorySchemaAttributes(),
	}
}

func (r *repositoryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *repositoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data repositoryModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	// Initialize API clients
	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	// Check for errors before proceeding
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to API model
	repo, err := data.toAPIModel()
	if err != nil {
		resp.Diagnostics.AddError("Failed to convert repository model", err.Error())
		return
	}

	timeout := 2 * time.Minute

	// Create repository with retry logic for SSH handshake issues
	var createdRepo *v1alpha1.Repository

	retryErr := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		sync.RepositoryMutex.Lock()
		defer sync.RepositoryMutex.Unlock()

		var createErr error
		createdRepo, createErr = r.si.RepositoryClient.CreateRepository(
			ctx,
			&repository.RepoCreateRequest{
				Repo:   repo,
				Upsert: false,
			},
		)

		if createErr != nil {
			// Check for SSH handshake issues and retry
			if matched, _ := regexp.MatchString("ssh: handshake failed: knownhosts: key is unknown", createErr.Error()); matched {
				tflog.Warn(ctx, fmt.Sprintf("SSH handshake failed for repository %s, retrying in case a repository certificate has been set recently", repo.Repo))
				return retry.RetryableError(createErr)
			}

			return retry.NonRetryableError(createErr)
		}

		if createdRepo == nil {
			return retry.NonRetryableError(fmt.Errorf("ArgoCD did not return an error or a repository result"))
		}

		if createdRepo.ConnectionState.Status == v1alpha1.ConnectionStatusFailed {
			return retry.NonRetryableError(fmt.Errorf("could not connect to repository %s: %s", repo.Repo, createdRepo.ConnectionState.Message))
		}

		return nil
	})

	if retryErr != nil {
		resp.Diagnostics.Append(diagnostics.ArgoCDAPIError("create", "repository", repo.Repo, retryErr)...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created repository %s", createdRepo.Repo))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, data.updateFromAPI(repo))...)

	// Perform a read to get the latest state with connection status
	if !resp.Diagnostics.HasError() {
		readResp := &resource.ReadResponse{State: resp.State, Diagnostics: resp.Diagnostics}
		r.Read(ctx, resource.ReadRequest{State: resp.State}, readResp)
		resp.Diagnostics = readResp.Diagnostics
		resp.State = readResp.State
	}
}

func (r *repositoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data repositoryModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	// Initialize API clients
	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	// Check for errors before proceeding
	if resp.Diagnostics.HasError() {
		return
	}

	// Read repository from API
	repo, diags := r.readRepository(ctx, data.ID.ValueString(), data.Project.ValueString())
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If repository was not found, remove from state
	if repo == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, data.updateFromAPI(repo))...)
}

func (r *repositoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data repositoryModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	// Initialize API clients
	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	// Check for errors before proceeding
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to API model
	repo, err := data.toAPIModel()
	if err != nil {
		resp.Diagnostics.AddError("Failed to convert repository model", err.Error())
		return
	}

	// Update repository
	sync.RepositoryMutex.Lock()
	defer sync.RepositoryMutex.Unlock()

	updatedRepo, err := r.si.RepositoryClient.UpdateRepository(
		ctx,
		&repository.RepoUpdateRequest{Repo: repo},
	)

	if err != nil {
		resp.Diagnostics.Append(diagnostics.ArgoCDAPIError("update", "repository", repo.Repo, err)...)
		return
	}

	if updatedRepo == nil {
		resp.Diagnostics.AddError("ArgoCD did not return an error or a repository result", "")
		return
	}

	if updatedRepo.ConnectionState.Status == v1alpha1.ConnectionStatusFailed {
		resp.Diagnostics.AddError(
			"Repository connection failed",
			fmt.Sprintf("could not connect to repository %s: %s", repo.Repo, updatedRepo.ConnectionState.Message),
		)

		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated repository %s", updatedRepo.Repo))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)

	// Perform a read to get the latest state
	if !resp.Diagnostics.HasError() {
		readResp := &resource.ReadResponse{State: resp.State, Diagnostics: resp.Diagnostics}
		r.Read(ctx, resource.ReadRequest{State: resp.State}, readResp)
		resp.Diagnostics = readResp.Diagnostics
		resp.State = readResp.State
	}
}

func (r *repositoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data repositoryModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	// Initialize API clients
	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	// Check for errors before proceeding
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete repository
	sync.RepositoryMutex.Lock()
	defer sync.RepositoryMutex.Unlock()

	_, err := r.si.RepositoryClient.DeleteRepository(
		ctx,
		&repository.RepoQuery{
			Repo:       data.ID.ValueString(),
			AppProject: data.Project.ValueString(),
		},
	)

	if err != nil {
		if !strings.Contains(err.Error(), "NotFound") {
			resp.Diagnostics.Append(diagnostics.ArgoCDAPIError("delete", "repository", data.ID.ValueString(), err)...)
			return
		}
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted repository %s", data.ID.ValueString()))
}

func (r *repositoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *repositoryResource) readRepository(ctx context.Context, repoURL, project string) (*v1alpha1.Repository, diag.Diagnostics) {
	var diags diag.Diagnostics

	sync.RepositoryMutex.RLock()
	defer sync.RepositoryMutex.RUnlock()

	repos, err := r.si.RepositoryClient.List(ctx, &repository.RepoQuery{
		AppProject: project,
	})

	var finalRepo *v1alpha1.Repository

	if repos != nil {
		for _, repo := range repos.Items {
			if repo.Repo == repoURL {
				finalRepo = repo
			}
		}
	}

	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			// Repository has been deleted out-of-band
			return nil, diags
		}

		diags.Append(diagnostics.ArgoCDAPIError("read", "repository", repoURL, err)...)

		return nil, diags
	}

	return finalRepo, diags
}

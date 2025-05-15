package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	application "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"

	"github.com/argoproj-labs/terraform-provider-argocd/internal/diagnostics"
	"github.com/argoproj-labs/terraform-provider-argocd/internal/sync"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/repository"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &repositoryResource{}

type repositoryResource struct {
	si *ServerInterface
}

func NewRepositoryResource() resource.Resource {
	return repositoryResource{}
}

func (r repositoryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository"
}

// Schema implements resource.Resource.
func (r repositoryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "",
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

func (r *repositoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r repositoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var repo repositoryModel

	// read Terraform configuration into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &repo)...)

	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	// check for errors before proceeding
	if resp.Diagnostics.HasError() {
		return
	}

	appID, _ := strconv.ParseInt(repo.GithubAppID.String(), 10, 64)
	appInstallID, _ := strconv.ParseInt(repo.GithubAppInstallationID.String(), 10, 64)

	// create repository
	sync.RepositoryMutex.Lock()
	returnedRepo, err := r.si.RepositoryClient.CreateRepository(
		ctx,
		&repository.RepoCreateRequest{
			Repo: &application.Repository{
				Repo:                       repo.Repo.String(),
				EnableLFS:                  repo.EnableLFS.ValueBool(),
				InheritedCreds:             repo.InheritedCreds.ValueBool(),
				Insecure:                   repo.Insecure.ValueBool(),
				Name:                       repo.Name.String(),
				Project:                    repo.Project.String(),
				Username:                   repo.Username.String(),
				Password:                   repo.Password.String(),
				SSHPrivateKey:              repo.SSHPrivateKey.String(),
				TLSClientCertData:          repo.TLSClientCertData.String(),
				TLSClientCertKey:           repo.TLSClientCertKey.String(),
				EnableOCI:                  repo.EnableOCI.ValueBool(),
				Type:                       repo.Type.String(),
				GithubAppId:                appID,
				GithubAppInstallationId:    appInstallID,
				GitHubAppEnterpriseBaseURL: repo.GitHubAppEnterpriseBaseURL.String(),
				GithubAppPrivateKey:        repo.GithubAppPrivateKey.String(),
			},
			Upsert: false,
		},
	)
	sync.RepositoryMutex.Unlock()

	if err != nil {
		// TODO: better way to detect ssh handshake failing ?
		if matched, _ := regexp.MatchString("ssh: handshake failed: knownhosts: key is unknown", err.Error()); matched {
			resp.Diagnostics.Append(diagnostics.Error("handskace failed for repository", fmt.Errorf("handshake failed for repository %s, retrying in case a repository certificate has been set recently", repo.Repo.String()))...)
			return
		}

		resp.Diagnostics.Append(diagnostics.ArgoCDAPIError("create", "repository", repo.Name.String(), err)...)

		return
	}

	if returnedRepo == nil {
		resp.Diagnostics.Append(diagnostics.Error("Empty return from Argo CD", fmt.Errorf("ArgoCD did not return an error or a repository result: %s", err))...)

		return
	}

	if returnedRepo.ConnectionState.Status == application.ConnectionStatusFailed {
		resp.Diagnostics.Append(diagnostics.Error("could not connect to repository", fmt.Errorf("could not connect to repository %s: %s", repo.Repo.String(), returnedRepo.ConnectionState.Message))...)

		return
	}

	diags := resp.State.Set(ctx, newRepository(returnedRepo))
	resp.Diagnostics.Append(diags...)
}

func (r repositoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var repo repositoryModel

	// read Terraform configuration into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &repo)...)

	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	// check for errors before proceeding
	if resp.Diagnostics.HasError() {
		return
	}

	sync.RepositoryMutex.RLock()
	returnedRepo, err := r.si.RepositoryClient.Get(ctx, &repository.RepoQuery{
		Repo:         repo.Repo.String(),
		AppProject:   repo.Project.String(),
		ForceRefresh: true,
	})

	sync.RepositoryMutex.RUnlock()

	if err != nil {
		// Repository has already been deleted in an out-of-band fashion
		if strings.Contains(err.Error(), "NotFound") {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.Append(diagnostics.ArgoCDAPIError("read", "repository", repo.Repo.String(), err)...)

		return
	}

	diags := resp.State.Set(ctx, newRepository(returnedRepo))
	resp.Diagnostics.Append(diags...)
}

func (r repositoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var repo repositoryModel

	// read Terraform configuration into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &repo)...)

	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	// check for errors before proceeding
	if resp.Diagnostics.HasError() {
		return
	}

	appID, _ := strconv.ParseInt(repo.GithubAppID.String(), 10, 64)
	appInstallID, _ := strconv.ParseInt(repo.GithubAppInstallationID.String(), 10, 64)

	sync.RepositoryMutex.Lock()
	returnedRepo, err := r.si.RepositoryClient.UpdateRepository(
		ctx,
		&repository.RepoUpdateRequest{Repo: &application.Repository{
			Repo:                       repo.Repo.String(),
			EnableLFS:                  repo.EnableLFS.ValueBool(),
			InheritedCreds:             repo.InheritedCreds.ValueBool(),
			Insecure:                   repo.Insecure.ValueBool(),
			Name:                       repo.Name.String(),
			Project:                    repo.Project.String(),
			Username:                   repo.Username.String(),
			Password:                   repo.Password.String(),
			SSHPrivateKey:              repo.SSHPrivateKey.String(),
			TLSClientCertData:          repo.TLSClientCertData.String(),
			TLSClientCertKey:           repo.TLSClientCertKey.String(),
			EnableOCI:                  repo.EnableOCI.ValueBool(),
			Type:                       repo.Type.String(),
			GithubAppId:                appID,
			GithubAppInstallationId:    appInstallID,
			GitHubAppEnterpriseBaseURL: repo.GitHubAppEnterpriseBaseURL.String(),
			GithubAppPrivateKey:        repo.GithubAppPrivateKey.String(),
		}},
	)
	sync.RepositoryMutex.Unlock()

	if err != nil {
		resp.Diagnostics.Append(diagnostics.ArgoCDAPIError("update", "repository", repo.Repo.String(), err)...)
		return
	}

	if returnedRepo == nil {
		resp.Diagnostics.Append(diagnostics.Error("Failed reading repository", fmt.Errorf("ArgoCD did not return an error or a repository result for ID %s: %q", repo.Repo.String(), err))...)
		return
	}

	if returnedRepo.ConnectionState.Status == application.ConnectionStatusFailed {
		resp.Diagnostics.Append(diagnostics.Error("could not connect to repository", fmt.Errorf("could not connect to repository %s: %s", repo.Repo.String(), returnedRepo.ConnectionState.Message))...)
		return
	}

	diags := resp.State.Set(ctx, newRepository(returnedRepo))
	resp.Diagnostics.Append(diags...)
}

func (r repositoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var repo repositoryModel

	// read Terraform configuration into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &repo)...)

	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	// check for errors before proceeding
	if resp.Diagnostics.HasError() {
		return
	}

	sync.RepositoryMutex.Lock()
	_, err := r.si.RepositoryClient.DeleteRepository(
		ctx,
		&repository.RepoQuery{Repo: repo.Repo.String(), AppProject: repo.Project.String()},
	)
	sync.RepositoryMutex.Unlock()

	if err != nil {
		// Repository has already been deleted in an out-of-band fashion
		if strings.Contains(err.Error(), "NotFound") {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.Append(diagnostics.ArgoCDAPIError("delete", "repository", repo.Repo.String(), err)...)

		return
	}
}

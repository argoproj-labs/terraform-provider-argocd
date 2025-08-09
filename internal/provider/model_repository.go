package provider

import (
	"strconv"

	"github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type repositoryModel struct {
	ID                         types.String `tfsdk:"id"`
	Repo                       types.String `tfsdk:"repo"`
	Name                       types.String `tfsdk:"name"`
	Type                       types.String `tfsdk:"type"`
	Project                    types.String `tfsdk:"project"`
	Username                   types.String `tfsdk:"username"`
	Password                   types.String `tfsdk:"password"`
	SSHPrivateKey              types.String `tfsdk:"ssh_private_key"`
	TLSClientCertData          types.String `tfsdk:"tls_client_cert_data"`
	TLSClientCertKey           types.String `tfsdk:"tls_client_cert_key"`
	EnableLFS                  types.Bool   `tfsdk:"enable_lfs"`
	EnableOCI                  types.Bool   `tfsdk:"enable_oci"`
	Insecure                   types.Bool   `tfsdk:"insecure"`
	InheritedCreds             types.Bool   `tfsdk:"inherited_creds"`
	ConnectionStateStatus      types.String `tfsdk:"connection_state_status"`
	GitHubAppID                types.String `tfsdk:"githubapp_id"`
	GitHubAppInstallationID    types.String `tfsdk:"githubapp_installation_id"`
	GitHubAppEnterpriseBaseURL types.String `tfsdk:"githubapp_enterprise_base_url"`
	GitHubAppPrivateKey        types.String `tfsdk:"githubapp_private_key"`
}

func repositorySchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "Repository identifier",
			Computed:            true,
		},
		"repo": schema.StringAttribute{
			MarkdownDescription: "URL of the repository.",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name to be used for this repo. Only used with Helm repos.",
			Optional:            true,
		},
		"type": schema.StringAttribute{
			MarkdownDescription: "Type of the repo. Can be either `git` or `helm`. `git` is assumed if empty or absent.",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("git"),
			Validators: []validator.String{
				stringvalidator.OneOf("git", "helm"),
			},
		},
		"project": schema.StringAttribute{
			MarkdownDescription: "The project name, in case the repository is project scoped.",
			Optional:            true,
		},
		"username": schema.StringAttribute{
			MarkdownDescription: "Username used for authenticating at the remote repository.",
			Optional:            true,
		},
		"password": schema.StringAttribute{
			MarkdownDescription: "Password or PAT used for authenticating at the remote repository.",
			Optional:            true,
			Sensitive:           true,
		},
		"ssh_private_key": schema.StringAttribute{
			MarkdownDescription: "PEM data for authenticating at the repo server. Only used with Git repos.",
			Optional:            true,
			Sensitive:           true,
		},
		"tls_client_cert_data": schema.StringAttribute{
			MarkdownDescription: "TLS client certificate in PEM format for authenticating at the repo server.",
			Optional:            true,
		},
		"tls_client_cert_key": schema.StringAttribute{
			MarkdownDescription: "TLS client certificate private key in PEM format for authenticating at the repo server.",
			Optional:            true,
			Sensitive:           true,
		},
		"enable_lfs": schema.BoolAttribute{
			MarkdownDescription: "Whether `git-lfs` support should be enabled for this repository.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"enable_oci": schema.BoolAttribute{
			MarkdownDescription: "Whether `helm-oci` support should be enabled for this repository.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"insecure": schema.BoolAttribute{
			MarkdownDescription: "Whether the connection to the repository ignores any errors when verifying TLS certificates or SSH host keys.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"inherited_creds": schema.BoolAttribute{
			MarkdownDescription: "Whether credentials were inherited from a credential set.",
			Computed:            true,
		},
		"connection_state_status": schema.StringAttribute{
			MarkdownDescription: "Contains information about the current state of connection to the repository server.",
			Computed:            true,
		},
		"githubapp_id": schema.StringAttribute{
			MarkdownDescription: "ID of the GitHub app used to access the repo.",
			Optional:            true,
		},
		"githubapp_installation_id": schema.StringAttribute{
			MarkdownDescription: "The installation ID of the GitHub App used to access the repo.",
			Optional:            true,
		},
		"githubapp_enterprise_base_url": schema.StringAttribute{
			MarkdownDescription: "GitHub API URL for GitHub app authentication.",
			Optional:            true,
		},
		"githubapp_private_key": schema.StringAttribute{
			MarkdownDescription: "Private key data (PEM) for authentication via GitHub app.",
			Optional:            true,
			Sensitive:           true,
		},
	}
}

func (m *repositoryModel) toAPIModel() (*v1alpha1.Repository, error) {
	repo := &v1alpha1.Repository{
		Repo:                       m.Repo.ValueString(),
		Name:                       m.Name.ValueString(),
		Type:                       m.Type.ValueString(),
		Project:                    m.Project.ValueString(),
		Username:                   m.Username.ValueString(),
		Password:                   m.Password.ValueString(),
		SSHPrivateKey:              m.SSHPrivateKey.ValueString(),
		TLSClientCertData:          m.TLSClientCertData.ValueString(),
		TLSClientCertKey:           m.TLSClientCertKey.ValueString(),
		EnableLFS:                  m.EnableLFS.ValueBool(),
		EnableOCI:                  m.EnableOCI.ValueBool(),
		Insecure:                   m.Insecure.ValueBool(),
		InheritedCreds:             m.InheritedCreds.ValueBool(),
		GitHubAppEnterpriseBaseURL: m.GitHubAppEnterpriseBaseURL.ValueString(),
		GithubAppPrivateKey:        m.GitHubAppPrivateKey.ValueString(),
	}

	// Handle GitHub App ID conversion
	if !m.GitHubAppID.IsNull() && !m.GitHubAppID.IsUnknown() {
		id, err := strconv.ParseInt(m.GitHubAppID.ValueString(), 10, 64)
		if err != nil {
			return nil, err
		}

		repo.GithubAppId = id
	}

	// Handle GitHub App Installation ID conversion
	if !m.GitHubAppInstallationID.IsNull() && !m.GitHubAppInstallationID.IsUnknown() {
		id, err := strconv.ParseInt(m.GitHubAppInstallationID.ValueString(), 10, 64)
		if err != nil {
			return nil, err
		}

		repo.GithubAppInstallationId = id
	}

	return repo, nil
}

func (m *repositoryModel) updateFromAPI(repo *v1alpha1.Repository) *repositoryModel {
	m.ID = types.StringValue(repo.Repo)
	m.Repo = types.StringValue(repo.Repo)
	m.Type = types.StringValue(repo.Type)
	m.EnableLFS = types.BoolValue(repo.EnableLFS)
	m.EnableOCI = types.BoolValue(repo.EnableOCI)
	m.Insecure = types.BoolValue(repo.Insecure)
	m.InheritedCreds = types.BoolValue(repo.InheritedCreds)

	if repo.Name != "" {
		m.Name = types.StringValue(repo.Name)
	}

	// Handle connection state status
	if repo.ConnectionState.Status != "" {
		m.ConnectionStateStatus = types.StringValue(repo.ConnectionState.Status)
	}

	if repo.Project != "" {
		m.Project = types.StringValue(repo.Project)
	}

	// Handle username based on inheritance
	if !repo.InheritedCreds {
		if repo.Username != "" {
			m.Username = types.StringValue(repo.Username)
		}
	}

	if repo.GitHubAppEnterpriseBaseURL != "" {
		m.GitHubAppEnterpriseBaseURL = types.StringValue(repo.GitHubAppEnterpriseBaseURL)
	}

	// Handle GitHub App ID conversion
	if repo.GithubAppId > 0 {
		m.GitHubAppID = types.StringValue(strconv.FormatInt(repo.GithubAppId, 10))
	}

	// Handle GitHub App Installation ID conversion
	if repo.GithubAppInstallationId > 0 {
		m.GitHubAppInstallationID = types.StringValue(strconv.FormatInt(repo.GithubAppInstallationId, 10))
	}

	return m
}

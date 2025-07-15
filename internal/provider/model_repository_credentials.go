package provider

import (
	"strconv"

	"github.com/argoproj-labs/terraform-provider-argocd/internal/validators"
	"github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type repositoryCredentialsModel struct {
	ID                         types.String `tfsdk:"id"`
	URL                        types.String `tfsdk:"url"`
	Username                   types.String `tfsdk:"username"`
	Password                   types.String `tfsdk:"password"`
	SSHPrivateKey              types.String `tfsdk:"ssh_private_key"`
	TLSClientCertData          types.String `tfsdk:"tls_client_cert_data"`
	TLSClientCertKey           types.String `tfsdk:"tls_client_cert_key"`
	EnableOCI                  types.Bool   `tfsdk:"enable_oci"`
	GitHubAppID                types.String `tfsdk:"githubapp_id"`
	GitHubAppInstallationID    types.String `tfsdk:"githubapp_installation_id"`
	GitHubAppEnterpriseBaseURL types.String `tfsdk:"githubapp_enterprise_base_url"`
	GitHubAppPrivateKey        types.String `tfsdk:"githubapp_private_key"`
}

func repositoryCredentialsSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "Repository credentials identifier",
			Computed:            true,
		},
		"url": schema.StringAttribute{
			MarkdownDescription: "URL that these credentials match to",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"username": schema.StringAttribute{
			MarkdownDescription: "Username for authenticating at the repo server",
			Optional:            true,
		},
		"password": schema.StringAttribute{
			MarkdownDescription: "Password for authenticating at the repo server",
			Optional:            true,
			Sensitive:           true,
		},
		"ssh_private_key": schema.StringAttribute{
			MarkdownDescription: "Private key data for authenticating at the repo server using SSH (only Git repos)",
			Optional:            true,
			Sensitive:           true,
			Validators: []validator.String{
				validators.SSHPrivateKey(),
			},
		},
		"tls_client_cert_data": schema.StringAttribute{
			MarkdownDescription: "TLS client cert data for authenticating at the repo server",
			Optional:            true,
		},
		"tls_client_cert_key": schema.StringAttribute{
			MarkdownDescription: "TLS client cert key for authenticating at the repo server",
			Optional:            true,
			Sensitive:           true,
		},
		"enable_oci": schema.BoolAttribute{
			MarkdownDescription: "Whether `helm-oci` support should be enabled for this repo",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"githubapp_id": schema.StringAttribute{
			MarkdownDescription: "GitHub App ID of the app used to access the repo for GitHub app authentication",
			Optional:            true,
			Validators: []validator.String{
				validators.PositiveInteger(),
			},
		},
		"githubapp_installation_id": schema.StringAttribute{
			MarkdownDescription: "ID of the installed GitHub App for GitHub app authentication",
			Optional:            true,
			Validators: []validator.String{
				validators.PositiveInteger(),
			},
		},
		"githubapp_enterprise_base_url": schema.StringAttribute{
			MarkdownDescription: "GitHub API URL for GitHub app authentication",
			Optional:            true,
		},
		"githubapp_private_key": schema.StringAttribute{
			MarkdownDescription: "Private key data (PEM) for authentication via GitHub app",
			Optional:            true,
			Sensitive:           true,
			Validators: []validator.String{
				validators.SSHPrivateKey(),
			},
		},
	}
}

func (m *repositoryCredentialsModel) toAPIModel() (*v1alpha1.RepoCreds, error) {
	creds := &v1alpha1.RepoCreds{
		URL:                        m.URL.ValueString(),
		Username:                   m.Username.ValueString(),
		Password:                   m.Password.ValueString(),
		SSHPrivateKey:              m.SSHPrivateKey.ValueString(),
		TLSClientCertData:          m.TLSClientCertData.ValueString(),
		TLSClientCertKey:           m.TLSClientCertKey.ValueString(),
		EnableOCI:                  m.EnableOCI.ValueBool(),
		GitHubAppEnterpriseBaseURL: m.GitHubAppEnterpriseBaseURL.ValueString(),
		GithubAppPrivateKey:        m.GitHubAppPrivateKey.ValueString(),
	}

	// Handle GitHub App ID conversion
	if !m.GitHubAppID.IsNull() && !m.GitHubAppID.IsUnknown() {
		id, err := strconv.ParseInt(m.GitHubAppID.ValueString(), 10, 64)
		if err != nil {
			return nil, err
		}

		creds.GithubAppId = id
	}

	// Handle GitHub App Installation ID conversion
	if !m.GitHubAppInstallationID.IsNull() && !m.GitHubAppInstallationID.IsUnknown() {
		id, err := strconv.ParseInt(m.GitHubAppInstallationID.ValueString(), 10, 64)
		if err != nil {
			return nil, err
		}

		creds.GithubAppInstallationId = id
	}

	return creds, nil
}

func newRepositoryCredentialsModel(creds *v1alpha1.RepoCreds) *repositoryCredentialsModel {
	model := &repositoryCredentialsModel{
		ID:        types.StringValue(creds.URL),
		URL:       types.StringValue(creds.URL),
		EnableOCI: types.BoolValue(creds.EnableOCI),
	}

	// Handle username - only set if not empty
	if creds.Username != "" {
		model.Username = types.StringValue(creds.Username)
	} else {
		model.Username = types.StringNull()
	}

	// Set string fields to null if empty, otherwise use the value
	if creds.GitHubAppEnterpriseBaseURL != "" {
		model.GitHubAppEnterpriseBaseURL = types.StringValue(creds.GitHubAppEnterpriseBaseURL)
	} else {
		model.GitHubAppEnterpriseBaseURL = types.StringNull()
	}

	if creds.TLSClientCertData != "" {
		model.TLSClientCertData = types.StringValue(creds.TLSClientCertData)
	} else {
		model.TLSClientCertData = types.StringNull()
	}

	// Handle GitHub App ID conversion
	if creds.GithubAppId > 0 {
		model.GitHubAppID = types.StringValue(strconv.FormatInt(creds.GithubAppId, 10))
	} else {
		model.GitHubAppID = types.StringNull()
	}

	// Handle GitHub App Installation ID conversion
	if creds.GithubAppInstallationId > 0 {
		model.GitHubAppInstallationID = types.StringValue(strconv.FormatInt(creds.GithubAppInstallationId, 10))
	} else {
		model.GitHubAppInstallationID = types.StringNull()
	}

	// Note: Sensitive fields (password, ssh_private_key, tls_client_cert_key, githubapp_private_key)
	// are not returned by the ArgoCD API, so they remain as configured in Terraform state
	model.Password = types.StringNull()
	model.SSHPrivateKey = types.StringNull()
	model.TLSClientCertKey = types.StringNull()
	model.GitHubAppPrivateKey = types.StringNull()

	return model
}

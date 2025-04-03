package provider

import (
	"regexp"
	"strconv"

	"github.com/argoproj-labs/terraform-provider-argocd/internal/validators"
	application "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type repositoryModel struct {
	Repo                       types.String `tfsdk:"repo"`
	EnableLFS                  types.Bool   `tfsdk:"enable_lfs"`
	InheritedCreds             types.Bool   `tfsdk:"inherited_creds"`
	Insecure                   types.Bool   `tfsdk:"insecure"`
	Name                       types.String `tfsdk:"name"`
	Project                    types.String `tfsdk:"project"`
	Username                   types.String `tfsdk:"username"`
	Password                   types.String `tfsdk:"password"`
	SSHPrivateKey              types.String `tfsdk:"ssh_private_key"`
	TLSClientCertData          types.String `tfsdk:"tls_client_cert_data"`
	TLSClientCertKey           types.String `tfsdk:"tls_client_cert_key"`
	EnableOCI                  types.Bool   `tfsdk:"enable_oci"`
	Type                       types.String `tfsdk:"type"`
	ConnectionStateStatus      types.String `tfsdk:"connection_state_status"`
	GithubAppID                types.String `tfsdk:"githubapp_id"`
	GithubAppInstallationID    types.String `tfsdk:"githubapp_installation_id"`
	GitHubAppEnterpriseBaseURL types.String `tfsdk:"githubapp_enterprise_base_url"`
	GithubAppPrivateKey        types.String `tfsdk:"githubapp_private_key"`
}

func repositorySchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"repo": schema.StringAttribute{
			MarkdownDescription: "URL of the repository.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Required: true,
		},
		"enable_lfs": schema.BoolAttribute{
			MarkdownDescription: "Whether `git-lfs` support should be enabled for this repository.",
			Optional:            true,
		},
		"inherited_creds": schema.BoolAttribute{
			MarkdownDescription: "Whether credentials were inherited from a credential set.",
			Computed:            true,
		},
		"insecure": schema.BoolAttribute{
			MarkdownDescription: "Whether the connection to the repository ignores any errors when verifying TLS certificates or SSH host keys.",
			Optional:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name to be used for this repo. Only used with Helm repos.",
			Optional:            true,
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
			Sensitive:           true,
			Optional:            true,
		},
		"ssh_private_key": schema.StringAttribute{
			MarkdownDescription: "PEM data for authenticating at the repo server. Only used with Git repos.",
			Sensitive:           true,
			Validators: []validator.String{
				validators.IsSSHPrivateKey(),
			},
			Optional: true,
		},
		"tls_client_cert_data": schema.StringAttribute{
			MarkdownDescription: "TLS client certificate in PEM format for authenticating at the repo server.",
			// TODO: add a validator
			Optional: true,
		},
		"tls_client_cert_key": schema.StringAttribute{
			MarkdownDescription: "TLS client certificate private key in PEM format for authenticating at the repo server.",
			Sensitive:           true,
			// TODO: add a validator
			Optional: true,
		},
		"enable_oci": schema.BoolAttribute{
			MarkdownDescription: "Whether `helm-oci` support should be enabled for this repository.",
			Optional:            true,
		},
		"type": schema.StringAttribute{
			MarkdownDescription: "Type of the repo. Can be either `git` or `helm`. `git` is assumed if empty or absent.",
			Default:             stringdefault.StaticString("git"),
			Validators: []validator.String{
				stringvalidator.OneOf("helm", "git"),
			},
			Optional: true,
			Computed: true,
		},
		"connection_state_status": schema.StringAttribute{
			MarkdownDescription: "Contains information about the current state of connection to the repository server.",
			Computed:            true,
		},
		"githubapp_id": schema.StringAttribute{
			MarkdownDescription: "ID of the GitHub app used to access the repo.",
			Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile(`^[+]?\d+?$`), "String input must match a positive integer, e.g.'12345'"),
			},
			Optional: true,
		},
		"githubapp_installation_id": schema.StringAttribute{
			MarkdownDescription: "The installation ID of the GitHub App used to access the repo.",
			Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile(`^[+]?\d+?$`), "String input must match a positive integer, e.g.'12345'"),
			},
			Optional: true,
		},
		"githubapp_enterprise_base_url": schema.StringAttribute{
			MarkdownDescription: "GitHub API URL for GitHub app authentication.",
			Optional:            true,
		},
		"githubapp_private_key": schema.StringAttribute{
			MarkdownDescription: "Private key data (PEM) for authentication via GitHub app.",
			Validators: []validator.String{
				validators.IsSSHPrivateKey(),
			},
			Sensitive: true,
			Optional:  true,
		},
	}
}

func newRepository(r *application.Repository) *repositoryModel {
	return &repositoryModel{
		Repo:                       types.StringValue(r.Repo),
		EnableLFS:                  types.BoolValue(r.EnableLFS),
		InheritedCreds:             types.BoolValue(r.InheritedCreds),
		Insecure:                   types.BoolValue(r.Insecure),
		Name:                       types.StringValue(r.Name),
		Project:                    types.StringValue(r.Project),
		Username:                   types.StringValue(r.Username),
		Password:                   types.StringValue(r.Password),
		SSHPrivateKey:              types.StringValue(r.SSHPrivateKey),
		TLSClientCertData:          types.StringValue(r.TLSClientCertData),
		TLSClientCertKey:           types.StringValue(r.TLSClientCertKey),
		EnableOCI:                  types.BoolValue(r.EnableOCI),
		Type:                       types.StringValue(r.Type),
		ConnectionStateStatus:      types.StringValue(r.ConnectionState.Status),
		GithubAppID:                types.StringValue(strconv.FormatInt(r.GithubAppId, 10)),
		GithubAppInstallationID:    types.StringValue(strconv.FormatInt(r.GithubAppInstallationId, 10)),
		GitHubAppEnterpriseBaseURL: types.StringValue(r.GitHubAppEnterpriseBaseURL),
		GithubAppPrivateKey:        types.StringValue(r.GithubAppPrivateKey),
	}
}

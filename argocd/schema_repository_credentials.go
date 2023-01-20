package argocd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func repositoryCredentialsSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"url": {
			Type:        schema.TypeString,
			Description: "URL that these credentials matches to.",
			Required:    true,
		},
		"username": {
			Type:        schema.TypeString,
			Description: "Username for authenticating at the repo server.",
			Optional:    true,
		},
		"password": {
			Type:        schema.TypeString,
			Sensitive:   true,
			Description: "Password for authenticating at the repo server.",
			Optional:    true,
		},
		"ssh_private_key": {
			Type:         schema.TypeString,
			Sensitive:    true,
			Description:  "Private key data for authenticating at the repo server using SSH (only Git repos).",
			ValidateFunc: validateSSHPrivateKey,
			Optional:     true,
		},
		"tls_client_cert_data": {
			Type:        schema.TypeString,
			Description: "TLS client cert data for authenticating at the repo server.",
			// TODO: add a validator
			Optional: true,
		},
		"tls_client_cert_key": {
			Type:        schema.TypeString,
			Sensitive:   true,
			Description: "TLS client cert key for authenticating at the repo server.",
			// TODO: add a validator
			Optional: true,
		},
		"enable_oci": {
			Type:        schema.TypeBool,
			Description: "Whether `helm-oci` support should be enabled for this repo.",
			Optional:    true,
		},
		"githubapp_id": {
			Type:         schema.TypeString,
			Description:  "Github App ID of the app used to access the repo for GitHub app authentication.",
			ValidateFunc: validatePositiveInteger,
			Optional:     true,
		},
		"githubapp_installation_id": {
			Type:         schema.TypeString,
			Description:  "ID of the installed GitHub App for GitHub app authentication.",
			ValidateFunc: validatePositiveInteger,
			Optional:     true,
		},
		"githubapp_enterprise_base_url": {
			Type:        schema.TypeString,
			Description: "GitHub API URL for GitHub app authentication.",
			Optional:    true,
		},
		"githubapp_private_key": {
			Type:         schema.TypeString,
			Sensitive:    true,
			Description:  "Private key data (PEM) for authentication via GitHub app.",
			ValidateFunc: validateSSHPrivateKey,
			Optional:     true,
		},
	}
}

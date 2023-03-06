package argocd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func repositoryCredentialsSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"url": {
			Type:        schema.TypeString,
			Description: "URL is the URL that these credentials matches to",
			Required:    true,
		},
		"username": {
			Type:        schema.TypeString,
			Description: "Username for authenticating at the repo server",
			Optional:    true,
		},
		"password": {
			Type:        schema.TypeString,
			Sensitive:   true,
			Description: "Password for authenticating at the repo server",
			Optional:    true,
		},
		"ssh_private_key": {
			Type:         schema.TypeString,
			Sensitive:    true,
			Description:  "SSH private key data for authenticating at the repo server only for Git repos",
			ValidateFunc: validateSSHPrivateKey,
			Optional:     true,
		},
		"tls_client_cert_data": {
			Type:        schema.TypeString,
			Description: "TLS client cert data for authenticating at the repo server",
			// TODO: add a validator
			Optional: true,
		},
		"tls_client_cert_key": {
			Type:        schema.TypeString,
			Sensitive:   true,
			Description: "TLS client cert key for authenticating at the repo server ",
			// TODO: add a validator
			Optional: true,
		},
		"enable_oci": {
			Type:        schema.TypeBool,
			Description: "Specify whether the repo server should be viewed as OCI compliant",
			Optional:    true,
		},
		"githubapp_id": {
			Type:        schema.TypeString,
			Description: "GitHub App id for authenticating at the repo server only for GitHub repos",
			Optional:    true,
		},
		"githubapp_installation_id": {
			Type:        schema.TypeString,
			Description: "GitHub App installation id for authenticating at the repo server only for GitHub repos",
			Optional:    true,
		},
		"githubapp_enterprise_base_url": {
			Type:        schema.TypeString,
			Description: "If using GitHub App for a GitHub Enterprise repository the host url is required",
			Optional:    true,
		},
		"githubapp_private_key": {
			Type:         schema.TypeString,
			Sensitive:    true,
			Description:  "Private key data (pem) of GitHub App for authenticating at the repo server only for GitHub repos",
			ValidateFunc: validateSSHPrivateKey,
			Optional:     true,
		},
	}
}

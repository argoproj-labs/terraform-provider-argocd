package argocd

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func repositorySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"repo": {
			Type:        schema.TypeString,
			Description: "URL of the repository.",
			ForceNew:    true,
			Required:    true,
		},
		"enable_lfs": {
			Type:        schema.TypeBool,
			Description: "Whether `git-lfs` support should be enabled for this repository.",
			Optional:    true,
		},
		"inherited_creds": {
			Type:        schema.TypeBool,
			Description: "Whether credentials were inherited from a credential set.",
			Computed:    true,
		},
		"insecure": {
			Type:        schema.TypeBool,
			Description: "Whether the connection to the repository ignores any errors when verifying TLS certificates or SSH host keys.",
			Optional:    true,
		},
		"name": {
			Type:        schema.TypeString,
			Description: "Name to be used for this repo. Only used with Helm repos.",
			Optional:    true,
		},
		"project": {
			Type:        schema.TypeString,
			Description: "The project name, in case the repository is project scoped.",
			Optional:    true,
		},
		"username": {
			Type:        schema.TypeString,
			Description: "Username used for authenticating at the remote repository.",
			Optional:    true,
		},
		"password": {
			Type:        schema.TypeString,
			Sensitive:   true,
			Description: "Password or PAT used for authenticating at the remote repository.",
			Optional:    true,
		},
		"ssh_private_key": {
			Type:         schema.TypeString,
			Sensitive:    true,
			Description:  "PEM data for authenticating at the repo server. Only used with Git repos.",
			ValidateFunc: validateSSHPrivateKey,
			Optional:     true,
		},
		"tls_client_cert_data": {
			Type:        schema.TypeString,
			Description: "TLS client certificate in PEM format for authenticating at the repo server.",
			// TODO: add a validator
			Optional: true,
		},
		"tls_client_cert_key": {
			Type:        schema.TypeString,
			Sensitive:   true,
			Description: "TLS client certificate private key in PEM format for authenticating at the repo server.",
			// TODO: add a validator
			Optional: true,
		},
		"enable_oci": {
			Type:        schema.TypeBool,
			Description: "Whether `helm-oci` support should be enabled for this repository.",
			Optional:    true,
		},
		"type": {
			Type:        schema.TypeString,
			Description: "Type of the repo. Can be either `git` or `helm`. `git` is assumed if empty or absent.",
			Default:     "git",
			ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
				v := val.(string)
				if v != "git" && v != "helm" {
					errs = append(errs, fmt.Errorf("type can only be 'git' or 'helm', got %s", v))
				}
				return
			},
			Optional: true,
		},
		"connection_state_status": {
			Description: "Contains information about the current state of connection to the repository server.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"githubapp_id": {
			Type:         schema.TypeString,
			Description:  "ID of the GitHub app used to access the repo.",
			ValidateFunc: validatePositiveInteger,
			Optional:     true,
		},
		"githubapp_installation_id": {
			Type:         schema.TypeString,
			Description:  "The installation ID of the GitHub App used to access the repo.",
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

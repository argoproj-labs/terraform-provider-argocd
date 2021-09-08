package argocd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func repositorySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"repo": {
			Type:        schema.TypeString,
			Description: "URL of the repo",
			ForceNew:    true,
			Required:    true,
		},
		"enable_lfs": {
			Type:        schema.TypeBool,
			Description: "Whether git-lfs support should be enabled for this repo",
			Optional:    true,
		},
		"inherited_creds": {
			Type:        schema.TypeBool,
			Description: "Whether credentials were inherited from a credential set",
			Computed:    true,
		},
		"insecure": {
			Type:        schema.TypeBool,
			Description: "Whether the repo is insecure",
			Optional:    true,
		},
		"name": {
			Type:        schema.TypeString,
			Description: "only for Helm repos",
			Optional:    true,
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
		"type": {
			Type:        schema.TypeString,
			Description: "type of the repo, may be 'git' or 'helm', defaults to 'git'",
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
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

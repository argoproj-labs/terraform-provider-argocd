package argocd

import (
	"context"
	"strings"

	"github.com/argoproj-labs/terraform-provider-argocd/internal/provider"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/account"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceArgoCDAccount() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages ArgoCD [accounts](https://argo-cd.readthedocs.io/en/latest/operator-manual/user-management/) for user authentication and authorization.\n\n~> **Note** This resource manages account information and password updates. When the password is changed, the previous state value is used as the current password for the API call. For token management, use the `argocd_account_token` resource.",
		CreateContext: resourceArgoCDAccountCreate,
		ReadContext:   resourceArgoCDAccountRead,
		UpdateContext: resourceArgoCDAccountUpdate,
		DeleteContext: resourceArgoCDAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the ArgoCD account.",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "The password for the account. When changed, will update the account password using the previous state value as the current password. Note: ArgoCD API requires the current password to update/set a password, even for accounts without existing passwords. Initial passwords must be set through ArgoCD configuration.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the account is enabled.",
			},
			"capabilities": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of account capabilities.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"tokens": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of active tokens for the account.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Token ID.",
						},
						"issued_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Unix timestamp when the token was issued.",
						},
						"expires_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Unix timestamp when the token expires, 0 if no expiration.",
						},
					},
				},
			},
		},
	}
}

func resourceArgoCDAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// ArgoCD accounts are typically created through ArgoCD configuration
	// This resource primarily manages existing accounts
	name := d.Get("name").(string)
	d.SetId(name)

	// Verify the account exists by reading it
	return resourceArgoCDAccountRead(ctx, d, meta)
}

func resourceArgoCDAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*provider.ServerInterface)
	if diags := si.InitClients(ctx); diags != nil {
		return pluginSDKDiags(diags)
	}

	accountName := d.Id()
	if accountName == "" {
		accountName = d.Get("name").(string)
	}

	tokenMutexConfiguration.RLock()
	resp, err := si.AccountClient.GetAccount(ctx, &account.GetAccountRequest{
		Name: accountName,
	})
	tokenMutexConfiguration.RUnlock()

	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			d.SetId("")
			return nil
		}

		return argoCDAPIError("read", "account", accountName, err)
	}

	if resp == nil {
		d.SetId("")
		return nil
	}

	if err := d.Set("name", resp.Name); err != nil {
		return errorToDiagnostics("failed to set account name", err)
	}

	if err := d.Set("enabled", resp.Enabled); err != nil {
		return errorToDiagnostics("failed to set account enabled status", err)
	}

	if err := d.Set("capabilities", resp.Capabilities); err != nil {
		return errorToDiagnostics("failed to set account capabilities", err)
	}

	// Parse tokens
	tokens := make([]map[string]interface{}, len(resp.Tokens))
	for i, token := range resp.Tokens {
		tokens[i] = map[string]interface{}{
			"id":         token.Id,
			"issued_at":  convertInt64ToString(token.IssuedAt),
			"expires_at": convertInt64ToString(token.ExpiresAt),
		}
	}

	if err := d.Set("tokens", tokens); err != nil {
		return errorToDiagnostics("failed to set account tokens", err)
	}

	return nil
}

func resourceArgoCDAccountUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*provider.ServerInterface)
	if diags := si.InitClients(ctx); diags != nil {
		return pluginSDKDiags(diags)
	}

	accountName := d.Id()

	// Check if password has changed
	if d.HasChange("password") {
		oldPassword, newPassword := d.GetChange("password")

		// Both old and new passwords must be non-empty strings
		oldPasswordStr := oldPassword.(string)
		newPasswordStr := newPassword.(string)

		if newPasswordStr == "" {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  "password cannot be empty",
				},
			}
		}

		updateReq := &account.UpdatePasswordRequest{
			Name:            accountName,
			CurrentPassword: oldPasswordStr,
			NewPassword:     newPasswordStr,
		}

		tokenMutexConfiguration.Lock()
		_, err := si.AccountClient.UpdatePassword(ctx, updateReq)
		tokenMutexConfiguration.Unlock()

		if err != nil {
			return argoCDAPIError("update", "account password", accountName, err)
		}
	}

	return resourceArgoCDAccountRead(ctx, d, meta)
}

func resourceArgoCDAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// ArgoCD accounts are typically managed through ArgoCD configuration
	// This resource doesn't actually delete accounts from ArgoCD
	// It only removes the account from Terraform state
	d.SetId("")

	return nil
}

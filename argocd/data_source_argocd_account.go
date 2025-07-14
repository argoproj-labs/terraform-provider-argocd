package argocd

import (
	"context"
	"strings"

	"github.com/argoproj-labs/terraform-provider-argocd/internal/provider"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/account"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceArgoCDAccount() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves information about a specific ArgoCD account.",
		ReadContext: dataSourceArgoCDAccountRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the ArgoCD account to retrieve.",
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

func dataSourceArgoCDAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*provider.ServerInterface)
	if diags := si.InitClients(ctx); diags != nil {
		return pluginSDKDiags(diags)
	}

	accountName := d.Get("name").(string)

	tokenMutexConfiguration.RLock()
	resp, err := si.AccountClient.GetAccount(ctx, &account.GetAccountRequest{
		Name: accountName,
	})
	tokenMutexConfiguration.RUnlock()

	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			return diag.Errorf("account %s not found", accountName)
		}

		return argoCDAPIError("read", "account", accountName, err)
	}

	if resp == nil {
		return diag.Errorf("account %s not found", accountName)
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

	d.SetId(accountName)

	return nil
}

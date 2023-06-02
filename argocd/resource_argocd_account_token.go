package argocd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/account"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/session"
	"github.com/cristalhq/jwt/v3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceArgoCDAccountToken() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages ArgoCD [account](https://argo-cd.readthedocs.io/en/latest/user-guide/commands/argocd_account/) JWT tokens.",
		CreateContext: resourceArgoCDAccountTokenCreate,
		ReadContext:   resourceArgoCDAccountTokenRead,
		UpdateContext: resourceArgoCDAccountTokenUpdate,
		DeleteContext: resourceArgoCDAccountTokenDelete,
		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
			ia := d.Get("issued_at").(string)
			if ia == "" {
				// Blank issued_at indicates a new token - nothing to do here
				return nil
			}

			issuedAt, err := convertStringToInt64(ia)
			if err != nil {
				return fmt.Errorf("invalid issued_at: %w", err)
			}

			if ra, ok := d.GetOk("renew_after"); ok {
				renewAfterDuration, err := time.ParseDuration(ra.(string))
				if err != nil {
					return fmt.Errorf("invalid renew_after: %w", err)
				}

				if time.Now().Unix()-issuedAt > int64(renewAfterDuration.Seconds()) {
					// Token is older than renewAfterDuration - force recreation
					if err := d.SetNewComputed("issued_at"); err != nil {
						return fmt.Errorf("failed to force new resource on field %q: %w", "issued_at", err)
					}

					return nil
				}
			}

			ea, ok := d.GetOk("expires_at")
			if !ok {
				return nil
			}

			expiresAt, err := convertStringToInt64(ea.(string))
			if err != nil {
				return fmt.Errorf("invalid expires_at: %w", err)
			}

			if expiresAt == 0 {
				// Token not set to expire - no need to check anything else
				return nil
			}

			if expiresAt < time.Now().Unix() {
				// Token has expired - force recreation
				if err := d.SetNewComputed("expires_at"); err != nil {
					return fmt.Errorf("failed to force new resource on field %q: %w", "expires_at", err)
				}

				return nil
			}

			rb, ok := d.GetOk("renew_before")
			if !ok {
				return nil
			}

			renewBeforeDuration, err := time.ParseDuration(rb.(string))
			if err != nil {
				return fmt.Errorf("invalid renew_before: %w", err)
			}

			if expiresAt-time.Now().Unix() < int64(renewBeforeDuration.Seconds()) {
				// Token will expire within renewBeforeDuration - force recreation
				if err := d.SetNewComputed("issued_at"); err != nil {
					return fmt.Errorf("failed to force new resource on field %q: %w", "issued_at", err)
				}
			}

			return nil
		},

		Schema: map[string]*schema.Schema{
			"account": {
				Type:        schema.TypeString,
				Description: "Account name. Defaults to the current account. I.e. the account configured on the `provider` block.",
				Optional:    true,
				ForceNew:    true,
			},
			"expires_in": {
				Type:         schema.TypeString,
				Description:  "Duration before the token will expire. Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`. E.g. `12h`, `7d`. Default: No expiration.",
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateDuration,
			},
			"renew_after": {
				Type:         schema.TypeString,
				Description:  "Duration to control token silent regeneration based on token age. Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`. If set, then the token will be regenerated if it is older than `renew_after`. I.e. if `currentDate - issued_at > renew_after`.",
				Optional:     true,
				ValidateFunc: validateDuration,
			},
			"renew_before": {
				Type:         schema.TypeString,
				Description:  "Duration to control token silent regeneration based on remaining token lifetime. If `expires_in` is set, Terraform will regenerate the token if `expires_at - currentDate < renew_before`. Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.",
				Optional:     true,
				ValidateFunc: validateDuration,
				RequiredWith: []string{"expires_in"},
			},
			"jwt": {
				Type:        schema.TypeString,
				Description: "The raw JWT.",
				Computed:    true,
				Sensitive:   true,
			},
			"issued_at": {
				Type:        schema.TypeString,
				Description: "Unix timestamp at which the token was issued.",
				Computed:    true,
				ForceNew:    true,
			},
			"expires_at": {
				Type:        schema.TypeString,
				Description: "If `expires_in` is set, Unix timestamp upon which the token will expire.",
				Computed:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceArgoCDAccountTokenCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*ServerInterface)
	if err := si.initClients(ctx); err != nil {
		return errorToDiagnostics("failed to init clients", err)
	}

	accountName, err := getAccount(ctx, si, d)
	if err != nil {
		return errorToDiagnostics("failed to get account", err)
	}

	opts := &account.CreateTokenRequest{
		Name: accountName,
	}

	var expiresIn int64

	_expiresIn, expiresInOk := d.GetOk("expires_in")
	if expiresInOk {
		ei := _expiresIn.(string)
		expiresInDuration, err := time.ParseDuration(ei)

		if err != nil {
			return errorToDiagnostics(fmt.Sprintf("token expiration duration (%s) for account %s could not be parsed", ei, accountName), err)
		}

		expiresIn = int64(expiresInDuration.Seconds())
		opts.ExpiresIn = expiresIn
	}

	_renewBefore, renewBeforeOk := d.GetOk("renew_before")
	if renewBeforeOk {
		rb := _renewBefore.(string)
		renewBeforeDuration, err := time.ParseDuration(rb)

		if err != nil {
			return errorToDiagnostics(fmt.Sprintf("token renewal duration (%s) for account %s could not be parsed", rb, accountName), err)
		}

		renewBefore := int64(renewBeforeDuration.Seconds())

		if renewBefore > expiresIn {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("renew_before (%d) cannot be greater than expires_in (%d) for account token", renewBefore, expiresIn),
				},
			}
		}
	}

	tokenMutexSecrets.Lock()
	resp, err := si.AccountClient.CreateToken(ctx, opts)
	tokenMutexSecrets.Unlock()

	if err != nil {
		return argoCDAPIError("create", "token for account", accountName, err)
	}

	token, err := jwt.ParseString(resp.GetToken())
	if err != nil {
		return errorToDiagnostics(fmt.Sprintf("token for account %s is not a valid jwt", accountName), err)
	}

	var claims jwt.StandardClaims
	if err = json.Unmarshal(token.RawClaims(), &claims); err != nil {
		return errorToDiagnostics(fmt.Sprintf("token claims for account %s could not be parsed", accountName), err)
	}

	if expiresInOk {
		if claims.ExpiresAt == nil {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("token claims expiration date for account %s is missing", accountName),
				},
			}
		} else {
			err = d.Set("expires_at", convertInt64ToString(claims.ExpiresAt.Unix()))
			if err != nil {
				return errorToDiagnostics(fmt.Sprintf("token claims expiration date for account %s could not be persisted to state", accountName), err)
			}
		}
	}

	if err = d.Set("issued_at", convertInt64ToString(claims.IssuedAt.Unix())); err != nil {
		return errorToDiagnostics(fmt.Sprintf("token claims issue date for account %s could not be persisted to state", accountName), err)
	}

	if err := d.Set("jwt", token.String()); err != nil {
		return errorToDiagnostics(fmt.Sprintf("token for account %s could not be persisted to state", accountName), err)
	}

	d.SetId(claims.ID)

	return resourceArgoCDAccountTokenRead(ctx, d, meta)
}

func resourceArgoCDAccountTokenRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*ServerInterface)
	if err := si.initClients(ctx); err != nil {
		return errorToDiagnostics("failed to init clients", err)
	}

	accountName, err := getAccount(ctx, si, d)
	if err != nil {
		return errorToDiagnostics("failed to get account", err)
	}

	tokenMutexConfiguration.RLock() // Yes, this is a different mutex - accounts are stored in `argocd-cm` whereas tokens are stored in `argocd-secret`
	_, err = si.AccountClient.GetAccount(ctx, &account.GetAccountRequest{
		Name: accountName,
	})
	tokenMutexConfiguration.RUnlock()

	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			// Delete token from state if account has been deleted in an out-of-band fashion
			d.SetId("")
			return nil
		} else {
			return argoCDAPIError("read", "account", accountName, err)
		}
	}

	return nil
}

func resourceArgoCDAccountTokenUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	accountName := d.Get("account").(string)

	var expiresIn int64

	_expiresIn, expiresInOk := d.GetOk("expires_in")
	if expiresInOk {
		ei := _expiresIn.(string)
		expiresInDuration, err := time.ParseDuration(ei)

		if err != nil {
			return errorToDiagnostics(fmt.Sprintf("token expiration duration (%s) for account %s could not be parsed", ei, accountName), err)
		}

		expiresIn = int64(expiresInDuration.Seconds())
	}

	_renewBefore, renewBeforeOk := d.GetOk("renew_before")
	if renewBeforeOk {
		rb := _renewBefore.(string)
		renewBeforeDuration, err := time.ParseDuration(rb)

		if err != nil {
			return errorToDiagnostics(fmt.Sprintf("token renewal duration (%s) for account %s could not be parsed", rb, accountName), err)
		}

		renewBefore := int64(renewBeforeDuration.Seconds())
		if renewBefore > expiresIn {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("renew_before (%d) cannot be greater than expires_in (%d) for account %s", renewBefore, expiresIn, accountName),
				},
			}
		}
	}

	return resourceArgoCDAccountTokenRead(ctx, d, meta)
}

func resourceArgoCDAccountTokenDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*ServerInterface)
	if err := si.initClients(ctx); err != nil {
		return errorToDiagnostics("failed to init clients", err)
	}

	accountName, err := getAccount(ctx, si, d)
	if err != nil {
		return errorToDiagnostics("failed to get account", err)
	}

	tokenMutexSecrets.Lock()
	_, err = si.AccountClient.DeleteToken(ctx, &account.DeleteTokenRequest{
		Name: accountName,
		Id:   d.Id(),
	})
	tokenMutexSecrets.Unlock()

	if err != nil && !strings.Contains(err.Error(), "NotFound") {
		return argoCDAPIError("delete", "token for account", accountName, err)
	}

	d.SetId("")

	return nil
}

func getAccount(ctx context.Context, si *ServerInterface, d *schema.ResourceData) (string, error) {
	accountName := d.Get("account").(string)
	if len(accountName) > 0 {
		return accountName, nil
	}

	userInfo, err := si.SessionClient.GetUserInfo(ctx, &session.GetUserInfoRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to get current account: %w", err)
	}

	return userInfo.Username, nil
}

package argocd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/project"
	"github.com/cristalhq/jwt/v3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceArgoCDProjectToken() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages ArgoCD project role JWT tokens. See [Project Roles](https://argo-cd.readthedocs.io/en/stable/user-guide/projects/#project-roles) for more info.",
		CreateContext: resourceArgoCDProjectTokenCreate,
		ReadContext:   resourceArgoCDProjectTokenRead,
		UpdateContext: resourceArgoCDProjectTokenUpdate,
		DeleteContext: resourceArgoCDProjectTokenDelete,
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
			"project": {
				Type:        schema.TypeString,
				Description: "The project associated with the token.",
				Required:    true,
				ForceNew:    true,
			},
			"role": {
				Type:        schema.TypeString,
				Description: "The name of the role in the project associated with the token.",
				Required:    true,
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
			"description": {
				Type:        schema.TypeString,
				Description: "Description of the token.",
				Optional:    true,
				ForceNew:    true,
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

func resourceArgoCDProjectTokenCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*ServerInterface)
	if err := si.initClients(ctx); err != nil {
		return errorToDiagnostics("failed to init clients", err)
	}

	projectName := d.Get("project").(string)
	role := d.Get("role").(string)
	opts := &project.ProjectTokenCreateRequest{
		Project: projectName,
		Role:    role,
	}

	if _, ok := tokenMutexProjectMap[projectName]; !ok {
		tokenMutexProjectMap[projectName] = &sync.RWMutex{}
	}

	if d, ok := d.GetOk("description"); ok {
		opts.Description = d.(string)
	}

	var expiresIn int64

	_expiresIn, expiresInOk := d.GetOk("expires_in")
	if expiresInOk {
		expiresInDuration, err := time.ParseDuration(_expiresIn.(string))
		if err != nil {
			return errorToDiagnostics(fmt.Sprintf("token expiration duration for project %s could not be parsed", projectName), err)
		}

		expiresIn = int64(expiresInDuration.Seconds())
		opts.ExpiresIn = expiresIn
	}

	_renewBefore, renewBeforeOk := d.GetOk("renew_before")
	if renewBeforeOk {
		renewBeforeDuration, err := time.ParseDuration(_renewBefore.(string))
		if err != nil {
			return errorToDiagnostics(fmt.Sprintf("token renewal duration for project %s could not be parsed", projectName), err)
		}

		renewBefore := int64(renewBeforeDuration.Seconds())

		if renewBefore > expiresIn {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("renew_before (%d) cannot be greater than expires_in (%d) for project %s", renewBefore, expiresIn, projectName),
				},
			}
		}
	}

	tokenMutexProjectMap[projectName].Lock()
	resp, err := si.ProjectClient.CreateToken(ctx, opts)
	tokenMutexProjectMap[projectName].Unlock()

	if err != nil {
		return argoCDAPIError("create", "token for project", projectName, err)
	}

	token, err := jwt.ParseString(resp.GetToken())
	if err != nil {
		return errorToDiagnostics(fmt.Sprintf("token for project %s is not a valid jwt", projectName), err)
	}

	var claims jwt.StandardClaims
	if err = json.Unmarshal(token.RawClaims(), &claims); err != nil {
		return errorToDiagnostics(fmt.Sprintf("token claims for project %s could not be parsed", projectName), err)
	}

	if claims.IssuedAt == nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("token claims issue date for project %s is missing", projectName),
			},
		}
	}

	if expiresInOk {
		if claims.ExpiresAt == nil {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("token claims expiration date for project %s is missing", projectName),
				},
			}
		} else {
			err = d.Set("expires_at", convertInt64ToString(claims.ExpiresAt.Unix()))
			if err != nil {
				return errorToDiagnostics(fmt.Sprintf("token claims expiration date for project %s could not be persisted to state", projectName), err)
			}
		}
	}

	if err = d.Set("issued_at", convertInt64ToString(claims.IssuedAt.Unix())); err != nil {
		return errorToDiagnostics(fmt.Sprintf("token claims issue date for project %s could not be persisted to state", projectName), err)
	}

	if err := d.Set("jwt", token.String()); err != nil {
		return errorToDiagnostics(fmt.Sprintf("token for project %s could not be persisted to state", projectName), err)
	}

	if claims.ID == "" {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("token claims ID for project %s is missing", projectName),
			},
		}
	}

	d.SetId(claims.ID)

	return resourceArgoCDProjectTokenRead(ctx, d, meta)
}

func resourceArgoCDProjectTokenRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*ServerInterface)
	if err := si.initClients(ctx); err != nil {
		return errorToDiagnostics("failed to init clients", err)
	}

	projectName := d.Get("project").(string)
	if _, ok := tokenMutexProjectMap[projectName]; !ok {
		tokenMutexProjectMap[projectName] = &sync.RWMutex{}
	}

	// Delete token from state if project has been deleted in an out-of-band fashion
	tokenMutexProjectMap[projectName].RLock()
	p, err := si.ProjectClient.Get(ctx, &project.ProjectQuery{
		Name: projectName,
	})
	tokenMutexProjectMap[projectName].RUnlock()

	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			d.SetId("")
			return nil
		}

		return argoCDAPIError("read", "project", projectName, err)
	}

	tokenMutexProjectMap[projectName].RLock()
	token, _, err := p.GetJWTToken(
		d.Get("role").(string),
		0,
		d.Id(),
	)
	tokenMutexProjectMap[projectName].RUnlock()

	if err != nil {
		// Token has been deleted in an out-of-band fashion
		d.SetId("")
		return nil
	}

	if err = d.Set("issued_at", convertInt64ToString(token.IssuedAt)); err != nil {
		return errorToDiagnostics(fmt.Sprintf("token claims issue date for project %s could not be persisted to state", projectName), err)
	}

	if err = d.Set("expires_at", convertInt64ToString(token.ExpiresAt)); err != nil {
		return errorToDiagnostics(fmt.Sprintf("token claims expiration date for project %s could not be persisted to state", projectName), err)
	}

	return nil
}

func resourceArgoCDProjectTokenUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	projectName := d.Get("project").(string)

	var expiresIn int64

	_expiresIn, expiresInOk := d.GetOk("expires_in")
	if expiresInOk {
		expiresInDuration, err := time.ParseDuration(_expiresIn.(string))
		if err != nil {
			return errorToDiagnostics(fmt.Sprintf("token expiration duration for project %s could not be parsed", projectName), err)
		}

		expiresIn = int64(expiresInDuration.Seconds())
	}

	_renewBefore, renewBeforeOk := d.GetOk("renew_before")
	if renewBeforeOk {
		renewBeforeDuration, err := time.ParseDuration(_renewBefore.(string))
		if err != nil {
			return errorToDiagnostics(fmt.Sprintf("token renewal duration for project %s could not be parsed", projectName), err)
		}

		renewBefore := int64(renewBeforeDuration.Seconds())
		if renewBefore > expiresIn {
			return errorToDiagnostics(fmt.Sprintf("renew_before (%d) cannot be greater than expires_in (%d) for project %s", renewBefore, expiresIn, projectName), err)
		}
	}

	return resourceArgoCDProjectTokenRead(ctx, d, meta)
}

func resourceArgoCDProjectTokenDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*ServerInterface)
	if err := si.initClients(ctx); err != nil {
		return errorToDiagnostics("failed to init clients", err)
	}

	projectName := d.Get("project").(string)

	if _, ok := tokenMutexProjectMap[projectName]; !ok {
		tokenMutexProjectMap[projectName] = &sync.RWMutex{}
	}

	tokenMutexProjectMap[projectName].Lock()

	_, err := si.ProjectClient.DeleteToken(ctx, &project.ProjectTokenDeleteRequest{
		Id:      d.Id(),
		Project: projectName,
		Role:    d.Get("role").(string),
	})

	tokenMutexProjectMap[projectName].Unlock()

	if err != nil {
		return argoCDAPIError("delete", "token for project", projectName, err)
	}

	d.SetId("")

	return nil
}

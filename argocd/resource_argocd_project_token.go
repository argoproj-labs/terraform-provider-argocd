package argocd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/project"
	application "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/cristalhq/jwt/v3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceArgoCDProjectToken() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceArgoCDProjectTokenCreate,
		ReadContext:   resourceArgoCDProjectTokenRead,
		UpdateContext: resourceArgoCDProjectTokenUpdate,
		DeleteContext: resourceArgoCDProjectTokenDelete,

		Schema: map[string]*schema.Schema{
			"project": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"role": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"expires_in": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateDuration,
			},
			"renew_before": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateDuration,
				RequiredWith: []string{"expires_in"},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"jwt": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"issued_at": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"expires_at": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func resourceArgoCDProjectTokenCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	server := meta.(*ServerInterface)
	if err := server.initClients(); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Failed to init clients"),
				Detail:   err.Error(),
			},
		}
	}
	c := *server.ProjectClient
	var claims jwt.StandardClaims
	var expiresIn int64

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
	_expiresIn, expiresInOk := d.GetOk("expires_in")
	if expiresInOk {
		expiresInDuration, err := time.ParseDuration(_expiresIn.(string))
		if err != nil {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("token expiration duration for project %s could not be parsed", projectName),
					Detail:   err.Error(),
				},
			}
		}
		expiresIn = int64(expiresInDuration.Seconds())
		opts.ExpiresIn = expiresIn
	}

	_renewBefore, renewBeforeOk := d.GetOk("renew_before")
	if renewBeforeOk {
		renewBeforeDuration, err := time.ParseDuration(_renewBefore.(string))
		if err != nil {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("token renewal duration for project %s could not be parsed", projectName),
					Detail:   err.Error(),
				},
			}
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
		// Arbitrary protection against misconfiguration
		if 300 > expiresIn-renewBefore {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  "token will expire within 5 minutes, check your settings",
				},
			}
		}
	}

	featureTokenIDSupported, err := server.isFeatureSupported(featureTokenIDs)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "explicit token ID support could be not be checked",
				Detail:   err.Error(),
			},
		}
	}

	tokenMutexProjectMap[projectName].Lock()
	resp, err := c.CreateToken(ctx, opts)
	// ensure issuedAt is unique upon multiple simultaneous resource creation invocations
	// as this is the unique ID for old tokens
	if !featureTokenIDSupported {
		time.Sleep(1 * time.Second)
	}
	tokenMutexProjectMap[projectName].Unlock()
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("token for project %s could not be created", projectName),
				Detail:   err.Error(),
			},
		}
	}
	token, err := jwt.ParseString(resp.GetToken())
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("token for project %s is not a valid jwt", projectName),
				Detail:   err.Error(),
			},
		}
	}
	if err := json.Unmarshal(token.RawClaims(), &claims); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("token claims for project %s could not be parsed", projectName),
				Detail:   err.Error(),
			},
		}
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
				return []diag.Diagnostic{
					{
						Severity: diag.Error,
						Summary:  fmt.Sprintf("token claims expiration date for project %s could not be persisted to state", projectName),
						Detail:   err.Error(),
					},
				}
			}
		}
	}
	if err = d.Set("issued_at", convertInt64ToString(claims.IssuedAt.Unix())); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("token claims issue date for project %s could not be persisted to state", projectName),
				Detail:   err.Error(),
			},
		}
	}
	if err := d.Set("jwt", token.String()); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("token for project %s could not be persisted to state", projectName),
				Detail:   err.Error(),
			},
		}
	}
	if featureTokenIDSupported {
		if claims.ID == "" {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("token claims ID for project %s is missing", projectName),
				},
			}
		}
		d.SetId(claims.ID)
	} else {
		d.SetId(fmt.Sprintf("%s-%s-%d", projectName, role, claims.IssuedAt.Unix()))
	}
	return resourceArgoCDProjectTokenRead(ctx, d, meta)
}

func resourceArgoCDProjectTokenRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var token *application.JWTToken
	var expiresIn int64
	var renewBefore int64
	var requestTokenID string
	var requestTokenIAT int64 = 0

	server := meta.(*ServerInterface)
	if err := server.initClients(); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Failed to init clients"),
				Detail:   err.Error(),
			},
		}
	}
	c := *server.ProjectClient
	projectName := d.Get("project").(string)
	if _, ok := tokenMutexProjectMap[projectName]; !ok {
		tokenMutexProjectMap[projectName] = &sync.RWMutex{}
	}

	// Delete token from state if project has been deleted in an out-of-band fashion
	tokenMutexProjectMap[projectName].RLock()
	p, err := c.Get(ctx, &project.ProjectQuery{
		Name: projectName,
	})
	tokenMutexProjectMap[projectName].RUnlock()

	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			d.SetId("")
			return nil
		} else {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("token for project %s could not be read", projectName),
					Detail:   err.Error(),
				},
			}
		}
	}

	featureTokenIDSupported, err := server.isFeatureSupported(featureTokenIDs)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "explicit token ID support could be not be checked",
				Detail:   err.Error(),
			},
		}
	}
	if featureTokenIDSupported {
		requestTokenID = d.Id()
	} else {
		iat, ok := d.GetOk("issued_at")
		if ok {
			requestTokenIAT, err = convertStringToInt64(iat.(string))
			if err != nil {
				return []diag.Diagnostic{
					{
						Severity: diag.Error,
						Summary:  fmt.Sprintf("token issue date for project %s could not be parsed", projectName),
						Detail:   err.Error(),
					},
				}
			}
		} else {
			d.SetId("")
			return nil
		}
	}

	tokenMutexProjectMap[projectName].RLock()
	token, _, err = p.GetJWTToken(
		d.Get("role").(string),
		requestTokenIAT,
		requestTokenID,
	)
	tokenMutexProjectMap[projectName].RUnlock()
	if err != nil {
		// Token has been deleted in an out-of-band fashion
		d.SetId("")
		return nil
	}

	computedExpiresIn := expiresIn - renewBefore
	if err := isValidToken(token, computedExpiresIn); err != nil {
		d.SetId("")
		return nil
	}
	if err = d.Set("issued_at", convertInt64ToString(token.IssuedAt)); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("token claims issue date for project %s could not be persisted to state", projectName),
				Detail:   err.Error(),
			},
		}
	}
	if err = d.Set("expires_at", convertInt64ToString(token.ExpiresAt)); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("token claims expiration date for project %s could not be persisted to state", projectName),
				Detail:   err.Error(),
			},
		}
	}
	return nil
}

func resourceArgoCDProjectTokenUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var expiresIn int64
	projectName := d.Get("project").(string)

	_expiresIn, expiresInOk := d.GetOk("expires_in")
	if expiresInOk {
		expiresInDuration, err := time.ParseDuration(_expiresIn.(string))
		if err != nil {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("token expiration duration for project %s could not be parsed", projectName),
					Detail:   err.Error(),
				},
			}
		}
		expiresIn = int64(expiresInDuration.Seconds())
	}

	_renewBefore, renewBeforeOk := d.GetOk("renew_before")
	if renewBeforeOk {
		renewBeforeDuration, err := time.ParseDuration(_renewBefore.(string))
		if err != nil {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("token renewal duration for project %s could not be parsed", projectName),
					Detail:   err.Error(),
				},
			}
		}
		renewBefore := int64(renewBeforeDuration.Seconds())
		if renewBefore > expiresIn {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("renew_before (%d) cannot be greater than expires_in (%d) for project %s", renewBefore, expiresIn, projectName),
					Detail:   err.Error(),
				},
			}
		}
	}
	return resourceArgoCDProjectRead(ctx, d, meta)
}

func resourceArgoCDProjectTokenDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	server := meta.(*ServerInterface)
	if err := server.initClients(); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Failed to init clients"),
				Detail:   err.Error(),
			},
		}
	}
	c := *server.ProjectClient

	projectName := d.Get("project").(string)
	role := d.Get("role").(string)
	opts := &project.ProjectTokenDeleteRequest{
		Project: projectName,
		Role:    role,
	}

	if _, ok := tokenMutexProjectMap[projectName]; !ok {
		tokenMutexProjectMap[projectName] = &sync.RWMutex{}
	}

	featureTokenIDSupported, err := server.isFeatureSupported(featureTokenIDs)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "explicit token ID support could be not be checked",
				Detail:   err.Error(),
			},
		}
	}

	if featureTokenIDSupported {
		opts.Id = d.Id()
	} else {
		if iat, ok := d.GetOk("issued_at"); ok {
			opts.Iat, err = convertStringToInt64(iat.(string))
			if err != nil {
				return []diag.Diagnostic{
					{
						Severity: diag.Error,
						Summary:  fmt.Sprintf("token issue date for project %s could not be parsed", projectName),
						Detail:   err.Error(),
					},
				}
			}
		}
	}

	tokenMutexProjectMap[projectName].Lock()
	if _, err := c.DeleteToken(ctx, opts); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("token for project %s could not be deleted", projectName),
				Detail:   err.Error(),
			},
		}
	}
	tokenMutexProjectMap[projectName].Unlock()
	d.SetId("")
	return nil
}

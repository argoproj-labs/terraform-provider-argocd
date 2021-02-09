package argocd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/argoproj/argo-cd/pkg/apiclient/project"
	application "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/cristalhq/jwt/v3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
	"sync"
	"time"
)

func resourceArgoCDProjectToken() *schema.Resource {
	return &schema.Resource{
		Create: resourceArgoCDProjectTokenCreate,
		Read:   resourceArgoCDProjectTokenRead,
		Delete: resourceArgoCDProjectTokenDelete,
		Update: resourceArgoCDProjectTokenUpdate,

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

func resourceArgoCDProjectTokenCreate(d *schema.ResourceData, meta interface{}) error {
	server := meta.(ServerInterface)
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
			return err
		}
		expiresIn = int64(expiresInDuration.Seconds())
		opts.ExpiresIn = expiresIn
	}

	_renewBefore, renewBeforeOk := d.GetOk("renew_before")
	if renewBeforeOk {
		renewBeforeDuration, err := time.ParseDuration(_renewBefore.(string))
		if err != nil {
			return err
		}
		renewBefore := int64(renewBeforeDuration.Seconds())
		if renewBefore > expiresIn {
			return fmt.Errorf("renew_before (%d) cannot be greater than expires_in (%d)", renewBefore, expiresIn)
		}
		// Arbitrary protection against misconfiguration
		if 300 > expiresIn-renewBefore {
			return fmt.Errorf("token will expire within 5 minutes, check your settings")
		}
	}

	featureTokenIDSupported, err := server.isFeatureSupported(featureTokenIDs)
	if err != nil {
		return err
	}

	tokenMutexProjectMap[projectName].Lock()
	resp, err := c.CreateToken(context.Background(), opts)
	// ensure issuedAt is unique upon multiple simultaneous resource creation invocations
	// as this is the unique ID for old tokens
	if !featureTokenIDSupported {
		time.Sleep(1 * time.Second)
	}
	tokenMutexProjectMap[projectName].Unlock()
	if err != nil {
		return err
	}
	token, err := jwt.ParseString(resp.GetToken())
	if err != nil {
		return err
	}
	if err := json.Unmarshal(token.RawClaims(), &claims); err != nil {
		return err
	}
	if claims.IssuedAt == nil {
		return fmt.Errorf("returned issued_at is nil")
	}
	if expiresInOk {
		switch claims.ExpiresAt {
		case nil:
			return fmt.Errorf("returned expires_at is nil")
		default:
			err = d.Set("expires_at", convertInt64ToString(claims.ExpiresAt.Unix()))
			if err != nil {
				return fmt.Errorf("error persisting 'expires_at' attribute to state: %s", err)
			}
		}
	}
	if err = d.Set("issued_at", convertInt64ToString(claims.IssuedAt.Unix())); err != nil {
		return fmt.Errorf("error persisting 'issued_at' attribute to state: %s", err)
	}
	if err := d.Set("jwt", token.String()); err != nil {
		return err
	}
	if featureTokenIDSupported {
		if claims.ID == "" {
			return fmt.Errorf("ID claim is empty")
		}
		d.SetId(claims.ID)
	} else {
		d.SetId(fmt.Sprintf("%s-%s-%d", projectName, role, claims.IssuedAt.Unix()))
	}
	return resourceArgoCDProjectTokenRead(d, meta)
}

func resourceArgoCDProjectTokenRead(d *schema.ResourceData, meta interface{}) error {
	var token *application.JWTToken
	var expiresIn int64
	var renewBefore int64
	var requestTokenID string
	var requestTokenIAT int64 = 0

	server := meta.(ServerInterface)
	c := *server.ProjectClient
	projectName := d.Get("project").(string)
	if _, ok := tokenMutexProjectMap[projectName]; !ok {
		tokenMutexProjectMap[projectName] = &sync.RWMutex{}
	}

	// Delete token from state if project has been deleted in an out-of-band fashion
	tokenMutexProjectMap[projectName].RLock()
	p, err := c.Get(context.Background(), &project.ProjectQuery{
		Name: projectName,
	})
	tokenMutexProjectMap[projectName].RUnlock()

	if err != nil {
		switch strings.Contains(err.Error(), "NotFound") {
		case true:
			d.SetId("")
			return nil
		default:
			return err
		}
	}

	featureTokenIDSupported, err := server.isFeatureSupported(featureTokenIDs)
	if err != nil {
		return err
	}
	if featureTokenIDSupported {
		requestTokenID = d.Id()
	} else {
		iat, ok := d.GetOk("issued_at")
		if ok {
			requestTokenIAT, err = convertStringToInt64(iat.(string))
			if err != nil {
				return err
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
		return fmt.Errorf("could not persist 'issued_at' in state: %s", err)
	}
	if err = d.Set("expires_at", convertInt64ToString(token.ExpiresAt)); err != nil {
		return fmt.Errorf("could not persist 'expires_at' in state: %s", err)
	}
	return nil
}

func resourceArgoCDProjectTokenUpdate(d *schema.ResourceData, meta interface{}) error {
	var expiresIn int64

	_expiresIn, expiresInOk := d.GetOk("expires_in")
	if expiresInOk {
		expiresInDuration, err := time.ParseDuration(_expiresIn.(string))
		if err != nil {
			return err
		}
		expiresIn = int64(expiresInDuration.Seconds())
	}

	_renewBefore, renewBeforeOk := d.GetOk("renew_before")
	if renewBeforeOk {
		renewBeforeDuration, err := time.ParseDuration(_renewBefore.(string))
		if err != nil {
			return err
		}
		renewBefore := int64(renewBeforeDuration.Seconds())
		if renewBefore > expiresIn {
			return fmt.Errorf("renew_before (%d) cannot be greater than expires_in (%d)", renewBefore, expiresIn)
		}
	}
	return resourceArgoCDProjectRead(d, meta)
}

func resourceArgoCDProjectTokenDelete(d *schema.ResourceData, meta interface{}) error {
	server := meta.(ServerInterface)
	c := *server.ProjectClient

	p := d.Get("project").(string)
	role := d.Get("role").(string)
	opts := &project.ProjectTokenDeleteRequest{
		Project: p,
		Role:    role,
	}

	if _, ok := tokenMutexProjectMap[p]; !ok {
		tokenMutexProjectMap[p] = &sync.RWMutex{}
	}

	featureTokenIDSupported, err := server.isFeatureSupported(featureTokenIDs)
	if err != nil {
		return err
	}

	if featureTokenIDSupported {
		opts.Id = d.Id()
	} else {
		if iat, ok := d.GetOk("issued_at"); ok {
			opts.Iat, err = convertStringToInt64(iat.(string))
			if err != nil {
				return err
			}
		}
	}

	tokenMutexProjectMap[p].Lock()
	if _, err := c.DeleteToken(context.Background(), opts); err != nil {
		return err
	}
	tokenMutexProjectMap[p].Unlock()
	d.SetId("")
	return nil
}

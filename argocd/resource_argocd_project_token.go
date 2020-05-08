package argocd

import (
	"context"
	"encoding/json"
	"fmt"
	argoCDProject "github.com/argoproj/argo-cd/pkg/apiclient/project"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/util"
	"github.com/cristalhq/jwt/v2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"strings"
	"sync"
	"time"
)

// For each project, implement a sync.RWMutex
var tokenMutexProjectMap = make(map[string]*sync.RWMutex, 0)

func resourceArgoCDProjectToken() *schema.Resource {
	return &schema.Resource{
		Create: resourceArgoCDProjectTokenCreate,
		Read:   resourceArgoCDProjectTokenRead,
		Delete: resourceArgoCDProjectTokenDelete,

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
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
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
	apiClient := server.ApiClient

	project := d.Get("project").(string)
	role := d.Get("role").(string)

	opts := &argoCDProject.ProjectTokenCreateRequest{
		Project: project,
		Role:    role,
	}

	if _, ok := tokenMutexProjectMap[project]; !ok {
		tokenMutexProjectMap[project] = &sync.RWMutex{}
	}

	if d, ok := d.GetOk("description"); ok {
		opts.Description = d.(string)
	}
	expiresIn, expiresInOk := d.GetOk("expires_in")
	if expiresInOk {
		opts.ExpiresIn = int64(expiresIn.(int))
	}

	closer, c, err := apiClient.NewProjectClient()
	if err != nil {
		return err
	}
	defer util.Close(closer)

	featureTokenIDSupported, err := server.isFeatureSupported(featureTokenIDs)
	if err != nil {
		return err
	}

	tokenMutexProjectMap[project].Lock()
	resp, err := c.CreateToken(context.Background(), opts)
	// ensure issuedAt is unique upon multiple simultaneous resource creation invocations
	// as this is the unique ID for old tokens
	if !featureTokenIDSupported {
		time.Sleep(1 * time.Second)
	}
	tokenMutexProjectMap[project].Unlock()
	if err != nil {
		return err
	}

	token, err := jwt.ParseString(resp.GetToken())
	if err != nil {
		return err
	}

	var claims jwt.StandardClaims
	if err := json.Unmarshal(token.RawClaims(), &claims); err != nil {
		return err
	}
	if claims.IssuedAt == nil {
		return fmt.Errorf("returned issued_at is nil")
	}
	_ = d.Set("issued_at", claims.IssuedAt.String())

	if expiresInOk {
		switch claims.ExpiresAt {
		case nil:
			return fmt.Errorf("returned expires_at is nil")
		default:
			_ = d.Set("expires_at", claims.ExpiresAt.String())
		}
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
		d.SetId(fmt.Sprintf("%s-%s-%s", project, role, claims.IssuedAt.String()))
	}
	return resourceArgoCDProjectTokenRead(d, meta)
}

func resourceArgoCDProjectTokenRead(d *schema.ResourceData, meta interface{}) error {
	var token *v1alpha1.JWTToken
	var requestTokenID string
	var requestTokenIAT int64 = 0

	server := meta.(ServerInterface)
	apiClient := server.ApiClient
	closer, c, err := apiClient.NewProjectClient()
	if err != nil {
		return err
	}
	defer util.Close(closer)

	// Delete token from state if project has been deleted in an out-of-band fashion
	project, err := c.Get(context.Background(), &argoCDProject.ProjectQuery{
		Name: d.Get("project").(string),
	})
	if err != nil {
		switch strings.Contains(err.Error(), "NotFound") {
		case true:
			d.SetId("")
			return nil
		default:
			return err
		}
	}

	if _, ok := tokenMutexProjectMap[project.Name]; !ok {
		tokenMutexProjectMap[project.Name] = &sync.RWMutex{}
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
		}
	}

	tokenMutexProjectMap[project.Name].RLock()
	token, _, err = project.GetJWTToken(
		d.Get("role").(string),
		requestTokenIAT,
		requestTokenID,
	)
	tokenMutexProjectMap[project.Name].RUnlock()
	if err != nil {
		// Token has been deleted in an out-of-band fashion
		d.SetId("")
		return nil
	}
	_ = d.Set("issued_at", convertInt64ToString(token.IssuedAt))
	_ = d.Set("expires_at", convertInt64ToString(token.ExpiresAt))
	return nil
}

func resourceArgoCDProjectTokenDelete(d *schema.ResourceData, meta interface{}) error {
	server := meta.(ServerInterface)
	apiClient := server.ApiClient
	closer, c, err := apiClient.NewProjectClient()
	if err != nil {
		return err
	}
	defer util.Close(closer)

	project := d.Get("project").(string)
	role := d.Get("role").(string)
	opts := &argoCDProject.ProjectTokenDeleteRequest{
		Project: project,
		Role:    role,
	}

	if _, ok := tokenMutexProjectMap[project]; !ok {
		tokenMutexProjectMap[project] = &sync.RWMutex{}
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

	tokenMutexProjectMap[project].Lock()
	if _, err := c.DeleteToken(context.Background(), opts); err != nil {
		return err
	}
	tokenMutexProjectMap[project].Unlock()
	d.SetId("")
	return nil
}

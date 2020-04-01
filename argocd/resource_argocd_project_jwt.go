package argocd

import (
	"context"
	"fmt"
	argoCDApiClient "github.com/argoproj/argo-cd/pkg/apiclient"
	argoCDProject "github.com/argoproj/argo-cd/pkg/apiclient/project"
	"github.com/argoproj/argo-cd/util"
	"github.com/argoproj/argo-cd/util/jwt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	jwtGo "github.com/square/go-jose/jwt"
	"strconv"
)

func resourceArgoCDProjectToken() *schema.Resource {
	return &schema.Resource{
		Create: resourceArgoCDProjectJWTCreate,
		Read:   resourceArgoCDProjectJWTRead,
		Delete: resourceArgoCDProjectJWTDelete,

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

func resourceArgoCDProjectJWTCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(argoCDApiClient.Client)
	project := d.Get("project").(string)
	role := d.Get("role").(string)

	opts := &argoCDProject.ProjectTokenCreateRequest{
		Project: project,
		Role:    role,
	}

	if d, ok := d.GetOk("description"); ok {
		opts.Description = d.(string)
	}
	if d, ok := d.GetOk("expires_in"); ok {
		exp, err := strconv.ParseInt(d.(string), 10, 64)
		if err != nil {
			return err
		}
		opts.ExpiresIn = exp
	}

	closer, c, err := client.NewProjectClient()
	if err != nil {
		return err
	}
	defer util.Close(closer)

	resp, err := c.CreateToken(context.Background(), opts)
	if err != nil {
		return err
	}
	// TODO: check signing algorithm, issuer (argocd) and subject (should be proj:PROJECT:ROLE)
	token, err := jwtGo.ParseSigned(resp.GetToken())
	if err != nil {
		return err
	}

	claims := make(map[string]interface{})
	if err := token.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return err
	}

	iat, err := jwt.GetIssuedAt(claims)
	if err != nil {
		return err
	}
	_ = d.Set("issued_at", strconv.FormatInt(iat, 10))

	exp := jwt.GetField(claims, "exp")
	if exp != "" {
		_ = d.Set("expires_at", exp)
	}

	if err := d.Set("jwt", resp.GetToken()); err != nil {
		return err
	}
	d.SetId(fmt.Sprintf("%s-%s-%d", project, role, iat))
	return resourceArgoCDProjectJWTRead(d, meta)
}

func resourceArgoCDProjectJWTRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(argoCDApiClient.Client)
	closer, c, err := client.NewProjectClient()
	if err != nil {
		return err
	}
	defer util.Close(closer)

	project, err := c.Get(context.Background(), &argoCDProject.ProjectQuery{
		Name: d.Get("project").(string),
	})
	if err != nil {
		return err
	}
	_iat, ok := d.GetOk("issued_at")
	switch ok {
	case false:
		d.SetId("")
	default:
		iat, err := strconv.ParseInt(_iat.(string), 10, 64)
		if err != nil {
			return err
		}
		token, _, err := project.GetJWTToken(d.Get("role").(string), iat)
		if err != nil {
			// Token has been deleted in an out-of-band fashion
			d.SetId("")
			return nil
		}
		// TODO: check for signature, ask for ArgoCD devs to implement HS256 sig alg, and/or check that a session can be created with that token meaning its signature is validated by the server

		_ = d.Set("issued_at", strconv.FormatInt(token.IssuedAt, 10))
		_ = d.Set("expires_at", strconv.FormatInt(token.ExpiresAt, 10))
	}
	return nil
}

func resourceArgoCDProjectJWTDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(argoCDApiClient.Client)
	closer, c, err := client.NewProjectClient()
	if err != nil {
		return err
	}
	defer util.Close(closer)

	if _iat, ok := d.GetOk("issued_at"); ok {
		iat, err := strconv.ParseInt(_iat.(string), 10, 64)
		if err != nil {
			return err
		}
		opts := &argoCDProject.ProjectTokenDeleteRequest{
			Project: d.Get("project").(string),
			Role:    d.Get("role").(string),
			Iat:     iat,
		}
		if _, err := c.DeleteToken(context.Background(), opts); err != nil {
			return err
		}
		d.SetId("")
	}
	return nil
}

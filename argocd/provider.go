package argocd

import (
	"context"
	argoCDApiClient "github.com/argoproj/argo-cd/pkg/apiclient"
	"github.com/argoproj/argo-cd/pkg/apiclient/session"
	"github.com/argoproj/argo-cd/util"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"server_addr": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARGOCD_SERVER", nil),
			},
			"auth_token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARGOCD_AUTH_TOKEN", nil),
				ConflictsWith: []string{
					"username",
					"password",
				},
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARGOCD_AUTH_USERNAME", nil),
				ConflictsWith: []string{
					"auth_token",
				},
				AtLeastOneOf: []string{
					"password",
				},
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARGOCD_AUTH_PASSWORD", nil),
				ConflictsWith: []string{
					"auth_token",
				},
				AtLeastOneOf: []string{
					"username",
				},
			},
			"cert_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"plain_text": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"context": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARGOCD_CONTEXT", nil),
			},
			"user_agent": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"grpc_web": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"port_forward": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"port_forward_with_namespace": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"headers": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARGOCD_INSECURE", false),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"argocd_project":       resourceArgoCDProject(),
			"argocd_project_token": resourceArgoCDProjectToken(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	opts := argoCDApiClient.ClientOptions{}

	if d, ok := d.GetOk("server_addr"); ok {
		opts.ServerAddr = d.(string)
	}
	if d, ok := d.GetOk("plain_text"); ok {
		opts.PlainText = d.(bool)
	}
	if d, ok := d.GetOk("insecure"); ok {
		opts.Insecure = d.(bool)
	}
	if d, ok := d.GetOk("cert_file"); ok {
		opts.CertFile = d.(string)
	}
	if d, ok := d.GetOk("context"); ok {
		opts.Context = d.(string)
	}
	if d, ok := d.GetOk("user_agent"); ok {
		opts.UserAgent = d.(string)
	}
	if d, ok := d.GetOk("grpc_web"); ok {
		opts.GRPCWeb = d.(bool)
	}
	if d, ok := d.GetOk("port_forward"); ok {
		opts.PortForward = d.(bool)
	}
	if d, ok := d.GetOk("port_forward_with_namespace"); ok {
		opts.PortForwardNamespace = d.(string)
	}
	if d, ok := d.GetOk("headers"); ok {
		opts.Headers = d.([]string)
	}

	authToken, authTokenOk := d.GetOk("auth_token")

	switch authTokenOk {
	case true:
		opts.AuthToken = authToken.(string)
	case false:
		userName, userNameOk := d.GetOk("username")
		password, passwordOk := d.GetOk("password")
		if userNameOk && passwordOk {
			c, err := argoCDApiClient.NewClient(&opts)
			if err != nil {
				return c, err
			}
			closer, sc, err := c.NewSessionClient()
			if err != nil {
				return c, err
			}
			defer util.Close(closer)
			sessionOpts := session.SessionCreateRequest{
				Username: userName.(string),
				Password: password.(string),
			}
			resp, err := sc.Create(context.Background(), &sessionOpts)
			if err != nil {
				return c, err
			}
			opts.AuthToken = resp.Token
		}
	}

	client, err := argoCDApiClient.NewClient(&opts)
	return client, err
}

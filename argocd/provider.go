package argocd

import (
	"context"
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/argoproj/argo-cd/pkg/apiclient"
	"github.com/argoproj/argo-cd/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/pkg/apiclient/project"
	"github.com/argoproj/argo-cd/pkg/apiclient/repository"
	"github.com/argoproj/argo-cd/pkg/apiclient/session"
	util "github.com/argoproj/gitops-engine/pkg/utils/io"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var apiClientConnOpts apiclient.ClientOptions

func Provider(doneCh chan bool) terraform.ResourceProvider {
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
					"auth_token",
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
					"auth_token",
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
			"argocd_application":   resourceArgoCDApplication(),
			"argocd_project":       resourceArgoCDProject(),
			"argocd_project_token": resourceArgoCDProjectToken(),
			"argocd_repository":    resourceArgoCDRepository(),
		},
		ConfigureFunc: func(d *schema.ResourceData) (interface{}, error) {
			apiClient, err := initApiClient(d)
			if err != nil {
				return nil, err
			}
			pcCloser, projectClient, err := apiClient.NewProjectClient()
			if err != nil {
				return nil, err
			}

			acCloser, applicationClient, err := apiClient.NewApplicationClient()
			if err != nil {
				return nil, err
			}

			rcCloser, repositoryClient, err := apiClient.NewRepoClient()
			if err != nil {
				return nil, err
			}

			// Clients connection pooling, close when the provider execution ends
			go func(done chan bool) {
				<-done
				util.Close(pcCloser)
				util.Close(acCloser)
				util.Close(rcCloser)
			}(doneCh)
			return initServerInterface(
				apiClient,
				projectClient,
				applicationClient,
				repositoryClient,
			)
		},
	}
}

func initServerInterface(
	apiClient apiclient.Client,
	projectClient project.ProjectServiceClient,
	applicationClient application.ApplicationServiceClient,
	repositoryClient repository.RepositoryServiceClient,
) (interface{}, error) {
	acCloser, versionClient, err := apiClient.NewVersionClient()
	if err != nil {
		return nil, err
	}
	defer util.Close(acCloser)

	serverVersionMessage, err := versionClient.Version(context.Background(), &empty.Empty{})
	if err != nil {
		return nil, err
	}
	if serverVersionMessage == nil {
		return nil, fmt.Errorf("could not get server version information")
	}
	serverVersion, err := semver.NewVersion(serverVersionMessage.Version)
	if err != nil {
		return nil, fmt.Errorf("could not parse server semantic version: %s", serverVersionMessage.Version)
	}

	return ServerInterface{
		&apiClient,
		&applicationClient,
		&projectClient,
		&repositoryClient,
		serverVersion,
		serverVersionMessage}, err
}

func initApiClient(d *schema.ResourceData) (
	apiClient apiclient.Client,
	err error) {

	var opts apiclient.ClientOptions

	if v, ok := d.GetOk("server_addr"); ok {
		opts.ServerAddr = v.(string)
	}
	if v, ok := d.GetOk("plain_text"); ok {
		opts.PlainText = v.(bool)
	}
	if v, ok := d.GetOk("insecure"); ok {
		opts.Insecure = v.(bool)
	}
	if v, ok := d.GetOk("cert_file"); ok {
		opts.CertFile = v.(string)
	}
	if v, ok := d.GetOk("context"); ok {
		opts.Context = v.(string)
	}
	if v, ok := d.GetOk("user_agent"); ok {
		opts.UserAgent = v.(string)
	}
	if v, ok := d.GetOk("grpc_web"); ok {
		opts.GRPCWeb = v.(bool)
	}
	if v, ok := d.GetOk("port_forward"); ok {
		opts.PortForward = v.(bool)
	}
	if v, ok := d.GetOk("port_forward_with_namespace"); ok {
		opts.PortForwardNamespace = v.(string)
	}
	if v, ok := d.GetOk("headers"); ok {
		opts.Headers = v.([]string)
	}

	// Export provider API client connections options for use in other spawned api clients
	apiClientConnOpts = opts

	authToken, authTokenOk := d.GetOk("auth_token")
	switch authTokenOk {
	case true:
		opts.AuthToken = authToken.(string)
	case false:
		userName, userNameOk := d.GetOk("username")
		password, passwordOk := d.GetOk("password")
		if userNameOk && passwordOk {
			apiClient, err = apiclient.NewClient(&opts)
			if err != nil {
				return apiClient, err
			}
			closer, sc, err := apiClient.NewSessionClient()
			if err != nil {
				return apiClient, err
			}
			defer util.Close(closer)
			sessionOpts := session.SessionCreateRequest{
				Username: userName.(string),
				Password: password.(string),
			}
			resp, err := sc.Create(context.Background(), &sessionOpts)
			if err != nil {
				return apiClient, err
			}
			opts.AuthToken = resp.Token
		}
	}
	return apiclient.NewClient(&opts)
}

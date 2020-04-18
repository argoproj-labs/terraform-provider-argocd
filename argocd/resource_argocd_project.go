package argocd

import (
	argoCDApiClient "github.com/argoproj/argo-cd/pkg/apiclient"
	argoCDProject "github.com/argoproj/argo-cd/pkg/apiclient/project"
	"github.com/argoproj/argo-cd/util"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"strconv"
)

func resourceArgoCDProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceArgoCDProjectCreate,
		Read:   resourceArgoCDProjectRead,
		Update: resourceArgoCDProjectUpdate,
		Delete: resourceArgoCDProjectDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"namespace": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "argocd",
				ForceNew: true,
			},
			"cluster_resource_whitelist": {
				Type:     schema.TypeSet,
				Set:      schema.HashSchema(&schema.Schema{Type: schema.TypeMap}),
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeMap},
				// TODO: add a validatefunc
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"destinations": {
				Type:     schema.TypeSet,
				Set:      schema.HashSchema(&schema.Schema{Type: schema.TypeMap}),
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeMap},
				// TODO: add a validatefunc
			},
			"namespace_resource_blacklist": {
				Type:     schema.TypeSet,
				Set:      schema.HashSchema(&schema.Schema{Type: schema.TypeMap}),
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeMap},
				// TODO: add a validatefunc
			},
			"orphaned_resources": {
				Type:     schema.TypeMap,
				Optional: true,
				// TODO: add a validatefunc
			},
			"roles": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"groups": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"jwtTokens": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      schema.HashSchema(&schema.Schema{Type: schema.TypeMap}),
							// TODO: add a Diffsuppressfunc to allow for argocd_project_token resources to coexist
							//DiffSuppressFunc:
							// TODO: add a validatefunc
							Elem: &schema.Schema{Type: schema.TypeMap},
						},
						"policies": {
							Type:     schema.TypeList,
							Optional: true,
							// TODO: add a validatefunc
							Elem: &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"source_repos": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      schema.HashString,
				// TODO: add a validatefunc
				Elem: &schema.Schema{Type: schema.TypeString},
			},
			"sync_windows": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"applications": {
							Type: schema.TypeSet,
							Set:  schema.HashString,
							Elem: schema.Schema{Type: schema.TypeString},
						},
						"clusters": {
							Type: schema.TypeSet,
							Set:  schema.HashString,
							Elem: schema.Schema{Type: schema.TypeString},
						},
						"duration": {
							Type: schema.TypeString,
						},
						"kind": {
							Type: schema.TypeString,
						},
						"manual_sync": {
							Type: schema.TypeBool,
						},
						"namespaces": {
							Type: schema.TypeSet,
							Set:  schema.HashString,
							Elem: schema.Schema{Type: schema.TypeString},
						},
						"schedule": {
							Type: schema.TypeString,
						},
					},
				},
			},
		},
	}
}

func resourceArgoCDProjectCreate(d *schema.ResourceData, meta interface{}) error {
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
	return resourceArgoCDProjectRead(d, meta)
}

func resourceArgoCDProjectRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(argoCDApiClient.Client)
	closer, c, err := client.NewProjectClient()
	if err != nil {
		return err
	}
	defer util.Close(closer)
	return nil
}

func resourceArgoCDProjectUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(argoCDApiClient.Client)
	closer, c, err := client.NewProjectClient()
	if err != nil {
		return err
	}
	defer util.Close(closer)
	return resourceArgoCDProjectRead(d, meta)
}

func resourceArgoCDProjectDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(argoCDApiClient.Client)
	closer, c, err := client.NewProjectClient()
	if err != nil {
		return err
	}
	defer util.Close(closer)
	return nil
}

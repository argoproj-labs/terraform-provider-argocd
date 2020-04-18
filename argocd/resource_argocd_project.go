package argocd

import (
	"context"
	"encoding/json"
	argoCDApiClient "github.com/argoproj/argo-cd/pkg/apiclient"
	argoCDProject "github.com/argoproj/argo-cd/pkg/apiclient/project"
	argoCDAppv1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/util"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func resourceArgoCDProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceArgoCDProjectCreate,
		Read:   resourceArgoCDProjectRead,
		//Update: resourceArgoCDProjectUpdate,
		//Delete: resourceArgoCDProjectDelete,

		Schema: map[string]*schema.Schema{
			"metadata": {
				Type:        schema.TypeMap,
				Description: "Kubernetes resource metadata, such as name, namespace, annotations. At least name and namespace are required",
				Required:    true,
			},
			"spec": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: schema.Resource{
					Schema: map[string]*schema.Schema{
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
				},
			},
		},
	}
}

func resourceArgoCDProjectCreate(d *schema.ResourceData, meta interface{}) error {
	// Expand project metadata
	m := d.Get("metadata")
	objectMeta := k8smetav1.ObjectMeta{}
	if err := json.Unmarshal(m.([]byte), &objectMeta); err != nil {
		return err
	}
	// Expand project spec
	s := d.Get("spec")
	spec := argoCDAppv1.AppProjectSpec{}
	if err := json.Unmarshal(s.([]byte), &spec); err != nil {
		return err
	}

	client := meta.(argoCDApiClient.Client)
	closer, c, err := client.NewProjectClient()
	if err != nil {
		return err
	}
	defer util.Close(closer)

	p, err := c.Create(context.Background(), &argoCDProject.ProjectCreateRequest{
		Project: &argoCDAppv1.AppProject{
			ObjectMeta: objectMeta,
			Spec:       spec,
		},
		// TODO: remember to investigate upsert behavior
		Upsert: false,
	})

	if err != nil {
		return err
	}
	d.SetId(p.Name)
	return resourceArgoCDProjectRead(d, meta)
}

func resourceArgoCDProjectRead(d *schema.ResourceData, meta interface{}) error {
	m := d.Get("metadata")
	objectMeta := k8smetav1.ObjectMeta{}
	if err := json.Unmarshal(m.([]byte), &objectMeta); err != nil {
		return err
	}

	client := meta.(argoCDApiClient.Client)
	closer, c, err := client.NewProjectClient()
	if err != nil {
		return err
	}
	defer util.Close(closer)

	p, err := c.Get(context.Background(), &argoCDProject.ProjectQuery{
		Name: d.Id(),
	})
	if err != nil {
		return err
	}
	if err := d.Set("metadata", p.ObjectMeta); err != nil {
		return err
	}
	if err := d.Set("spec", p.Spec); err != nil {
		return err
	}
	return nil
}

//
//func resourceArgoCDProjectUpdate(d *schema.ResourceData, meta interface{}) error {
//	client := meta.(argoCDApiClient.Client)
//	closer, c, err := client.NewProjectClient()
//	if err != nil {
//		return err
//	}
//	defer util.Close(closer)
//	return resourceArgoCDProjectRead(d, meta)
//}
//
//func resourceArgoCDProjectDelete(d *schema.ResourceData, meta interface{}) error {
//	client := meta.(argoCDApiClient.Client)
//	closer, c, err := client.NewProjectClient()
//	if err != nil {
//		return err
//	}
//	defer util.Close(closer)
//	return nil
//}

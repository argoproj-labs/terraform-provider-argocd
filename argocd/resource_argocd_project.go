package argocd

import (
	"context"
	"fmt"
	argoCDApiClient "github.com/argoproj/argo-cd/pkg/apiclient"
	argoCDProject "github.com/argoproj/argo-cd/pkg/apiclient/project"
	argoCDAppv1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/util"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/mitchellh/mapstructure"
	//"github.com/mitchellh/copystructure"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func resourceArgoCDProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceArgoCDProjectCreate,
		Read:   resourceArgoCDProjectRead,
		Update: resourceArgoCDProjectUpdate,
		Delete: resourceArgoCDProjectDelete,
		// TODO: add importer

		Schema: map[string]*schema.Schema{
			"metadata": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Kubernetes resource metadata, such as name, namespace, annotations. At least name and namespace are required",
				Required:    true,
				// TODO: add validatefunc to ensure name/namespace are present
			},
			"spec": {
				Type:     schema.TypeList,
				MinItems: 1,
				Required: true,
				Elem: &schema.Resource{
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
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"server": {
										Type:     schema.TypeString,
										Required: true,
									},
									"namespace": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
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
							Type:     schema.TypeList,
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
									"jwt_tokens": {
										Type:     schema.TypeList,
										Optional: true,
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
										Type:     schema.TypeSet,
										Set:      schema.HashString,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"clusters": {
										Type:     schema.TypeSet,
										Set:      schema.HashString,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"duration": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"kind": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"manual_sync": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"namespaces": {
										Type:     schema.TypeSet,
										Set:      schema.HashString,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"schedule": {
										Type:     schema.TypeString,
										Optional: true,
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

// TODO: needs comprehensive unit tests
func expandArgoCDProject(d *schema.ResourceData) (k8smetav1.ObjectMeta, argoCDAppv1.AppProjectSpec, error) {
	objectMeta := k8smetav1.ObjectMeta{}
	spec := argoCDAppv1.AppProjectSpec{}

	// Expand project metadata
	m := d.Get("metadata")
	if err := mapstructure.Decode(m, &objectMeta); err != nil {
		return objectMeta, spec, fmt.Errorf("metadata expansion: %s | %v", err, m)
	}
	// Expand project spec
	s := d.Get("spec.0")
	if err := mapstructure.Decode(s, &spec); err != nil {
		return objectMeta, spec, fmt.Errorf("spec expansion: %s | %v ", err, s)
	}
	return objectMeta, spec, nil
}

func resourceArgoCDProjectCreate(d *schema.ResourceData, meta interface{}) error {
	objectMeta, spec, err := expandArgoCDProject(d)
	if err != nil {
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
	// TODO: needs flattening function
	if err := d.Set("metadata", p.ObjectMeta); err != nil {
		return fmt.Errorf("error persisting metadata: %s | %v", err, p.ObjectMeta)
	}
	// TODO: needs flattening function
	if err := d.Set("spec.0", p.Spec); err != nil {
		return fmt.Errorf("error persisting spec: %s | %v", err, p.Spec)
	}
	return nil
}

func resourceArgoCDProjectUpdate(d *schema.ResourceData, meta interface{}) error {
	if ok := d.HasChanges("metadata", "spec"); ok {
		objectMeta, spec, err := expandArgoCDProject(d)
		if err != nil {
			return err
		}

		client := meta.(argoCDApiClient.Client)
		closer, c, err := client.NewProjectClient()
		if err != nil {
			return err
		}
		defer util.Close(closer)

		p, err := c.Update(context.Background(), &argoCDProject.ProjectUpdateRequest{
			Project: &argoCDAppv1.AppProject{
				ObjectMeta: objectMeta,
				Spec:       spec,
			}})
		if err != nil {
			return err
		}
		if err := d.Set("metadata", p.ObjectMeta); err != nil {
			return err
		}
		if err := d.Set("spec", p.Spec); err != nil {
			return err
		}
	}
	return resourceArgoCDProjectRead(d, meta)
}

func resourceArgoCDProjectDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(argoCDApiClient.Client)
	closer, c, err := client.NewProjectClient()
	if err != nil {
		return err
	}
	defer util.Close(closer)

	_, err = c.Delete(context.Background(), &argoCDProject.ProjectQuery{Name: d.Id()})
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}

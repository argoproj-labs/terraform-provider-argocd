package argocd

import (
	"context"
	"encoding/json"
	"fmt"
	argoCDApiClient "github.com/argoproj/argo-cd/pkg/apiclient"
	argoCDProject "github.com/argoproj/argo-cd/pkg/apiclient/project"
	argoCDAppv1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/util"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceArgoCDProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceArgoCDProjectCreate,
		Read:   resourceArgoCDProjectRead,
		Update: resourceArgoCDProjectUpdate,
		Delete: resourceArgoCDProjectDelete,
		// TODO: add an importer

		Schema: map[string]*schema.Schema{
			"metadata": metadataSchema(),
			"spec": {
				Type:        schema.TypeList,
				MinItems:    1,
				MaxItems:    1,
				Description: "ArgoCD App project resource specs. Required attributes: destinations, source_repos.",
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cluster_resource_whitelist": {
							Type:     schema.TypeSet,
							Set:      schema.HashSchema(&schema.Schema{Type: schema.TypeMap}),
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeMap},
							// TODO: add a validatefunc to ensure group and kind only are present
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
							Elem: &schema.Schema{
								Type: schema.TypeMap,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
							// TODO: add a validatefunc to ensure group and kind only are present
						},
						"orphaned_resources": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeBool,
							},
							// TODO: add a validatefunc to ensure only warn is present
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
										// TODO: add a Diffsuppressfunc to allow for argocd_project_token resources, and future named tokens to coexist
										//DiffSuppressFunc:
										// TODO: add a validatefunc to ensure issued_at, expires_at (and name?) only are present.
										Elem: &schema.Schema{
											Type: schema.TypeMap,
											Elem: &schema.Schema{Type: schema.TypeString},
										},
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

func storeArgoCDProjectToState(d *schema.ResourceData, p *argoCDAppv1.AppProject) error {
	if p == nil {
		return fmt.Errorf("project NPE")
	}
	f := flattenProject(p)
	if err := d.Set("metadata", f["metadata"]); err != nil {
		e, _ := json.MarshalIndent(f["metadata"], "", "\t")
		return fmt.Errorf("error persisting metadata: %s\n%s", err, e)
	}
	if err := d.Set("spec", f["spec"]); err != nil {
		e, _ := json.MarshalIndent(f["spec"], "", "\t")
		return fmt.Errorf("error persisting spec: %s\n%s", err, e)
	}
	return nil
}

func resourceArgoCDProjectCreate(d *schema.ResourceData, meta interface{}) error {
	objectMeta, spec, err := expandProject(d)
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
		// TODO: allow upsert instead of always requiring resource import?
		// TODO: make that a resource flag with proper acceptance tests
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
	return storeArgoCDProjectToState(d, p)
}

func resourceArgoCDProjectUpdate(d *schema.ResourceData, meta interface{}) error {
	if ok := d.HasChanges("metadata", "spec"); ok {
		objectMeta, spec, err := expandProject(d)
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
		if err := storeArgoCDProjectToState(d, p); err != nil {
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

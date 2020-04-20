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
	"strings"
	"time"
)

func resourceArgoCDProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceArgoCDProjectCreate,
		Read:   resourceArgoCDProjectRead,
		Update: resourceArgoCDProjectUpdate,
		Delete: resourceArgoCDProjectDelete,
		// TODO: add an importer

		Schema: map[string]*schema.Schema{
			"metadata": metadataSchema("appprojects.argoproj.io"),
			"spec":     projectSpecSchema(),
		},
	}
}

func storeArgoCDProjectToState(p *argoCDAppv1.AppProject, d *schema.ResourceData) error {
	if p == nil {
		return fmt.Errorf("project NPE")
	}
	f := flattenProject(p, d)
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

	p, err := c.Get(context.Background(), &argoCDProject.ProjectQuery{
		Name: objectMeta.Name,
	},
	)
	if err != nil {
		switch strings.Contains(err.Error(), "NotFound") {
		case true:
		default:
			return fmt.Errorf("foo %s", err)
		}
	}
	if p != nil {
		switch p.DeletionTimestamp {
		case nil:
		default:
			time.Sleep(time.Duration(*p.DeletionGracePeriodSeconds))
		}
	}
	p, err = c.Create(context.Background(), &argoCDProject.ProjectCreateRequest{
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
	if p == nil {
		return fmt.Errorf("something went wrong during project creation")
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
	if p == nil {
		d.SetId("")
		return nil
	}
	return storeArgoCDProjectToState(p, d)
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

		_, err = c.Update(context.Background(), &argoCDProject.ProjectUpdateRequest{
			Project: &argoCDAppv1.AppProject{
				ObjectMeta: objectMeta,
				Spec:       spec,
			}})
		if err != nil {
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

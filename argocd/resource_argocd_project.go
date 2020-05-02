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
			return err
		}
	}
	if p != nil {
		switch p.DeletionTimestamp {
		case nil:
		default:
			// Pre-existing project is still in Kubernetes soft deletion queue
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
		switch strings.Contains(err.Error(), "NotFound") {
		case true:
			d.SetId("")
			return nil
		default:
			return err
		}
	}

	fMetadata := flattenMetadata(p.ObjectMeta, d)
	fSpec, err := flattenProjectSpec(p.Spec)
	if err != nil {
		return err
	}

	if err := d.Set("spec", fSpec); err != nil {
		e, _ := json.MarshalIndent(fSpec, "", "\t")
		return fmt.Errorf("error persisting spec: %s\n%s", err, e)
	}
	if err := d.Set("metadata", fMetadata); err != nil {
		e, _ := json.MarshalIndent(fMetadata, "", "\t")
		return fmt.Errorf("error persisting metadata: %s\n%s", err, e)
	}

	return nil
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

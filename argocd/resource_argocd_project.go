package argocd

import (
	"context"
	"encoding/json"
	"fmt"
	projectClient "github.com/argoproj/argo-cd/pkg/apiclient/project"
	application "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
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
		// TODO: add importer acceptance tests
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
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
	server := meta.(ServerInterface)
	c := server.ProjectClient
	p, err := c.Get(context.Background(), &projectClient.ProjectQuery{
		Name: objectMeta.Name,
	})
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
	p, err = c.Create(context.Background(), &projectClient.ProjectCreateRequest{
		Project: &application.AppProject{
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
	server := meta.(ServerInterface)
	c := server.ProjectClient
	p, err := c.Get(context.Background(), &projectClient.ProjectQuery{
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
		server := meta.(ServerInterface)
		c := server.ProjectClient
		projectRequest := &projectClient.ProjectUpdateRequest{
			Project: &application.AppProject{
				ObjectMeta: objectMeta,
				Spec:       spec,
			}}
		p, err := c.Get(context.Background(), &projectClient.ProjectQuery{
			Name: d.Id(),
		})
		if p != nil {
			// Kubernetes API requires providing the up-to-date correct ResourceVersion for updates
			projectRequest.Project.ResourceVersion = p.ResourceVersion

			// Preserve preexisting JWTs for managed roles
			roles, err := expandProjectRoles(d.Get("spec.0.role").([]interface{}))
			if err != nil {
				return err
			}
			for _, r := range roles {
				pr, i, err := p.GetRoleByName(r.Name)
				if err != nil {
					return err
				}
				projectRequest.Project.Spec.Roles[i].JWTTokens = pr.JWTTokens
			}
		}
		_, err = c.Update(context.Background(), projectRequest)
		if err != nil {
			return err
		}
	}
	return resourceArgoCDProjectRead(d, meta)
}

func resourceArgoCDProjectDelete(d *schema.ResourceData, meta interface{}) error {
	server := meta.(ServerInterface)
	c := server.ProjectClient
	_, err := c.Delete(context.Background(), &projectClient.ProjectQuery{Name: d.Id()})
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

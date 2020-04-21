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

	f := flattenProject(p, d)
	fMetadata := f["metadata"]
	fSpec := f["spec"]

	if err := d.Set("metadata", fMetadata); err != nil {
		e, _ := json.MarshalIndent(fMetadata, "", "\t")
		return fmt.Errorf("error persisting metadata: %s\n%s", err, e)
	}

	// diffSupprFunc for roles' JWTTokens added out of Terraform
	managedJwtIatMap := make(map[string]string)
	currentJwtIatMap := make(map[string]string)
	
	// Merge old and new managed JWTs
	_oldRoles, _newRoles := d.GetChange("spec.0.role")
	_currentRoles := fSpec.([]map[string]interface{})[0]["role"]
	
	oldRoles := _oldRoles.(map[string]interface{})
	newRoles := _newRoles.(map[string]interface{})
	currentRoles := _currentRoles.(map[string]interface{})
	
	for rk, rv := range oldRoles {
		role := rv.(map[string]interface{})
		if rk == "jwt_token" {
			roleJwts := role["jwt_token"].([]map[string]string)
			for _, roleJwt := range roleJwts {
				managedJwtIatMap[roleJwt["iat"]] = role["name"].(string)
			}
		}
	}
	for rk, rv := range newRoles {
		role := rv.(map[string]interface{})
		if rk == "jwt_token" {
			roleJwts := role["jwt_token"].([]map[string]string)
			for _, roleJwt := range roleJwts {
				managedJwtIatMap[roleJwt["iat"]] = role["name"].(string)
			}
		}
	}
	
	for rk, rv := range currentRoles {
		role := rv.(map[string]interface{})
		if rk == "jwt_token" {
			roleJwts := role["jwt_token"].([]map[string]string)
			for _, roleJwt := range roleJwts {
				currentJwtIatMap[roleJwt["iat"]] = role["name"].(string)
			}
		}
	}

	for k, _ := range managedJwtIatMap {
		if _, ok := currentJwtIatMap[k]; ok {
			delete(currentJwtIatMap, k)
		}
	}

	// Modify the jwt_token array in the to-be-persisted roles to remove the unmanaged jwts
	
	if len(currentJwtIatMap) > 0 {
		currentJwtRoleNames := make(map[string]interface{})
		for _, v := range currentJwtIatMap {
			currentJwtRoleNames[v] = nil
		}
		
		filteredfSpecRoles := make([]map[string][]map[string]interface{})
		
		for iat, roleName := range currentJwtIatMap {
			 fs := fSpec.([]map[string]interface{})[0]
			 fsRoles := fs["role"].([]map[string]interface{})
			 
			 for ri, r := range fsRoles {
			 	
			 	switch r["name"] == roleName {
				case false:
				default:
			 		roleJwts := r["jwt_tokens"].([]map[string]string)
			 		for jwti, jwt := range roleJwts {
			 			if jwt["iat"] == iat {
						}
					}
				}
			 }
		}
	}

	if err := d.Set("spec", fSpec); err != nil {
		e, _ := json.MarshalIndent(fSpec, "", "\t")
		return fmt.Errorf("error persisting spec: %s\n%s", err, e)
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

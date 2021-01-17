package argocd

import (
	"context"
	"github.com/argoproj/argo-cd/pkg/apiclient/repocreds"
	application "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"strings"
)

func resourceArgoCDRepositoryCredentials() *schema.Resource {
	return &schema.Resource{
		Create: resourceArgoCDRepositoryCredentialsCreate,
		Read:   resourceArgoCDRepositoryCredentialsRead,
		Update: resourceArgoCDRepositoryCredentialsUpdate,
		Delete: resourceArgoCDRepositoryCredentialsDelete,
		// TODO: add importer acceptance tests
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: repositoryCredentialsSchema(),
	}
}

func resourceArgoCDRepositoryCredentialsCreate(d *schema.ResourceData, meta interface{}) error {
	server := meta.(ServerInterface)
	c := *server.RepoCredsClient
	repoCreds := expandRepositoryCredentials(d)

	tokenMutexConfiguration.Lock()
	rc, err := c.CreateRepositoryCredentials(
		context.Background(),
		&repocreds.RepoCredsCreateRequest{
			Creds:  repoCreds,
			Upsert: false,
		},
	)
	tokenMutexConfiguration.Unlock()

	if err != nil {
		return err
	}
	d.SetId(rc.URL)
	return resourceArgoCDRepositoryCredentialsRead(d, meta)
}

func resourceArgoCDRepositoryCredentialsRead(d *schema.ResourceData, meta interface{}) error {
	server := meta.(ServerInterface)
	c := *server.RepoCredsClient
	rc := application.RepoCreds{}

	tokenMutexConfiguration.RLock()
	rcl, err := c.ListRepositoryCredentials(context.Background(), &repocreds.RepoCredsQuery{
		Url: d.Id(),
	})
	tokenMutexConfiguration.RUnlock()

	if err != nil {
		// TODO: check for NotFound condition?
		return err
	}
	if rcl == nil {
		// Repository credentials have already been deleted in an out-of-band fashion
		d.SetId("")
		return nil
	}
	for i, _rc := range rcl.Items {
		if _rc.URL == d.Id() {
			rc = _rc
			break
		}
		// Repository credentials have already been deleted in an out-of-band fashion
		if i == len(rcl.Items)-1 {
			d.SetId("")
			return nil
		}
	}
	return flattenRepositoryCredentials(rc, d)
}

func resourceArgoCDRepositoryCredentialsUpdate(d *schema.ResourceData, meta interface{}) error {
	server := meta.(ServerInterface)
	c := *server.RepoCredsClient
	repoCreds := expandRepositoryCredentials(d)

	tokenMutexConfiguration.Lock()
	r, err := c.UpdateRepositoryCredentials(
		context.Background(),
		&repocreds.RepoCredsUpdateRequest{
			Creds: repoCreds},
	)
	tokenMutexConfiguration.Unlock()

	if err != nil {
		switch strings.Contains(err.Error(), "NotFound") {
		// Repository credentials have already been deleted in an out-of-band fashion
		case true:
			d.SetId("")
			return nil
		default:
			return err
		}
	}
	d.SetId(r.URL)
	return resourceArgoCDRepositoryCredentialsRead(d, meta)
}

func resourceArgoCDRepositoryCredentialsDelete(d *schema.ResourceData, meta interface{}) error {
	server := meta.(ServerInterface)
	c := *server.RepoCredsClient

	tokenMutexConfiguration.Lock()
	_, err := c.DeleteRepositoryCredentials(
		context.Background(),
		&repocreds.RepoCredsDeleteRequest{Url: d.Id()},
	)
	tokenMutexConfiguration.Unlock()

	if err != nil {
		switch strings.Contains(err.Error(), "NotFound") {
		// Repository credentials have already been deleted in an out-of-band fashion
		case true:
			d.SetId("")
			return nil
		default:
			return err
		}
	}
	d.SetId("")
	return nil
}

package argocd

import (
	"context"
	"fmt"
	"github.com/argoproj/argo-cd/pkg/apiclient/repository"
	application "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"strings"
)

func resourceArgoCDRepository() *schema.Resource {
	return &schema.Resource{
		Create: resourceArgoCDRepositoryCreate,
		Read:   resourceArgoCDRepositoryRead,
		Update: resourceArgoCDRepositoryUpdate,
		Delete: resourceArgoCDRepositoryDelete,
		// TODO: add importer acceptance tests
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: repositorySchema(),
	}
}

func resourceArgoCDRepositoryCreate(d *schema.ResourceData, meta interface{}) error {
	server := meta.(ServerInterface)
	c := server.RepositoryClient
	repo := expandRepository(d)
	r, err := c.CreateRepository(
		context.Background(),
		&repository.RepoCreateRequest{
			Repo:      repo,
			Upsert:    false,
			CredsOnly: false,
		},
	)
	if err != nil {
		return err
	}
	if r.ConnectionState.Status == application.ConnectionStatusFailed {
		return fmt.Errorf(
			"could not connect to repository %s: %s",
			repo.Repo,
			r.ConnectionState.Message,
		)
	}
	d.SetId(r.Repo)
	return resourceArgoCDRepositoryRead(d, meta)
}

func resourceArgoCDRepositoryRead(d *schema.ResourceData, meta interface{}) error {
	server := meta.(ServerInterface)
	c := server.RepositoryClient
	r, err := c.Get(context.Background(), &repository.RepoQuery{
		Repo:         d.Id(),
		ForceRefresh: false,
	})
	if err != nil {
		switch strings.Contains(err.Error(), "NotFound") {
		// Repository has already been deleted in an out-of-band fashion
		case true:
			d.SetId("")
			return nil
		default:
			return err
		}
	}
	return flattenRepository(r, d)
}

func resourceArgoCDRepositoryUpdate(d *schema.ResourceData, meta interface{}) error {
	server := meta.(ServerInterface)
	c := server.RepositoryClient
	repo := expandRepository(d)
	r, err := c.UpdateRepository(
		context.Background(),
		&repository.RepoUpdateRequest{Repo: repo},
	)
	if err != nil {
		switch strings.Contains(err.Error(), "NotFound") {
		// Repository has already been deleted in an out-of-band fashion
		case true:
			d.SetId("")
			return nil
		default:
			return err
		}
	}
	if r.ConnectionState.Status == application.ConnectionStatusFailed {
		return fmt.Errorf(
			"could not connect to repository %s: %s",
			repo.Repo,
			r.ConnectionState.Message,
		)
	}
	d.SetId(r.Repo)
	return resourceArgoCDRepositoryRead(d, meta)
}

func resourceArgoCDRepositoryDelete(d *schema.ResourceData, meta interface{}) error {
	server := meta.(ServerInterface)
	c := server.RepositoryClient
	_, err := c.DeleteRepository(
		context.Background(),
		&repository.RepoQuery{Repo: d.Id()},
	)
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

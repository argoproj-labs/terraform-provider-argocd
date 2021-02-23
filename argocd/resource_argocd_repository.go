package argocd

import (
	"context"
	"fmt"
	"strings"

	"github.com/argoproj/argo-cd/pkg/apiclient/repository"
	application "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceArgoCDRepository() *schema.Resource {
	return &schema.Resource{
		Create: resourceArgoCDRepositoryCreate,
		Read:   resourceArgoCDRepositoryRead,
		Update: resourceArgoCDRepositoryUpdate,
		Delete: resourceArgoCDRepositoryDelete,
		// TODO: add importer acceptance tests
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: repositorySchema(),
	}
}

func resourceArgoCDRepositoryCreate(d *schema.ResourceData, meta interface{}) error {
	server := meta.(ServerInterface)
	c := *server.RepositoryClient
	repo := expandRepository(d)

	tokenMutexConfiguration.Lock()
	r, err := c.CreateRepository(
		context.Background(),
		&repository.RepoCreateRequest{
			Repo:      repo,
			Upsert:    false,
			CredsOnly: false,
		},
	)
	tokenMutexConfiguration.Unlock()

	if err != nil {
		return err
	}
	if r == nil {
		return fmt.Errorf("ArgoCD did not return an error or a repository result")
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
	c := *server.RepositoryClient
	r := &application.Repository{}

	featureRepositoryGetSupported, err := server.isFeatureSupported(featureRepositoryGet)
	if err != nil {
		return err
	}

	switch featureRepositoryGetSupported {
	case true:
		tokenMutexConfiguration.RLock()
		r, err = c.Get(context.Background(), &repository.RepoQuery{
			Repo:         d.Id(),
			ForceRefresh: true,
		})
		tokenMutexConfiguration.RUnlock()

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
	case false:
		tokenMutexConfiguration.RLock()
		rl, err := c.ListRepositories(context.Background(), &repository.RepoQuery{
			Repo:         d.Id(),
			ForceRefresh: true,
		})
		tokenMutexConfiguration.RUnlock()

		if err != nil {
			// TODO: check for NotFound condition?
			return err
		}
		if rl == nil {
			// Repository has already been deleted in an out-of-band fashion
			d.SetId("")
			return nil
		}
		for i, _r := range rl.Items {
			if _r.Repo == d.Id() {
				r = _r
				break
			}
			// Repository has already been deleted in an out-of-band fashion
			if i == len(rl.Items)-1 {
				d.SetId("")
				return nil
			}
		}
	}
	return flattenRepository(r, d)
}

func resourceArgoCDRepositoryUpdate(d *schema.ResourceData, meta interface{}) error {
	server := meta.(ServerInterface)
	c := *server.RepositoryClient
	repo := expandRepository(d)

	tokenMutexConfiguration.Lock()
	r, err := c.UpdateRepository(
		context.Background(),
		&repository.RepoUpdateRequest{Repo: repo},
	)
	tokenMutexConfiguration.Unlock()

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
	if r == nil {
		return fmt.Errorf("ArgoCD did not return an error or a repository result")
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
	c := *server.RepositoryClient

	tokenMutexConfiguration.Lock()
	_, err := c.DeleteRepository(
		context.Background(),
		&repository.RepoQuery{Repo: d.Id()},
	)
	tokenMutexConfiguration.Unlock()

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
	d.SetId("")
	return nil
}

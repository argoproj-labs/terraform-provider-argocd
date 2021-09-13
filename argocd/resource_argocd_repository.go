package argocd

import (
	"context"
	"fmt"
	"strings"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/repository"
	application "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceArgoCDRepository() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceArgoCDRepositoryCreate,
		ReadContext:   resourceArgoCDRepositoryRead,
		UpdateContext: resourceArgoCDRepositoryUpdate,
		DeleteContext: resourceArgoCDRepositoryDelete,
		// TODO: add importer acceptance tests
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: repositorySchema(),
	}
}

func resourceArgoCDRepositoryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	server := meta.(*ServerInterface)
	if err := server.initClients(); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Failed to init clients"),
				Detail:   err.Error(),
			},
		}
	}

	c := *server.RepositoryClient
	repo := expandRepository(d)

	tokenMutexConfiguration.Lock()
	r, err := c.CreateRepository(
		ctx,
		&repository.RepoCreateRequest{
			Repo:   repo,
			Upsert: false,
		},
	)
	tokenMutexConfiguration.Unlock()

	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Repository %s not found", repo.Repo),
				Detail:   err.Error(),
			},
		}
	}
	if r == nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("ArgoCD did not return an error or a repository result"),
			},
		}
	}
	if r.ConnectionState.Status == application.ConnectionStatusFailed {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"could not connect to repository %s: %s",
					repo.Repo,
					r.ConnectionState.Message,
				),
			},
		}
	}
	d.SetId(r.Repo)
	return resourceArgoCDRepositoryRead(ctx, d, meta)
}

func resourceArgoCDRepositoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	server := meta.(*ServerInterface)
	if err := server.initClients(); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Failed to init clients"),
				Detail:   err.Error(),
			},
		}
	}
	c := *server.RepositoryClient
	r := &application.Repository{}

	featureRepositoryGetSupported, err := server.isFeatureSupported(featureRepositoryGet)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  `support for feature "repositoryGet" could not be checked`,
				Detail:   err.Error(),
			},
		}
	}

	if featureRepositoryGetSupported {
		tokenMutexConfiguration.RLock()
		r, err = c.Get(ctx, &repository.RepoQuery{
			Repo:         d.Id(),
			ForceRefresh: true,
		})
		tokenMutexConfiguration.RUnlock()

		if err != nil {
			// Repository has already been deleted in an out-of-band fashion
			if strings.Contains(err.Error(), "NotFound") {
				d.SetId("")
				return nil
			}
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("repository %s could not be retrieved", d.Id()),
					Detail:   err.Error(),
				},
			}
		}
	} else {
		tokenMutexConfiguration.RLock()
		rl, err := c.ListRepositories(ctx, &repository.RepoQuery{
			Repo:         d.Id(),
			ForceRefresh: true,
		})
		tokenMutexConfiguration.RUnlock()

		if err != nil {
			// TODO: check for NotFound condition?
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("repository %s could not be listed", d.Id()),
					Detail:   err.Error(),
				},
			}
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
	err = flattenRepository(r, d)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("repository %s could not be flattened", d.Id()),
				Detail:   err.Error(),
			},
		}
	}
	return nil
}

func resourceArgoCDRepositoryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	server := meta.(*ServerInterface)
	if err := server.initClients(); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Failed to init clients"),
				Detail:   err.Error(),
			},
		}
	}
	c := *server.RepositoryClient
	repo := expandRepository(d)

	tokenMutexConfiguration.Lock()
	r, err := c.UpdateRepository(
		ctx,
		&repository.RepoUpdateRequest{Repo: repo},
	)
	tokenMutexConfiguration.Unlock()

	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			// Repository has already been deleted in an out-of-band fashion
			d.SetId("")
			return nil
		}
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("repository %s could not be updated", d.Id()),
				Detail:   err.Error(),
			},
		}

	}
	if r == nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("argoCD did not return an error or a repository result for ID %s", d.Id()),
				Detail:   err.Error(),
			},
		}
	}
	if r.ConnectionState.Status == application.ConnectionStatusFailed {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("could not connect to repository %s: %s", repo.Repo, r.ConnectionState.Message),
			},
		}
	}
	d.SetId(r.Repo)
	return resourceArgoCDRepositoryRead(ctx, d, meta)
}

func resourceArgoCDRepositoryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	server := meta.(*ServerInterface)
	if err := server.initClients(); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Failed to init clients"),
				Detail:   err.Error(),
			},
		}
	}
	c := *server.RepositoryClient

	tokenMutexConfiguration.Lock()
	_, err := c.DeleteRepository(
		ctx,
		&repository.RepoQuery{Repo: d.Id()},
	)
	tokenMutexConfiguration.Unlock()

	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			// Repository has already been deleted in an out-of-band fashion
			d.SetId("")
			return nil
		}
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Repository %s not found", d.Id()),
				Detail:   err.Error(),
			},
		}
	}
	d.SetId("")
	return nil
}

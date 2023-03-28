package argocd

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/repository"
	application "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceArgoCDRepository() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages [repositories](https://argo-cd.readthedocs.io/en/stable/operator-manual/declarative-setup/#repositories) within ArgoCD.",
		CreateContext: resourceArgoCDRepositoryCreate,
		ReadContext:   resourceArgoCDRepositoryRead,
		UpdateContext: resourceArgoCDRepositoryUpdate,
		DeleteContext: resourceArgoCDRepositoryDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: repositorySchema(),
	}
}

func resourceArgoCDRepositoryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	server := meta.(*ServerInterface)
	if err := server.initClients(ctx); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "failed to init clients",
				Detail:   err.Error(),
			},
		}
	}

	c := *server.RepositoryClient

	repo, err := expandRepository(d)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("could not expand repository attributes: %s", err),
				Detail:   err.Error(),
			},
		}
	}

	featureProjectScopedRepositoriesSupported, err := server.isFeatureSupported(featureProjectScopedRepositories)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "feature not supported",
				Detail:   err.Error(),
			},
		}
	} else if !featureProjectScopedRepositoriesSupported && repo.Project != "" {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"repository project is only supported from ArgoCD %s onwards",
					featureVersionConstraintsMap[featureProjectScopedRepositories].String()),
				Detail: "See https://argo-cd.readthedocs.io/en/stable/user-guide/projects/#project-scoped-repositories-and-clusters",
			},
		}
	}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		tokenMutexConfiguration.Lock()

		var r *application.Repository

		r, err = c.CreateRepository(
			ctx,
			&repository.RepoCreateRequest{
				Repo:   repo,
				Upsert: false,
			},
		)
		tokenMutexConfiguration.Unlock()

		if err != nil {
			// TODO: better way to detect ssh handshake failing ?
			if matched, _ := regexp.MatchString("ssh: handshake failed: knownhosts: key is unknown", err.Error()); matched {
				return resource.RetryableError(fmt.Errorf("handshake failed for repository %s, retrying in case a repository certificate has been set recently", repo.Repo))
			}

			return resource.NonRetryableError(fmt.Errorf("repository %s not found: %s", repo.Repo, err))
		} else if r == nil {
			return resource.NonRetryableError(fmt.Errorf("ArgoCD did not return an error or a repository result: %s", err))
		} else if r.ConnectionState.Status == application.ConnectionStatusFailed {
			return resource.NonRetryableError(fmt.Errorf("could not connect to repository %s: %s", repo.Repo, r.ConnectionState.Message))
		}

		d.SetId(r.Repo)

		return nil
	})

	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Error while creating repository %s", repo.Name),
				Detail:   err.Error(),
			},
		}
	}

	return resourceArgoCDRepositoryRead(ctx, d, meta)
}

func resourceArgoCDRepositoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	server := meta.(*ServerInterface)
	if err := server.initClients(ctx); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "failed to init clients",
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
		var rl *application.RepositoryList

		tokenMutexConfiguration.RLock()
		rl, err = c.ListRepositories(ctx, &repository.RepoQuery{
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

	if err = flattenRepository(r, d); err != nil {
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
	if err := server.initClients(ctx); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "failed to init clients",
				Detail:   err.Error(),
			},
		}
	}

	c := *server.RepositoryClient

	repo, err := expandRepository(d)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("could not expand repository attributes: %s", err),
				Detail:   err.Error(),
			},
		}
	}

	featureProjectScopedRepositoriesSupported, err := server.isFeatureSupported(featureProjectScopedRepositories)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "feature not supported",
				Detail:   err.Error(),
			},
		}
	}

	if !featureProjectScopedRepositoriesSupported && repo.Project != "" {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"repository project is only supported from ArgoCD %s onwards",
					featureVersionConstraintsMap[featureProjectScopedRepositories].String()),
				Detail: "See https://argo-cd.readthedocs.io/en/stable/user-guide/projects/#project-scoped-repositories-and-clusters",
			},
		}
	}

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
	if err := server.initClients(ctx); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "failed to init clients",
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
				Summary:  fmt.Sprintf("repository %s could not be deleted", d.Id()),
				Detail:   err.Error(),
			},
		}
	}

	d.SetId("")

	return nil
}

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
	si := meta.(*ServerInterface)
	if err := si.initClients(ctx); err != nil {
		return errorToDiagnostics("failed to init clients", err)
	}

	repo, err := expandRepository(d)
	if err != nil {
		return errorToDiagnostics("failed to expand repository", err)
	}

	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		tokenMutexConfiguration.Lock()

		var r *application.Repository

		r, err := si.RepositoryClient.CreateRepository(
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

			return resource.NonRetryableError(err)
		} else if r == nil {
			return resource.NonRetryableError(fmt.Errorf("ArgoCD did not return an error or a repository result: %s", err))
		} else if r.ConnectionState.Status == application.ConnectionStatusFailed {
			return resource.NonRetryableError(fmt.Errorf("could not connect to repository %s: %s", repo.Repo, r.ConnectionState.Message))
		}

		d.SetId(r.Repo)

		return nil
	}); err != nil {
		return argoCDAPIError("create", "repository", repo.Repo, err)
	}

	return resourceArgoCDRepositoryRead(ctx, d, meta)
}

func resourceArgoCDRepositoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*ServerInterface)
	if err := si.initClients(ctx); err != nil {
		return errorToDiagnostics("failed to init clients", err)
	}

	tokenMutexConfiguration.RLock()
	r, err := si.RepositoryClient.Get(ctx, &repository.RepoQuery{
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

		return argoCDAPIError("read", "repository", d.Id(), err)
	}

	if err = flattenRepository(r, d); err != nil {
		return errorToDiagnostics(fmt.Sprintf("failed to flatten repository %s", d.Id()), err)
	}

	return nil
}

func resourceArgoCDRepositoryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*ServerInterface)
	if err := si.initClients(ctx); err != nil {
		return errorToDiagnostics("failed to init clients", err)
	}

	repo, err := expandRepository(d)
	if err != nil {
		return errorToDiagnostics(fmt.Sprintf("failed to expand repository %s", d.Id()), err)
	}

	tokenMutexConfiguration.Lock()
	r, err := si.RepositoryClient.UpdateRepository(
		ctx,
		&repository.RepoUpdateRequest{Repo: repo},
	)
	tokenMutexConfiguration.Unlock()

	if err != nil {
		return argoCDAPIError("update", "repository", repo.Repo, err)
	}

	if r == nil {
		return errorToDiagnostics(fmt.Sprintf("ArgoCD did not return an error or a repository result for ID %s", d.Id()), err)
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
	si := meta.(*ServerInterface)
	if err := si.initClients(ctx); err != nil {
		return errorToDiagnostics("failed to init clients", err)
	}

	tokenMutexConfiguration.Lock()
	_, err := si.RepositoryClient.DeleteRepository(
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

		return argoCDAPIError("delete", "repository", d.Id(), err)
	}

	d.SetId("")

	return nil
}

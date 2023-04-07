package argocd

import (
	"context"
	"fmt"
	"strings"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/repocreds"
	application "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceArgoCDRepositoryCredentials() *schema.Resource {
	return &schema.Resource{
		Description: "Manages [repository credentials](https://argo-cd.readthedocs.io/en/stable/user-guide/private-repositories/#credentials) within ArgoCD.\n\n" +
			"**Note**: due to restrictions in the ArgoCD API the provider is unable to track drift in this resource to fields other than `username`. I.e. the " +
			"provider is unable to detect changes to repository credentials that are made outside of Terraform (e.g. manual updates to the underlying Kubernetes " +
			"Secrets).",
		CreateContext: resourceArgoCDRepositoryCredentialsCreate,
		ReadContext:   resourceArgoCDRepositoryCredentialsRead,
		UpdateContext: resourceArgoCDRepositoryCredentialsUpdate,
		DeleteContext: resourceArgoCDRepositoryCredentialsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: repositoryCredentialsSchema(),
	}
}

func resourceArgoCDRepositoryCredentialsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	c := *server.RepoCredsClient

	repoCreds, err := expandRepositoryCredentials(d)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("could not expand repository credential attributes: %s", err),
				Detail:   err.Error(),
			},
		}
	}

	tokenMutexConfiguration.Lock()
	rc, err := c.CreateRepositoryCredentials(
		ctx,
		&repocreds.RepoCredsCreateRequest{
			Creds:  repoCreds,
			Upsert: false,
		},
	)
	tokenMutexConfiguration.Unlock()

	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("credentials for repository %s could not be created", repoCreds.URL),
				Detail:   err.Error(),
			},
		}
	}

	d.SetId(rc.URL)

	return resourceArgoCDRepositoryCredentialsRead(ctx, d, meta)
}

func resourceArgoCDRepositoryCredentialsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	c := *server.RepoCredsClient
	rc := application.RepoCreds{}

	tokenMutexConfiguration.RLock()
	rcl, err := c.ListRepositoryCredentials(ctx, &repocreds.RepoCredsQuery{
		Url: d.Id(),
	})
	tokenMutexConfiguration.RUnlock()

	if err != nil {
		// TODO: check for NotFound condition?
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("credentials for repository %s could not be listed", d.Id()),
				Detail:   err.Error(),
			},
		}
	} else if rcl == nil || len(rcl.Items) == 0 {
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

func resourceArgoCDRepositoryCredentialsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	c := *server.RepoCredsClient

	repoCreds, err := expandRepositoryCredentials(d)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("could not expand repository credential attributes: %s", err),
				Detail:   err.Error(),
			},
		}
	}

	tokenMutexConfiguration.Lock()
	r, err := c.UpdateRepositoryCredentials(
		ctx,
		&repocreds.RepoCredsUpdateRequest{
			Creds: repoCreds},
	)
	tokenMutexConfiguration.Unlock()

	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("credentials for repository %s could not be updated", repoCreds.URL),
				Detail:   err.Error(),
			},
		}
	}

	d.SetId(r.URL)

	return resourceArgoCDRepositoryCredentialsRead(ctx, d, meta)
}

func resourceArgoCDRepositoryCredentialsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	c := *server.RepoCredsClient

	tokenMutexConfiguration.Lock()
	_, err := c.DeleteRepositoryCredentials(
		ctx,
		&repocreds.RepoCredsDeleteRequest{Url: d.Id()},
	)
	tokenMutexConfiguration.Unlock()

	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			// Repository credentials have already been deleted in an out-of-band fashion
			d.SetId("")
			return nil
		}

		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("credentials for repository %s could not be deleted", d.Id()),
				Detail:   err.Error(),
			},
		}
	}

	d.SetId("")

	return nil
}

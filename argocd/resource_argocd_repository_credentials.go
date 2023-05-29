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
	si := meta.(*ServerInterface)
	if err := si.initClients(ctx); err != nil {
		return errorToDiagnostics("failed to init clients", err)
	}

	repoCreds, err := expandRepositoryCredentials(d)
	if err != nil {
		return errorToDiagnostics("failed to expand repository credentials", err)
	}

	tokenMutexConfiguration.Lock()
	rc, err := si.RepoCredsClient.CreateRepositoryCredentials(
		ctx,
		&repocreds.RepoCredsCreateRequest{
			Creds:  repoCreds,
			Upsert: false,
		},
	)
	tokenMutexConfiguration.Unlock()

	if err != nil {
		return argoCDAPIError("create", "repository credentials", repoCreds.URL, err)
	}

	d.SetId(rc.URL)

	return resourceArgoCDRepositoryCredentialsRead(ctx, d, meta)
}

func resourceArgoCDRepositoryCredentialsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*ServerInterface)
	if err := si.initClients(ctx); err != nil {
		return errorToDiagnostics("failed to init clients", err)
	}

	tokenMutexConfiguration.RLock()
	rcl, err := si.RepoCredsClient.ListRepositoryCredentials(ctx, &repocreds.RepoCredsQuery{
		Url: d.Id(),
	})
	tokenMutexConfiguration.RUnlock()

	if err != nil {
		return argoCDAPIError("read", "repository credentials", d.Id(), err)
	} else if rcl == nil || len(rcl.Items) == 0 {
		// Repository credentials have already been deleted in an out-of-band fashion
		d.SetId("")
		return nil
	}

	rc := application.RepoCreds{}

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
	si := meta.(*ServerInterface)
	if err := si.initClients(ctx); err != nil {
		return errorToDiagnostics("failed to init clients", err)
	}

	repoCreds, err := expandRepositoryCredentials(d)
	if err != nil {
		return errorToDiagnostics(fmt.Sprintf("failed to expand repository credentials %s", d.Id()), err)
	}

	tokenMutexConfiguration.Lock()
	r, err := si.RepoCredsClient.UpdateRepositoryCredentials(
		ctx,
		&repocreds.RepoCredsUpdateRequest{
			Creds: repoCreds},
	)
	tokenMutexConfiguration.Unlock()

	if err != nil {
		return argoCDAPIError("update", "repository credentials", repoCreds.URL, err)
	}

	d.SetId(r.URL)

	return resourceArgoCDRepositoryCredentialsRead(ctx, d, meta)
}

func resourceArgoCDRepositoryCredentialsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*ServerInterface)
	if err := si.initClients(ctx); err != nil {
		return errorToDiagnostics("failed to init clients", err)
	}

	tokenMutexConfiguration.Lock()
	_, err := si.RepoCredsClient.DeleteRepositoryCredentials(
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

		return argoCDAPIError("delete", "repository credentials", d.Id(), err)
	}

	d.SetId("")

	return nil
}

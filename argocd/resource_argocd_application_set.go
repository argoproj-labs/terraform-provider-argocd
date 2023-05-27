package argocd

import (
	"context"
	"fmt"
	"strings"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/applicationset"
	application "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func resourceArgoCDApplicationSet() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages [application sets](https://argo-cd.readthedocs.io/en/stable/user-guide/application-set/) within ArgoCD.",
		CreateContext: resourceArgoCDApplicationSetCreate,
		ReadContext:   resourceArgoCDApplicationSetRead,
		UpdateContext: resourceArgoCDApplicationSetUpdate,
		DeleteContext: resourceArgoCDApplicationSetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"metadata": metadataSchema("applicationsets.argoproj.io"),
			"spec":     applicationSetSpecSchemaV0(),
		},
	}
}

func resourceArgoCDApplicationSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*ServerInterface)
	if err := si.initClients(ctx); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "failed to init clients",
				Detail:   err.Error(),
			},
		}
	}

	if !si.isFeatureSupported(featureApplicationSet) {
		return featureNotSupported(featureApplicationSet)
	}

	objectMeta, spec, err := expandApplicationSet(d, si.isFeatureSupported(featureMultipleApplicationSources))
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("application set %s could not be created", d.Id()),
				Detail:   err.Error(),
			},
		}
	}

	if !si.isFeatureSupported(featureApplicationSetProgressiveSync) && spec.Strategy != nil {
		return featureNotSupported(featureApplicationSetProgressiveSync)
	}

	as, err := si.ApplicationSetClient.Create(ctx, &applicationset.ApplicationSetCreateRequest{
		Applicationset: &application.ApplicationSet{
			ObjectMeta: objectMeta,
			Spec:       spec,
			TypeMeta: metav1.TypeMeta{
				Kind:       "ApplicationSet",
				APIVersion: "argoproj.io/v1alpha1",
			},
		},
	})
	if err != nil {
		return argoCDAPIError("create", "application set", objectMeta.Name, err)
	} else if as == nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("application set %s not created: unknown reason", objectMeta.Name),
			},
		}
	}

	d.SetId(as.Name)

	return resourceArgoCDApplicationSetRead(ctx, d, meta)
}

func resourceArgoCDApplicationSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*ServerInterface)
	if err := si.initClients(ctx); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "failed to init clients",
				Detail:   err.Error(),
			},
		}
	}

	name := d.Id()

	appSet, err := si.ApplicationSetClient.Get(ctx, &applicationset.ApplicationSetGetQuery{
		Name: name,
	})
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			d.SetId("")
			return diag.Diagnostics{}
		}

		return argoCDAPIError("read", "application set", name, err)
	}

	err = flattenApplicationSet(appSet, d)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("application set %s could not be flattened", name),
				Detail:   err.Error(),
			},
		}
	}

	return nil
}

func resourceArgoCDApplicationSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*ServerInterface)
	if err := si.initClients(ctx); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "failed to init clients",
				Detail:   err.Error(),
			},
		}
	}

	if !si.isFeatureSupported(featureApplicationSet) {
		return featureNotSupported(featureApplicationSet)
	}

	if !d.HasChanges("metadata", "spec") {
		return nil
	}

	objectMeta, spec, err := expandApplicationSet(d, si.isFeatureSupported(featureMultipleApplicationSources))
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("application set %s could not be updated", d.Id()),
				Detail:   err.Error(),
			},
		}
	}

	if !si.isFeatureSupported(featureApplicationSetProgressiveSync) && spec.Strategy != nil {
		return featureNotSupported(featureApplicationSetProgressiveSync)
	}

	_, err = si.ApplicationSetClient.Create(ctx, &applicationset.ApplicationSetCreateRequest{
		Applicationset: &application.ApplicationSet{
			ObjectMeta: objectMeta,
			Spec:       spec,
			TypeMeta: metav1.TypeMeta{
				Kind:       "ApplicationSet",
				APIVersion: "argoproj.io/v1alpha1",
			},
		},
		Upsert: true,
	})

	if err != nil {
		return argoCDAPIError("update", "application set", objectMeta.Name, err)
	}

	return resourceArgoCDApplicationSetRead(ctx, d, meta)
}

func resourceArgoCDApplicationSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*ServerInterface)
	if err := si.initClients(ctx); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "failed to init clients",
				Detail:   err.Error(),
			},
		}
	}

	_, err := si.ApplicationSetClient.Delete(ctx, &applicationset.ApplicationSetDeleteRequest{
		Name: d.Id(),
	})

	if err != nil && !strings.Contains(err.Error(), "NotFound") {
		return argoCDAPIError("delete", "application set", d.Id(), err)
	}

	d.SetId("")

	return nil
}

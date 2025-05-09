package argocd

import (
	"context"
	"fmt"
	"strings"

	"github.com/argoproj-labs/terraform-provider-argocd/internal/features"
	"github.com/argoproj-labs/terraform-provider-argocd/internal/provider"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/applicationset"
	application "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
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
			"spec":     applicationSetSpecSchemaV1(),
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceArgoCDApplicationV1().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceArgoCDApplicationSetStateUpgradeV0,
				Version: 0,
			},
		},
	}
}

func resourceArgoCDApplicationSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*provider.ServerInterface)
	if diags := si.InitClients(ctx); diags != nil {
		return pluginSDKDiags(diags)
	}

	if !si.IsFeatureSupported(features.ApplicationSet) {
		return featureNotSupported(features.ApplicationSet)
	}

	objectMeta, spec, err := expandApplicationSet(d, si.IsFeatureSupported(features.MultipleApplicationSources), si.IsFeatureSupported(features.ApplicationSetIgnoreApplicationDifferences), si.IsFeatureSupported(features.ApplicationSetTemplatePatch))
	if err != nil {
		return errorToDiagnostics("failed to expand application set", err)
	}

	if !si.IsFeatureSupported(features.ApplicationSetProgressiveSync) && spec.Strategy != nil {
		return featureNotSupported(features.ApplicationSetProgressiveSync)
	}

	if !si.IsFeatureSupported(features.ApplicationSetIgnoreApplicationDifferences) && spec.IgnoreApplicationDifferences != nil {
		return featureNotSupported(features.ApplicationSetIgnoreApplicationDifferences)
	}

	if !si.IsFeatureSupported(features.ApplicationSetApplicationsSyncPolicy) && spec.SyncPolicy != nil && spec.SyncPolicy.ApplicationsSync != nil {
		return featureNotSupported(features.ApplicationSetApplicationsSyncPolicy)
	}

	if !si.IsFeatureSupported(features.ApplicationSetTemplatePatch) && spec.TemplatePatch != nil {
		return featureNotSupported(features.ApplicationSetTemplatePatch)
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

	d.SetId(fmt.Sprintf("%s:%s", as.Name, objectMeta.Namespace))

	return resourceArgoCDApplicationSetRead(ctx, d, meta)
}

func resourceArgoCDApplicationSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*provider.ServerInterface)
	if diags := si.InitClients(ctx); diags != nil {
		return pluginSDKDiags(diags)
	}

	ids := strings.Split(d.Id(), ":")
	appSetName := ids[0]
	namespace := ids[1]

	appSet, err := si.ApplicationSetClient.Get(ctx, &applicationset.ApplicationSetGetQuery{
		Name:            appSetName,
		AppsetNamespace: namespace,
	})
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			d.SetId("")
			return diag.Diagnostics{}
		}

		return argoCDAPIError("read", "application set", appSetName, err)
	}

	err = flattenApplicationSet(appSet, d)
	if err != nil {
		return errorToDiagnostics(fmt.Sprintf("failed to flatten application set %s", appSetName), err)
	}

	return nil
}

func resourceArgoCDApplicationSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*provider.ServerInterface)
	if diags := si.InitClients(ctx); diags != nil {
		return pluginSDKDiags(diags)
	}

	if !si.IsFeatureSupported(features.ApplicationSet) {
		return featureNotSupported(features.ApplicationSet)
	}

	if !d.HasChanges("metadata", "spec") {
		return nil
	}

	objectMeta, spec, err := expandApplicationSet(d, si.IsFeatureSupported(features.MultipleApplicationSources), si.IsFeatureSupported(features.ApplicationSetIgnoreApplicationDifferences), si.IsFeatureSupported(features.ApplicationSetTemplatePatch))
	if err != nil {
		return errorToDiagnostics(fmt.Sprintf("failed to expand application set %s", d.Id()), err)
	}

	if !si.IsFeatureSupported(features.ApplicationSetProgressiveSync) && spec.Strategy != nil {
		return featureNotSupported(features.ApplicationSetProgressiveSync)
	}

	if !si.IsFeatureSupported(features.ApplicationSetIgnoreApplicationDifferences) && spec.IgnoreApplicationDifferences != nil {
		return featureNotSupported(features.ApplicationSetIgnoreApplicationDifferences)
	}

	if !si.IsFeatureSupported(features.ApplicationSetApplicationsSyncPolicy) && spec.SyncPolicy != nil && spec.SyncPolicy.ApplicationsSync != nil {
		return featureNotSupported(features.ApplicationSetApplicationsSyncPolicy)
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
	si := meta.(*provider.ServerInterface)
	if diags := si.InitClients(ctx); diags != nil {
		return pluginSDKDiags(diags)
	}

	ids := strings.Split(d.Id(), ":")
	appSetName := ids[0]
	namespace := ids[1]

	if _, err := si.ApplicationSetClient.Delete(ctx, &applicationset.ApplicationSetDeleteRequest{
		Name:            appSetName,
		AppsetNamespace: namespace,
	}); err != nil && !strings.Contains(err.Error(), "NotFound") {
		return argoCDAPIError("delete", "application set", appSetName, err)
	}

	d.SetId("")

	return nil
}

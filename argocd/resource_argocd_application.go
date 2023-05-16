package argocd

import (
	"context"
	"fmt"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	applicationClient "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	application "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/gitops-engine/pkg/health"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceArgoCDApplication() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages [applications](https://argo-cd.readthedocs.io/en/stable/operator-manual/declarative-setup/#applications) within ArgoCD.",
		CreateContext: resourceArgoCDApplicationCreate,
		ReadContext:   resourceArgoCDApplicationRead,
		UpdateContext: resourceArgoCDApplicationUpdate,
		DeleteContext: resourceArgoCDApplicationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"metadata": metadataSchema("applications.argoproj.io"),
			"spec":     applicationSpecSchemaV4(false),
			"wait": {
				Type:        schema.TypeBool,
				Description: "Upon application creation or update, wait for application health/sync status to be healthy/Synced, upon application deletion, wait for application to be removed, when set to true. Wait timeouts are controlled by Terraform Create, Update and Delete resource timeouts (all default to 5 minutes). **Note**: if ArgoCD decides not to sync an application (e.g. because the project to which the application belongs has a `sync_window` applied) then you will experience an expected timeout event if `wait = true`.",
				Optional:    true,
				Default:     false,
			},
			"cascade": {
				Type:        schema.TypeBool,
				Description: "Whether to applying cascading deletion when application is removed.",
				Optional:    true,
				Default:     true,
			},
		},
		SchemaVersion: 4,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceArgoCDApplicationV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceArgoCDApplicationStateUpgradeV0,
				Version: 0,
			},
			{
				Type:    resourceArgoCDApplicationV1().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceArgoCDApplicationStateUpgradeV1,
				Version: 1,
			},
			{
				Type:    resourceArgoCDApplicationV2().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceArgoCDApplicationStateUpgradeV2,
				Version: 2,
			},
			{
				Type:    resourceArgoCDApplicationV3().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceArgoCDApplicationStateUpgradeV3,
				Version: 3,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func resourceArgoCDApplicationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	objectMeta, spec, err := expandApplication(d)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("application %s could not be created", objectMeta.Name),
				Detail:   err.Error(),
			},
		}
	}

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

	apps, err := si.ApplicationClient.List(ctx, &applicationClient.ApplicationQuery{
		Name:         &objectMeta.Name,
		AppNamespace: &objectMeta.Namespace,
	})
	if err != nil && !strings.Contains(err.Error(), "NotFound") {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("failed to get application %s", objectMeta.Name),
				Detail:   err.Error(),
			},
		}
	}

	if apps != nil {
		l := len(apps.Items)

		switch {
		case l < 1:
			break
		case l == 1:
			switch apps.Items[0].DeletionTimestamp {
			case nil:
			default:
				// Pre-existing app is still in Kubernetes soft deletion queue
				time.Sleep(time.Duration(*apps.Items[0].DeletionGracePeriodSeconds))
			}
		case l > 1:
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("found multiple applications matching name '%s' and namespace '%s'", objectMeta.Name, objectMeta.Namespace),
				},
			}
		}
	}

	featureApplicationLevelSyncOptionsSupported, err := si.isFeatureSupported(featureApplicationLevelSyncOptions)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "feature not supported",
				Detail:   err.Error(),
			},
		}
	}

	if !featureApplicationLevelSyncOptionsSupported &&
		spec.SyncPolicy != nil &&
		spec.SyncPolicy.SyncOptions != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"application-level sync_options is only supported from ArgoCD %s onwards",
					featureVersionConstraintsMap[featureApplicationLevelSyncOptions].String()),
				Detail: err.Error(),
			},
		}
	}

	featureIgnoreDiffJQPathExpressionsSupported, err := si.isFeatureSupported(featureIgnoreDiffJQPathExpressions)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "feature not supported",
				Detail:   err.Error(),
			},
		}
	}

	hasJQPathExpressions := false

	if spec.IgnoreDifferences != nil {
		for _, id := range spec.IgnoreDifferences {
			if id.JQPathExpressions != nil {
				hasJQPathExpressions = true
			}
		}
	}

	if !featureIgnoreDiffJQPathExpressionsSupported && hasJQPathExpressions {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"jq path expressions are only supported from ArgoCD %s onwards",
					featureVersionConstraintsMap[featureIgnoreDiffJQPathExpressions].String()),
				Detail: err.Error(),
			},
		}
	}

	featureMultipleApplicationSourcesSupported, err := si.isFeatureSupported(featureMultipleApplicationSources)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "feature not supported",
				Detail:   err.Error(),
			},
		}
	} else if !featureMultipleApplicationSourcesSupported {
		if len(spec.Sources) > 1 {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary: fmt.Sprintf(
						"multiple application sources is only supported from ArgoCD %s onwards",
						featureVersionConstraintsMap[featureMultipleApplicationSources].String()),
				},
			}
		}

		spec.Source = &spec.Sources[0]
		spec.Sources = nil
	}

	featureApplicationHelmSkipCrdsSupported, err := si.isFeatureSupported(featureApplicationHelmSkipCrds)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "feature not supported",
				Detail:   err.Error(),
			},
		}
	}

	if !featureApplicationHelmSkipCrdsSupported {
		_, skipCrdsOk := d.GetOk("spec.0.source.0.helm.0.skip_crds")
		if skipCrdsOk {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary: fmt.Sprintf(
						"application helm skip_crds is only supported from ArgoCD %s onwards",
						featureVersionConstraintsMap[featureApplicationHelmSkipCrds].String()),
				},
			}
		}
	}

	app, err := si.ApplicationClient.Create(ctx, &applicationClient.ApplicationCreateRequest{
		Application: &application.Application{
			ObjectMeta: objectMeta,
			Spec:       spec,
			TypeMeta: metav1.TypeMeta{
				Kind:       "Application",
				APIVersion: "argoproj.io/v1alpha1",
			},
		},
	})
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("application %s could not be created", objectMeta.Name),
				Detail:   err.Error(),
			},
		}
	} else if app == nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("application %s could not be created: unknown reason", objectMeta.Name),
			},
		}
	}

	d.SetId(fmt.Sprintf("%s:%s", app.Name, objectMeta.Namespace))

	if wait, ok := d.GetOk("wait"); ok && wait.(bool) {
		if err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
			var list *application.ApplicationList
			if list, err = si.ApplicationClient.List(ctx, &applicationClient.ApplicationQuery{
				Name:         &app.Name,
				AppNamespace: &app.Namespace,
			}); err != nil {
				return resource.NonRetryableError(fmt.Errorf("error while waiting for application %s to be synced and healthy: %s", app.Name, err))
			}

			if len(list.Items) != 1 {
				return resource.NonRetryableError(fmt.Errorf("found unexpected number of applications matching name '%s' and namespace '%s'. Items: %d", app.Name, app.Namespace, len(list.Items)))
			}

			if list.Items[0].Status.Health.Status != health.HealthStatusHealthy {
				return resource.RetryableError(fmt.Errorf("expected application health status to be healthy but was %s", list.Items[0].Status.Health.Status))
			}

			if list.Items[0].Status.Sync.Status != application.SyncStatusCodeSynced {
				return resource.RetryableError(fmt.Errorf("expected application sync status to be synced but was %s", list.Items[0].Status.Sync.Status))
			}

			return nil
		}); err != nil {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Error while waiting for application %s to be created", objectMeta.Name),
					Detail:   err.Error(),
				},
			}
		}
	}

	return resourceArgoCDApplicationRead(ctx, d, meta)
}

func resourceArgoCDApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	ids := strings.Split(d.Id(), ":")
	appName := ids[0]
	namespace := ids[1]

	apps, err := si.ApplicationClient.List(ctx, &applicationClient.ApplicationQuery{
		Name:         &appName,
		AppNamespace: &namespace,
	})
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			d.SetId("")
			return diag.Diagnostics{}
		}

		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("failed to get application %s", appName),
				Detail:   err.Error(),
			},
		}
	}

	l := len(apps.Items)

	switch {
	case l < 1:
		d.SetId("")
		return diag.Diagnostics{}
	case l == 1:
		break
	case l > 1:
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("found multiple applications matching name '%s' and namespace '%s'", appName, namespace),
			},
		}
	}

	err = flattenApplication(&apps.Items[0], d)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("application %s could not be flattened", appName),
				Detail:   err.Error(),
			},
		}
	}

	return nil
}

func resourceArgoCDApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if ok := d.HasChanges("metadata", "spec"); !ok {
		return resourceArgoCDApplicationRead(ctx, d, meta)
	}

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

	ids := strings.Split(d.Id(), ":")
	appQuery := &applicationClient.ApplicationQuery{
		Name:         &ids[0],
		AppNamespace: &ids[1],
	}

	objectMeta, spec, err := expandApplication(d)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("application %s could not be updated", *appQuery.Name),
				Detail:   err.Error(),
			},
		}
	}

	featureApplicationLevelSyncOptionsSupported, err := si.isFeatureSupported(featureApplicationLevelSyncOptions)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Feature not supported",
				Detail:   err.Error(),
			},
		}
	}

	if !featureApplicationLevelSyncOptionsSupported &&
		spec.SyncPolicy != nil &&
		spec.SyncPolicy.SyncOptions != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"application-level sync_options is only supported from ArgoCD %s onwards",
					featureVersionConstraintsMap[featureApplicationLevelSyncOptions].String()),
				Detail: err.Error(),
			},
		}
	}

	featureIgnoreDiffJQPathExpressionsSupported, err := si.isFeatureSupported(featureIgnoreDiffJQPathExpressions)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "feature not supported",
				Detail:   err.Error(),
			},
		}
	}

	hasJQPathExpressions := false

	if spec.IgnoreDifferences != nil {
		for _, id := range spec.IgnoreDifferences {
			if id.JQPathExpressions != nil {
				hasJQPathExpressions = true
			}
		}
	}

	if !featureIgnoreDiffJQPathExpressionsSupported && hasJQPathExpressions {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"jq path expressions are only supported from ArgoCD %s onwards",
					featureVersionConstraintsMap[featureIgnoreDiffJQPathExpressions].String()),
				Detail: err.Error(),
			},
		}
	}

	featureMultipleApplicationSourcesSupported, err := si.isFeatureSupported(featureMultipleApplicationSources)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "feature not supported",
				Detail:   err.Error(),
			},
		}
	} else if !featureMultipleApplicationSourcesSupported {
		if len(spec.Sources) > 1 {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary: fmt.Sprintf(
						"multiple application sources is only supported from ArgoCD %s onwards",
						featureVersionConstraintsMap[featureMultipleApplicationSources].String()),
				},
			}
		}

		spec.Source = &spec.Sources[0]
		spec.Sources = nil
	}

	featureApplicationHelmSkipCrdsSupported, err := si.isFeatureSupported(featureApplicationHelmSkipCrds)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "feature not supported",
				Detail:   err.Error(),
			},
		}
	} else if !featureApplicationHelmSkipCrdsSupported {
		_, skipCrdsOk := d.GetOk("spec.0.source.0.helm.0.skip_crds")
		if skipCrdsOk {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary: fmt.Sprintf(
						"application helm skip_crds is only supported from ArgoCD %s onwards",
						featureVersionConstraintsMap[featureApplicationHelmSkipCrds].String()),
				},
			}
		}
	}

	apps, err := si.ApplicationClient.List(ctx, appQuery)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "failed to get application",
				Detail:   err.Error(),
			},
		}
	}

	// Kubernetes API requires providing the up-to-date correct ResourceVersion for updates
	// FIXME ResourceVersion not available anymore
	// if app != nil {
	// 	 appRequest.ResourceVersion = app.ResourceVersion
	// }

	if len(apps.Items) > 1 {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("found multiple applications matching name '%s' and namespace '%s'", *appQuery.Name, *appQuery.AppNamespace),
				Detail:   err.Error(),
			},
		}
	}

	if _, err = si.ApplicationClient.Update(ctx, &applicationClient.ApplicationUpdateRequest{
		Application: &application.Application{
			ObjectMeta: objectMeta,
			Spec:       spec,
			TypeMeta: metav1.TypeMeta{
				Kind:       "Application",
				APIVersion: "argoproj.io/v1alpha1",
			},
		},
	}); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("application %s could not be updated", *appQuery.Name),
				Detail:   err.Error(),
			},
		}
	}

	if wait, _ok := d.GetOk("wait"); _ok && wait.(bool) {
		if err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			var list *application.ApplicationList
			if list, err = si.ApplicationClient.List(ctx, appQuery); err != nil {
				return resource.NonRetryableError(fmt.Errorf("error while waiting for application %s to be synced and healthy: %s", list.Items[0].Name, err))
			}

			if len(list.Items) != 1 {
				return resource.NonRetryableError(fmt.Errorf("found unexpected number of applications matching name '%s' and namespace '%s'. Items: %d", *appQuery.Name, *appQuery.AppNamespace, len(list.Items)))
			}

			if list.Items[0].Status.ReconciledAt.Equal(apps.Items[0].Status.ReconciledAt) {
				return resource.RetryableError(fmt.Errorf("reconciliation has not begun"))
			}

			if list.Items[0].Status.Health.Status != health.HealthStatusHealthy {
				return resource.RetryableError(fmt.Errorf("expected application health status to be healthy but was %s", list.Items[0].Status.Health.Status))
			}

			if list.Items[0].Status.Sync.Status != application.SyncStatusCodeSynced {
				return resource.RetryableError(fmt.Errorf("expected application sync status to be synced but was %s", list.Items[0].Status.Sync.Status))
			}

			return nil
		}); err != nil {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("something went wrong upon waiting for the application to be updated: %s", err),
					Detail:   err.Error(),
				},
			}
		}
	}

	return resourceArgoCDApplicationRead(ctx, d, meta)
}

func resourceArgoCDApplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	ids := strings.Split(d.Id(), ":")
	appName := ids[0]
	namespace := ids[1]
	cascade := d.Get("cascade").(bool)

	if _, err := si.ApplicationClient.Delete(ctx, &applicationClient.ApplicationDeleteRequest{
		Name:         &appName,
		Cascade:      &cascade,
		AppNamespace: &namespace,
	}); err != nil && !strings.Contains(err.Error(), "NotFound") {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("application %s could not be deleted", appName),
				Detail:   err.Error(),
			},
		}
	}

	if wait, ok := d.GetOk("wait"); ok && wait.(bool) {
		if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
			apps, err := si.ApplicationClient.List(ctx, &applicationClient.ApplicationQuery{
				Name:         &appName,
				AppNamespace: &namespace,
			})

			switch err {
			case nil:
				if apps != nil && len(apps.Items) > 0 {
					return resource.RetryableError(fmt.Errorf("application %s is still present", appName))
				}
			default:
				if !strings.Contains(err.Error(), "NotFound") {
					return resource.NonRetryableError(err)
				}
			}

			d.SetId("")

			return nil
		}); err != nil {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("application %s not be deleted", appName),
					Detail:   err.Error(),
				},
			}
		}
	}

	d.SetId("")

	return nil
}

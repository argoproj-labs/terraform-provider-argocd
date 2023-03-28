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
			"spec":     applicationSpecSchemaV4(),
			"wait": {
				Type:        schema.TypeBool,
				Description: "Upon application creation or update, wait for application health/sync status to be healthy/Synced, upon application deletion, wait for application to be removed, when set to true. Wait timeouts are controlled by Terraform Create, Update and Delete resource timeouts (all default to 5 minutes).",
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
	objectMeta, spec, diags := expandApplication(d)
	if diags != nil {
		return diags
	}
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
	c := *server.ApplicationClient
	apps, err := c.List(ctx, &applicationClient.ApplicationQuery{
		Name:         &objectMeta.Name,
		AppNamespace: &objectMeta.Namespace,
	})
	if err != nil && !strings.Contains(err.Error(), "NotFound") {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("application %s could not be created", objectMeta.Name),
				Detail:   err.Error(),
			},
		}
	}
	if apps != nil {
		if len(apps.Items) != 1 {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("found multiple applications matching name '%s' and namespace '%s'", objectMeta.Name, objectMeta.Namespace),
					Detail:   err.Error(),
				},
			}
		}

		switch apps.Items[0].DeletionTimestamp {
		case nil:
		default:
			// Pre-existing app is still in Kubernetes soft deletion queue
			time.Sleep(time.Duration(*apps.Items[0].DeletionGracePeriodSeconds))
		}
	}

	featureApplicationLevelSyncOptionsSupported, err := server.isFeatureSupported(featureApplicationLevelSyncOptions)
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

	featureIgnoreDiffJQPathExpressionsSupported, err := server.isFeatureSupported(featureIgnoreDiffJQPathExpressions)
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

	featureApplicationHelmSkipCrdsSupported, err := server.isFeatureSupported(featureApplicationHelmSkipCrds)
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

	app, err := c.Create(ctx, &applicationClient.ApplicationCreateRequest{
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
	}
	if app == nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("application %s could not be created: unknown reason", objectMeta.Name),
			},
		}
	}

	d.SetId(fmt.Sprintf("%s:%s", app.Name, objectMeta.Namespace))

	if wait, ok := d.GetOk("wait"); ok && wait.(bool) {
		err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
			list, err := c.List(ctx, &applicationClient.ApplicationQuery{
				Name:         &app.Name,
				AppNamespace: &app.Namespace,
			})
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf("error while waiting for application %s to be synced and healthy: %s", app.Name, err))
			}
			if len(list.Items) != 1 {
				return resource.NonRetryableError(fmt.Errorf("found multiple applications matching name '%s' and namespace '%s'", app.Name, app.Namespace))
			}
			if list.Items[0].Status.Health.Status != health.HealthStatusHealthy {
				return resource.RetryableError(fmt.Errorf("expected application health status to be healthy but was %s", list.Items[0].Status.Health.Status))
			}
			if list.Items[0].Status.Sync.Status != application.SyncStatusCodeSynced {
				return resource.RetryableError(fmt.Errorf("expected application sync status to be synced but was %s", list.Items[0].Status.Sync.Status))
			}
			return nil
		})
		if err != nil {
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

	c := *server.ApplicationClient

	ids := strings.Split(d.Id(), ":")
	appName := ids[0]
	namespace := ids[1]

	apps, err := c.List(ctx, &applicationClient.ApplicationQuery{
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
				Summary:  fmt.Sprintf("application %s not found", appName),
				Detail:   err.Error(),
			},
		}
	}
	if len(apps.Items) != 1 {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("found multiple applications matching name '%s' and namespace '%s'", appName, namespace),
				Detail:   err.Error(),
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
	ids := strings.Split(d.Id(), ":")
	appName := ids[0]
	namespace := ids[1]
	if ok := d.HasChanges("metadata", "spec"); ok {
		objectMeta, spec, diags := expandApplication(d)
		if diags != nil {
			return diags
		}
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
		c := *server.ApplicationClient
		appRequest := &applicationClient.ApplicationUpdateRequest{
			Application: &application.Application{
				ObjectMeta: objectMeta,
				Spec:       spec,
				TypeMeta: metav1.TypeMeta{
					Kind:       "Application",
					APIVersion: "argoproj.io/v1alpha1",
				},
			},
		}

		featureApplicationLevelSyncOptionsSupported, err := server.isFeatureSupported(featureApplicationLevelSyncOptions)
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

		featureIgnoreDiffJQPathExpressionsSupported, err := server.isFeatureSupported(featureIgnoreDiffJQPathExpressions)
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

		featureApplicationHelmSkipCrdsSupported, err := server.isFeatureSupported(featureApplicationHelmSkipCrds)
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
		apps, err := c.List(ctx, &applicationClient.ApplicationQuery{
			Name:         &appName,
			AppNamespace: &namespace,
		})
		if apps != nil {
			// Kubernetes API requires providing the up-to-date correct ResourceVersion for updates
			// FIXME ResourceVersion not available anymore
			// appRequest.ResourceVersion = app.ResourceVersion
		}
		_, err = c.Update(ctx, appRequest)
		if err != nil {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("application %s could not be updated", appName),
					Detail:   err.Error(),
				},
			}
		}
		if wait, _ok := d.GetOk("wait"); _ok && wait.(bool) {
			err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
				list, err := c.List(ctx, &applicationClient.ApplicationQuery{
					Name:         &appName,
					AppNamespace: &namespace,
				})
				if err != nil {
					return resource.NonRetryableError(fmt.Errorf("error while waiting for application %s to be synced and healthy: %s", list.Items[0].Name, err))
				}
				if len(list.Items) != 1 {
					return resource.NonRetryableError(fmt.Errorf("found multiple applications matching name '%s' and namespace '%s'", appName, namespace))
				}
				if list.Items[0].Status.Health.Status != health.HealthStatusHealthy {
					return resource.RetryableError(fmt.Errorf("expected application health status to be healthy but was %s", list.Items[0].Status.Health.Status))
				}
				if list.Items[0].Status.Sync.Status != application.SyncStatusCodeSynced {
					return resource.RetryableError(fmt.Errorf("expected application sync status to be synced but was %s", list.Items[0].Status.Sync.Status))
				}
				return nil
			})
			if err != nil {
				return []diag.Diagnostic{
					{
						Severity: diag.Error,
						Summary:  fmt.Sprintf("something went wrong upon waiting for the application to be updated: %s", err),
						Detail:   err.Error(),
					},
				}
			}
		}
	}
	return resourceArgoCDApplicationRead(ctx, d, meta)
}

func resourceArgoCDApplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	c := *server.ApplicationClient
	ids := strings.Split(d.Id(), ":")
	appName := ids[0]
	namespace := ids[1]
	cascade := d.Get("cascade").(bool)
	_, err := c.Delete(ctx, &applicationClient.ApplicationDeleteRequest{
		Name:         &appName,
		Cascade:      &cascade,
		AppNamespace: &namespace,
	})
	if err != nil && !strings.Contains(err.Error(), "NotFound") {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("application %s could not be deleted", appName),
				Detail:   err.Error(),
			},
		}
	}
	if wait, ok := d.GetOk("wait"); ok && wait.(bool) {
		err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
			_, err := c.List(ctx, &applicationClient.ApplicationQuery{
				Name:         &appName,
				AppNamespace: &namespace,
			})
			if err == nil {
				return resource.RetryableError(fmt.Errorf("application %s is still present", appName))
			}
			if !strings.Contains(err.Error(), "NotFound") {
				return resource.NonRetryableError(err)
			}
			d.SetId("")
			return nil
		})
		if err != nil {
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

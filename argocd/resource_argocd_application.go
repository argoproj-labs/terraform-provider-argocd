package argocd

import (
	"context"
	"fmt"
	"strings"
	"time"

	applicationClient "github.com/argoproj/argo-cd/pkg/apiclient/application"
	application "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/gitops-engine/pkg/health"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceArgoCDApplication() *schema.Resource {
	return &schema.Resource{
		Create: resourceArgoCDApplicationCreate,
		Read:   resourceArgoCDApplicationRead,
		Update: resourceArgoCDApplicationUpdate,
		Delete: resourceArgoCDApplicationDelete,
		// TODO: add importer acceptance tests
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"metadata": metadataSchema("applications.argoproj.io"),
			"spec":     applicationSpecSchema(),
			"wait": {
				Type:        schema.TypeBool,
				Description: "Upon application creation or update, wait for application health/sync status to be healthy/Synced, upon application deletion, wait for application to be removed, when set to true.",
				Optional:    true,
				Default:     false,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func resourceArgoCDApplicationCreate(d *schema.ResourceData, meta interface{}) error {
	objectMeta, spec, err := expandApplication(d)
	if err != nil {
		return err
	}
	server := meta.(ServerInterface)
	c := *server.ApplicationClient
	app, err := c.Get(context.Background(), &applicationClient.ApplicationQuery{
		Name: &objectMeta.Name,
	})
	if err != nil {
		switch strings.Contains(err.Error(), "NotFound") {
		case true:
		default:
			return err
		}
	}
	if app != nil {
		switch app.DeletionTimestamp {
		case nil:
		default:
			// Pre-existing app is still in Kubernetes soft deletion queue
			time.Sleep(time.Duration(*app.DeletionGracePeriodSeconds))
		}
	}

	featureApplicationLevelSyncOptionsSupported, err := server.isFeatureSupported(featureApplicationLevelSyncOptions)
	if err != nil {
		return err
	}
	if !featureApplicationLevelSyncOptionsSupported &&
		spec.SyncPolicy != nil &&
		spec.SyncPolicy.SyncOptions != nil {
		return fmt.Errorf(
			"application-level sync_options is only supported from ArgoCD %s onwards",
			featureVersionConstraintsMap[featureApplicationLevelSyncOptions].String())
	}

	app, err = c.Create(context.Background(), &applicationClient.ApplicationCreateRequest{
		Application: application.Application{
			ObjectMeta: objectMeta,
			Spec:       spec,
		},
	})
	if err != nil {
		return err
	}
	if app == nil {
		return fmt.Errorf("something went wrong during application creation")
	}
	d.SetId(app.Name)
	if wait, ok := d.GetOk("wait"); ok && wait.(bool) {
		err = resource.RetryContext(context.Background(), d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
			a, err := c.Get(context.Background(), &applicationClient.ApplicationQuery{
				Name: &app.Name,
			})
			if err != nil {
				if strings.Contains(err.Error(), "NotFound") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(fmt.Errorf("error while waiting for application %s to be synced and healthy: %s", app.Name, err))
			}
			if a.Status.Health.Status != health.HealthStatusHealthy {
				return resource.RetryableError(fmt.Errorf("expected application health status to be healthy but was %s", a.Status.Health.Status))
			}
			if a.Status.Sync.Status != application.SyncStatusCodeSynced {
				return resource.RetryableError(fmt.Errorf("expected application sync status to be synced but was %s", a.Status.Sync.Status))
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("something went wrong upon waiting for the application to be created: %s", err)
		}
	}
	return resourceArgoCDApplicationRead(d, meta)
}

func resourceArgoCDApplicationRead(d *schema.ResourceData, meta interface{}) error {
	server := meta.(ServerInterface)
	c := *server.ApplicationClient
	appName := d.Id()
	app, err := c.Get(context.Background(), &applicationClient.ApplicationQuery{
		Name: &appName,
	})
	if err != nil {
		switch strings.Contains(err.Error(), "NotFound") {
		case true:
			d.SetId("")
			return nil
		default:
			return err
		}
	}
	err = flattenApplication(app, d)
	return err
}

func resourceArgoCDApplicationUpdate(d *schema.ResourceData, meta interface{}) error {
	appName := d.Id()

	if ok := d.HasChanges("metadata", "spec"); ok {
		objectMeta, spec, err := expandApplication(d)
		if err != nil {
			return err
		}
		server := meta.(ServerInterface)
		c := *server.ApplicationClient
		appRequest := &applicationClient.ApplicationUpdateRequest{
			Application: &application.Application{
				ObjectMeta: objectMeta,
				Spec:       spec,
			}}

		featureApplicationLevelSyncOptionsSupported, err := server.isFeatureSupported(featureApplicationLevelSyncOptions)
		if err != nil {
			return err
		}
		if !featureApplicationLevelSyncOptionsSupported &&
			spec.SyncPolicy != nil &&
			spec.SyncPolicy.SyncOptions != nil {
			return fmt.Errorf(
				"application-level sync_options is only supported from ArgoCD %s onwards",
				featureVersionConstraintsMap[featureApplicationLevelSyncOptions].String())
		}

		app, err := c.Get(context.Background(), &applicationClient.ApplicationQuery{
			Name: &appName,
		})
		if app != nil {
			// Kubernetes API requires providing the up-to-date correct ResourceVersion for updates
			appRequest.Application.ResourceVersion = app.ResourceVersion
		}
		_, err = c.Update(context.Background(), appRequest)
		if err != nil {
			return err
		}
		if wait, _ok := d.GetOk("wait"); _ok && wait.(bool) {
			err = resource.RetryContext(context.Background(), d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
				a, err := c.Get(context.Background(), &applicationClient.ApplicationQuery{
					Name: &app.Name,
				})
				if err != nil {
					return resource.NonRetryableError(fmt.Errorf("error while waiting for application %s to be synced and healthy: %s", app.Name, err))
				}
				if a.Status.Health.Status != health.HealthStatusHealthy {
					return resource.RetryableError(fmt.Errorf("expected application health status to be healthy but was %s", a.Status.Health.Status))
				}
				if a.Status.Sync.Status != application.SyncStatusCodeSynced {
					return resource.RetryableError(fmt.Errorf("expected application sync status to be synced but was %s", a.Status.Sync.Status))
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("something went wrong upon waiting for the application to be updated: %s", err)
			}
		}
	}
	return resourceArgoCDApplicationRead(d, meta)
}

func resourceArgoCDApplicationDelete(d *schema.ResourceData, meta interface{}) error {
	server := meta.(ServerInterface)
	c := *server.ApplicationClient
	appName := d.Id()
	_, err := c.Delete(context.Background(), &applicationClient.ApplicationDeleteRequest{Name: &appName})
	if err != nil && !strings.Contains(err.Error(), "NotFound") {
		return err
	}
	if wait, ok := d.GetOk("wait"); ok && wait.(bool) {
		return resource.RetryContext(context.Background(), d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
			_, err = c.Get(context.Background(), &applicationClient.ApplicationQuery{
				Name: &appName,
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
	}
	d.SetId("")
	return nil
}

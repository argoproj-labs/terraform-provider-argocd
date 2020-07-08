package argocd

import (
	"context"
	"fmt"
	applicationClient "github.com/argoproj/argo-cd/pkg/apiclient/application"
	application "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"strings"
	"time"
)

func resourceArgoCDApplication() *schema.Resource {
	return &schema.Resource{
		Create: resourceArgoCDApplicationCreate,
		Read:   resourceArgoCDApplicationRead,
		Update: resourceArgoCDApplicationUpdate,
		Delete: resourceArgoCDApplicationDelete,
		// TODO: add importer acceptance tests
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"metadata": metadataSchema("applications.argoproj.io"),
			"spec":     applicationSpecSchema(),
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
	return resourceArgoCDApplicationRead(d, meta)
}

func resourceArgoCDApplicationRead(d *schema.ResourceData, meta interface{}) error {
	server := meta.(ServerInterface)
	c := *server.ApplicationClient
	appName := d.Id()
	app, err := c.Get(context.Background(), &applicationClient.ApplicationQuery{
		Name: &appName,
	},
	)
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
	}
	return resourceArgoCDApplicationRead(d, meta)
}

func resourceArgoCDApplicationDelete(d *schema.ResourceData, meta interface{}) error {
	server := meta.(ServerInterface)
	c := *server.ApplicationClient
	appName := d.Id()
	_, err := c.Delete(context.Background(), &applicationClient.ApplicationDeleteRequest{Name: &appName})
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

package argocd

import (
	"context"
	"fmt"
	clusterClient "github.com/argoproj/argo-cd/pkg/apiclient/cluster"
	application "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceArgoCDCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceArgoCDClusterCreate,
		Read:   resourceArgoCDClusterRead,
		Update: resourceArgoCDClusterUpdate,
		Delete: resourceArgoCDClusterDelete,
		// TODO: add importer acceptance tests
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: clusterSchema(),
	}
}

func resourceArgoCDClusterCreate(d *schema.ResourceData, meta interface{}) error {
	server := meta.(ServerInterface)
	client := *server.ClusterClient
	cluster, err := expandCluster(d)
	if err != nil {
		return fmt.Errorf("could not expand cluster attributes: %s", err)
	}

	c, err := client.Create(context.Background(), &clusterClient.ClusterCreateRequest{
		Cluster: cluster})
	if err != nil {
		return fmt.Errorf("something went wrong during cluster resource creation")
	}

	d.SetId(c.ID)
	return resourceArgoCDClusterRead(d, meta)
}

func resourceArgoCDClusterRead(d *schema.ResourceData, meta interface{}) error {
	server := meta.(ServerInterface)
	client := *server.ClusterClient

	err := flattenCluster(c, d)
	return err
}

func resourceArgoCDClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	server := meta.(ServerInterface)
	client := *server.ClusterClient

	return resourceArgoCDClusterRead(d, meta)
}

func resourceArgoCDClusterDelete(d *schema.ResourceData, meta interface{}) error {
	server := meta.(ServerInterface)
	client := *server.ClusterClient

	d.SetId("")
	return nil
}

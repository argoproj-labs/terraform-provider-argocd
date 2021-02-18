package argocd

import (
	"context"
	"fmt"
	clusterClient "github.com/argoproj/argo-cd/pkg/apiclient/cluster"
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
		Cluster: cluster, Upsert: false})
	if err != nil {
		return fmt.Errorf("something went wrong during cluster resource creation")
	}
	d.SetId(c.Server)
	return resourceArgoCDClusterRead(d, meta)
}

func resourceArgoCDClusterRead(d *schema.ResourceData, meta interface{}) error {
	server := meta.(ServerInterface)
	client := *server.ClusterClient
	clusterQuery := &clusterClient.ClusterQuery{Server: d.Id()}
	c, err := client.Get(context.Background(), clusterQuery)
	if err != nil {
		return fmt.Errorf("could not get cluster information: %s", err)
	}
	err = flattenCluster(c, d)
	return err
}

func resourceArgoCDClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	server := meta.(ServerInterface)
	client := *server.ClusterClient
	cluster, err := expandCluster(d)
	if err != nil {
		return fmt.Errorf("could not expand cluster attributes: %s", err)
	}
	_, err = client.Update(context.Background(), &clusterClient.ClusterUpdateRequest{Cluster: cluster})
	if err != nil {
		return fmt.Errorf("something went wrong during cluster update: %s", err)
	}
	return resourceArgoCDClusterRead(d, meta)
}

func resourceArgoCDClusterDelete(d *schema.ResourceData, meta interface{}) error {
	server := meta.(ServerInterface)
	client := *server.ClusterClient
	_, err := client.Delete(context.Background(), &clusterClient.ClusterQuery{Server: d.Id()})
	if err != nil {
		return fmt.Errorf("something went wrong during cluster deletion: %s", err)
	}
	d.SetId("")
	return nil
}

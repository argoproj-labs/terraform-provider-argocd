package argocd

import (
	"context"
	"fmt"
	"strings"

	"github.com/argoproj-labs/terraform-provider-argocd/internal/provider"
	clusterClient "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceArgoCDCluster() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages [clusters](https://argo-cd.readthedocs.io/en/stable/operator-manual/declarative-setup/#clusters) within ArgoCD.",
		CreateContext: resourceArgoCDClusterCreate,
		ReadContext:   resourceArgoCDClusterRead,
		UpdateContext: resourceArgoCDClusterUpdate,
		DeleteContext: resourceArgoCDClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: clusterSchema(),
	}
}

func resourceArgoCDClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*provider.ServerInterface)
	if diags := si.InitClients(ctx); diags != nil {
		return pluginSDKDiags(diags)
	}

	cluster, err := expandCluster(d)
	if err != nil {
		return errorToDiagnostics("failed to expand cluster", err)
	}

	// Need a full lock here to avoid race conditions between List existing clusters and creating a new one
	tokenMutexClusters.Lock()

	rtrimmedServer := strings.TrimRight(cluster.Server, "/")

	// Cluster are unique by "server address" so we should check there is no existing cluster with this address before
	existingClusters, err := si.ClusterClient.List(ctx, &clusterClient.ClusterQuery{
		// Starting argo-cd server v2.8.0 filtering on list api endpoint is fixed, else it is ignored, see:
		// - https://github.com/oboukili/terraform-provider-argocd/issues/266#issuecomment-1739122022
		// - https://github.com/argoproj/argo-cd/pull/13363
		Id: &clusterClient.ClusterID{
			Type:  "server",
			Value: rtrimmedServer,
		},
	})
	if err != nil {
		tokenMutexClusters.Unlock()
		return errorToDiagnostics(fmt.Sprintf("failed to list existing clusters when creating cluster %s", cluster.Server), err)
	}

	// Here we will filter ourselves on the list so that we are backward compatible for argo-cd server with version < v2.8.0 (see coment above)
	if len(existingClusters.Items) > 0 {
		for _, existingCluster := range existingClusters.Items {
			if rtrimmedServer == strings.TrimRight(existingCluster.Server, "/") {
				tokenMutexClusters.Unlock()

				return []diag.Diagnostic{
					{
						Severity: diag.Error,
						Summary:  fmt.Sprintf("cluster with server address %s already exists", cluster.Server),
					},
				}
			}
		}
	}

	c, err := si.ClusterClient.Create(ctx, &clusterClient.ClusterCreateRequest{
		Cluster: cluster, Upsert: false,
	})
	tokenMutexClusters.Unlock()

	if err != nil {
		return argoCDAPIError("create", "cluster", cluster.Server, err)
	}

	// Check if the name has been defaulted to server (when omitted)
	if c.Name != "" && c.Name != c.Server {
		d.SetId(fmt.Sprintf("%s/%s", c.Server, c.Name))
	} else {
		d.SetId(c.Server)
	}

	return resourceArgoCDClusterRead(ctx, d, meta)
}

func resourceArgoCDClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*provider.ServerInterface)
	if diags := si.InitClients(ctx); diags != nil {
		return pluginSDKDiags(diags)
	}

	tokenMutexClusters.RLock()
	c, err := si.ClusterClient.Get(ctx, getClusterQueryFromID(d))
	tokenMutexClusters.RUnlock()

	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			d.SetId("")
			return nil
		}

		// Fix for https://github.com/oboukili/terraform-provider-argocd/issues/266
		// This fix is added here as a workaround to ensure backward compatibility, as
		//   it is triggered only on the specific usecase where the issue happens.
		// Additional remarks about this code:
		// * it is a copy/paste of the code used by resourceArgoCDClusterCreate to check if
		//     the cluster already exists (with some obvious changes to return value and mutex type)
		// * it should at term replace the `si.ClusterClient.Get` code for this method
		if strings.Contains(err.Error(), "PermissionDenied") {
			cluster, err := expandCluster(d)
			if err != nil {
				return errorToDiagnostics("failed to expand cluster", err)
			}

			tokenMutexClusters.RLock()

			rtrimmedServer := strings.TrimRight(cluster.Server, "/")

			// Cluster are unique by "server address" so we should check there is no existing cluster with this address before
			existingClusters, err := si.ClusterClient.List(ctx, &clusterClient.ClusterQuery{
				// Starting argo-cd server v2.8.0 filtering on list api endpoint is fixed, else it is ignored, see:
				// - https://github.com/oboukili/terraform-provider-argocd/issues/266#issuecomment-1739122022
				// - https://github.com/argoproj/argo-cd/pull/13363
				Id: &clusterClient.ClusterID{
					Type:  "server",
					Value: rtrimmedServer,
				},
			})

			tokenMutexClusters.RUnlock()

			if err != nil {
				return errorToDiagnostics(fmt.Sprintf("failed to list existing clusters when reading cluster %s", cluster.Server), err)
			}

			// Here we will filter ourselves on the list so that we are backward compatible for argo-cd server with version < v2.8.0 (see coment above)
			if len(existingClusters.Items) > 0 {
				for _, existingCluster := range existingClusters.Items {
					if rtrimmedServer == strings.TrimRight(existingCluster.Server, "/") {
						// Cluster was found, return
						return nil
					}
				}
			}

			// Cluster was not found, return with empty Id
			d.SetId("")

			return nil
		} else {
			return argoCDAPIError("read", "cluster", d.Id(), err)
		}
	}

	if err = flattenCluster(c, d); err != nil {
		return errorToDiagnostics(fmt.Sprintf("failed to flatten cluster %s", d.Id()), err)
	}

	return nil
}

func resourceArgoCDClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*provider.ServerInterface)
	if diags := si.InitClients(ctx); diags != nil {
		return pluginSDKDiags(diags)
	}

	cluster, err := expandCluster(d)
	if err != nil {
		return errorToDiagnostics(fmt.Sprintf("failed to expand cluster %s", d.Id()), err)
	}

	tokenMutexClusters.Lock()
	_, err = si.ClusterClient.Update(ctx, &clusterClient.ClusterUpdateRequest{Cluster: cluster})
	tokenMutexClusters.Unlock()

	if err != nil {
		return argoCDAPIError("update", "cluster", cluster.Server, err)
	}

	return resourceArgoCDClusterRead(ctx, d, meta)
}

func resourceArgoCDClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*provider.ServerInterface)
	if diags := si.InitClients(ctx); diags != nil {
		return pluginSDKDiags(diags)
	}

	tokenMutexClusters.Lock()
	_, err := si.ClusterClient.Delete(ctx, getClusterQueryFromID(d))
	tokenMutexClusters.Unlock()

	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			d.SetId("")
			return nil
		}

		return argoCDAPIError("delete", "cluster", d.Id(), err)
	}

	d.SetId("")

	return nil
}

func getClusterQueryFromID(d *schema.ResourceData) *clusterClient.ClusterQuery {
	cq := &clusterClient.ClusterQuery{}

	id := strings.Split(strings.TrimPrefix(d.Id(), "https://"), "/")
	if len(id) > 1 {
		cq.Name = id[len(id)-1]
		cq.Server = fmt.Sprintf("https://%s", strings.Join(id[:len(id)-1], "/"))
	} else {
		cq.Server = d.Id()
	}

	return cq
}

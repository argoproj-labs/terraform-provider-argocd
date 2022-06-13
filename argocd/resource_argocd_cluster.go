package argocd

import (
	"context"
	"fmt"
	"strings"

	clusterClient "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceArgoCDCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceArgoCDClusterCreate,
		ReadContext:   resourceArgoCDClusterRead,
		UpdateContext: resourceArgoCDClusterUpdate,
		DeleteContext: resourceArgoCDClusterDelete,
		// TODO: add importer tests
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: clusterSchema(),
	}
}

func resourceArgoCDClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	client := *server.ClusterClient
	cluster, err := expandCluster(d)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("could not expand cluster attributes: %s", err),
				Detail:   err.Error(),
			},
		}

	}

	featureProjectScopedClustersSupported, err := server.isFeatureSupported(featureProjectScopedClusters)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "feature not supported",
				Detail:   err.Error(),
			},
		}
	}
	if !featureProjectScopedClustersSupported && cluster.Project != "" {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"cluster project is only supported from ArgoCD %s onwards",
					featureVersionConstraintsMap[featureProjectScopedClusters].String()),
				Detail: "See https://argo-cd.readthedocs.io/en/stable/user-guide/projects/#project-scoped-repositories-and-clusters",
			},
		}
	}

	featureClusterMetadataSupported, err := server.isFeatureSupported(featureClusterMetadata)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "feature not supported",
				Detail:   err.Error(),
			},
		}
	}

	if !featureClusterMetadataSupported && (len(cluster.Annotations) != 0 || len(cluster.Labels) != 0) {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"cluster metadata is only supported from ArgoCD %s onwards",
					featureVersionConstraintsMap[featureClusterMetadata].String()),
			},
		}
	}

	// Cluster are unique by "server address" so we should check there is no existing cluster with this address before
	tokenMutexClusters.RLock()
	existingClusters, err := client.List(ctx, &clusterClient.ClusterQuery{
		Id: &clusterClient.ClusterID{
			Type:  "server",
			Value: cluster.Server, // TODO: not used by backend, upstream bug ?
		},
	})
	tokenMutexClusters.RUnlock()
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("could not get current clusters list:  %s", err),
				Detail:   err.Error(),
			},
		}
	}

	rtrimmedServer := strings.TrimRight(cluster.Server, "/")
	if err == nil && len(existingClusters.Items) > 0 {
		for _, existingCluster := range existingClusters.Items {
			if rtrimmedServer == strings.TrimRight(existingCluster.Server, "/") {
				return []diag.Diagnostic{
					{
						Severity: diag.Error,
						Summary:  fmt.Sprintf("cluster with server address %s already exists", cluster.Server),
					},
				}
			}
		}
	}

	tokenMutexClusters.Lock()
	c, err := client.Create(ctx, &clusterClient.ClusterCreateRequest{
		Cluster: cluster, Upsert: true})
	tokenMutexClusters.Unlock()

	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("something went wrong during cluster resource creation: %s", err),
				Detail:   err.Error(),
			},
		}
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
	client := *server.ClusterClient

	tokenMutexClusters.RLock()
	c, err := client.Get(ctx, getClusterQueryFromID(d))
	tokenMutexClusters.RUnlock()

	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			d.SetId("")
			return nil
		} else {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("could not get cluster information: %s", err),
					Detail:   err.Error(),
				},
			}
		}
	}
	err = flattenCluster(c, d)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "could not flatten cluster",
				Detail:   err.Error(),
			},
		}
	}
	return nil
}

func resourceArgoCDClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	client := *server.ClusterClient
	cluster, err := expandCluster(d)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("could not expand cluster attributes: %s", err),
				Detail:   err.Error(),
			},
		}
	}

	featureProjectScopedClustersSupported, err := server.isFeatureSupported(featureProjectScopedClusters)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "feature not supported",
				Detail:   err.Error(),
			},
		}
	}
	if !featureProjectScopedClustersSupported && cluster.Project != "" {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"cluster project is only supported from ArgoCD %s onwards",
					featureVersionConstraintsMap[featureProjectScopedClusters].String()),
				Detail: "See https://argo-cd.readthedocs.io/en/stable/user-guide/projects/#project-scoped-repositories-and-clusters",
			},
		}
	}

	featureClusterMetadataSupported, err := server.isFeatureSupported(featureClusterMetadata)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "feature not supported",
				Detail:   err.Error(),
			},
		}
	}

	if !featureClusterMetadataSupported && (len(cluster.Annotations) != 0 || len(cluster.Labels) != 0) {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"cluster metadata is only supported from ArgoCD %s onwards",
					featureVersionConstraintsMap[featureClusterMetadata].String()),
			},
		}
	}

	tokenMutexClusters.Lock()
	_, err = client.Update(ctx, &clusterClient.ClusterUpdateRequest{Cluster: cluster})
	tokenMutexClusters.Unlock()

	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			d.SetId("")
			return nil
		} else {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("something went wrong during cluster update: %s", err),
					Detail:   err.Error(),
				},
			}
		}
	}
	return resourceArgoCDClusterRead(ctx, d, meta)
}

func resourceArgoCDClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	client := *server.ClusterClient
	tokenMutexClusters.Lock()
	_, err := client.Delete(ctx, getClusterQueryFromID(d))
	tokenMutexClusters.Unlock()

	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			d.SetId("")
			return nil
		} else {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("something went wrong during cluster deletion: %s", err),
					Detail:   err.Error(),
				},
			}
		}
	}
	d.SetId("")
	return nil
}

func getClusterQueryFromID(d *schema.ResourceData) *clusterClient.ClusterQuery {
	id := strings.Split(strings.TrimPrefix(d.Id(), "https://"), "/")
	cq := &clusterClient.ClusterQuery{}
	if len(id) > 1 {
		cq.Name = id[len(id)-1]
		cq.Server = fmt.Sprintf("https://%s", strings.Join(id[:len(id)-1], "/"))
	} else {
		cq.Server = d.Id()
	}
	return cq
}

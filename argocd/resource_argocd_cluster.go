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

	featureProjectScopedClustersSupported, err := si.isFeatureSupported(featureProjectScopedClusters)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "feature not supported",
				Detail:   err.Error(),
			},
		}
	} else if !featureProjectScopedClustersSupported && cluster.Project != "" {
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

	featureClusterMetadataSupported, err := si.isFeatureSupported(featureClusterMetadata)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "feature not supported",
				Detail:   err.Error(),
			},
		}
	} else if !featureClusterMetadataSupported && (len(cluster.Annotations) != 0 || len(cluster.Labels) != 0) {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"cluster metadata is only supported from ArgoCD %s onwards",
					featureVersionConstraintsMap[featureClusterMetadata].String()),
			},
		}
	}

	// Need a full lock here to avoid race conditions between List existing clusters and creating a new one
	tokenMutexClusters.Lock()

	// Cluster are unique by "server address" so we should check there is no existing cluster with this address before
	existingClusters, err := si.ClusterClient.List(ctx, &clusterClient.ClusterQuery{
		Id: &clusterClient.ClusterID{
			Type:  "server",
			Value: cluster.Server, // TODO: not used by backend, upstream bug ?
		},
	})

	if err != nil {
		tokenMutexClusters.Unlock()

		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("could not get current clusters list:  %s", err),
				Detail:   err.Error(),
			},
		}
	}

	rtrimmedServer := strings.TrimRight(cluster.Server, "/")

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
		Cluster: cluster, Upsert: false})
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

	tokenMutexClusters.RLock()
	c, err := si.ClusterClient.Get(ctx, getClusterQueryFromID(d))
	tokenMutexClusters.RUnlock()

	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			d.SetId("")
			return nil
		}

		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("could not get cluster information: %s", err),
				Detail:   err.Error(),
			},
		}
	}

	if err = flattenCluster(c, d); err != nil {
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

	featureProjectScopedClustersSupported, err := si.isFeatureSupported(featureProjectScopedClusters)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "feature not supported",
				Detail:   err.Error(),
			},
		}
	} else if !featureProjectScopedClustersSupported && cluster.Project != "" {
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

	featureClusterMetadataSupported, err := si.isFeatureSupported(featureClusterMetadata)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "feature not supported",
				Detail:   err.Error(),
			},
		}
	} else if !featureClusterMetadataSupported && (len(cluster.Annotations) != 0 || len(cluster.Labels) != 0) {
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
	_, err = si.ClusterClient.Update(ctx, &clusterClient.ClusterUpdateRequest{Cluster: cluster})
	tokenMutexClusters.Unlock()

	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("something went wrong during cluster update: %s", err),
				Detail:   err.Error(),
			},
		}
	}

	return resourceArgoCDClusterRead(ctx, d, meta)
}

func resourceArgoCDClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	tokenMutexClusters.Lock()
	_, err := si.ClusterClient.Delete(ctx, getClusterQueryFromID(d))
	tokenMutexClusters.Unlock()

	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			d.SetId("")
			return nil
		}

		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("something went wrong during cluster deletion: %s", err),
				Detail:   err.Error(),
			},
		}
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

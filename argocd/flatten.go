package argocd

import (
	argoCDAppv1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func flattenK8SGroupKinds(gks []metav1.GroupKind) (
	result []map[string]string) {
	for _, gk := range gks {
		result = append(result, map[string]string{
			"group": gk.Group,
			"kind":  gk.Kind,
		})
	}
	return
}

func flattenArgoCDDestinations(ds []argoCDAppv1.ApplicationDestination) (
	result []map[string]string) {
	for _, d := range ds {
		result = append(result, map[string]string{
			"server":    d.Server,
			"namespace": d.Namespace,
		})
	}
	return
}

func flattenArgoCDOrphanedResources(ors *argoCDAppv1.OrphanedResourcesMonitorSettings) (
	result map[string]bool) {
	if ors != nil {
		result = map[string]bool{
			"warn": *ors.Warn,
		}
	}
	return
}

func flattenArgoCDRoleJWTTokens(jwts []argoCDAppv1.JWTToken) (
	result []map[string]string) {
	for _, jwt := range jwts {
		result = append(result, map[string]string{
			"issued_at":  convertInt64ToString(jwt.IssuedAt),
			"expired_at": convertInt64ToString(jwt.ExpiresAt),
		})
	}
	return
}

func flattenArgoCDRoles(rs []argoCDAppv1.ProjectRole) (
	result []map[string]interface{}) {
	for _, r := range rs {
		result = append(result, map[string]interface{}{
			"name":        r.Name,
			"description": r.Description,
			"groups":      r.Groups,
			"jwt_tokens":  flattenArgoCDRoleJWTTokens(r.JWTTokens),
			"policies":    r.Policies,
		})
	}
	return
}

func flattenArgoCDSyncWindows(sws argoCDAppv1.SyncWindows) (
	result []map[string]interface{}) {
	for _, sw := range sws {
		result = append(result, map[string]interface{}{
			"applications": sw.Applications,
			"clusters":     sw.Clusters,
			"duration":     sw.Duration,
			"kind":         sw.Kind,
			"manual_sync":  sw.ManualSync,
			"namespaces":   sw.Namespaces,
		})
	}
	return
}

func flattenArgoCDProject(p *argoCDAppv1.AppProject) (
	map[string]interface{},
	error) {

	result := map[string]interface{}{
		"metadata": []map[string]interface{}{
			{
				"name":             p.Name,
				"namespace":        p.Namespace,
				"uid":              string(p.UID),
				"resource_version": p.ResourceVersion,
				"generation":       convertInt64ToString(p.Generation),
				// TODO: Time to string conversion
				"creation_timestamp": p.CreationTimestamp,
			},
		},
		"spec": []map[string]interface{}{
			{
				"cluster_resource_whitelist": flattenK8SGroupKinds(
					p.Spec.ClusterResourceWhitelist),
				"namespace_resource_blacklist": flattenK8SGroupKinds(
					p.Spec.NamespaceResourceBlacklist),

				"description":        p.Spec.Description,
				"destinations":       flattenArgoCDDestinations(p.Spec.Destinations),
				"orphaned_resources": flattenArgoCDOrphanedResources(p.Spec.OrphanedResources),
				"roles":              flattenArgoCDRoles(p.Spec.Roles),
				"source_repos":       p.Spec.SourceRepos,
				"sync_windows":       flattenArgoCDSyncWindows(p.Spec.SyncWindows),
			},
		},
	}
	return result, nil
}

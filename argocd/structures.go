package argocd

import (
	argoCDAppv1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func flattenDestinations(ds []argoCDAppv1.ApplicationDestination) (
	result []map[string]string) {
	for _, d := range ds {
		result = append(result, map[string]string{
			"server":    d.Server,
			"namespace": d.Namespace,
		})
	}
	return
}

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
func flattenOrphanedResources(ors *argoCDAppv1.OrphanedResourcesMonitorSettings) (
	result map[string]bool) {
	if ors != nil {
		result = map[string]bool{
			"warn": *ors.Warn,
		}
	}
	return
}

func flattenRoleJWTTokens(jwts []argoCDAppv1.JWTToken) (
	result []map[string]string) {
	for _, jwt := range jwts {
		result = append(result, map[string]string{
			"iat": convertInt64ToString(jwt.IssuedAt),
			"exp": convertInt64ToString(jwt.ExpiresAt),
		})
	}
	return
}

func flattenRoles(rs []argoCDAppv1.ProjectRole) (
	result []map[string]interface{}) {
	for _, r := range rs {
		result = append(result, map[string]interface{}{
			"name":        r.Name,
			"description": r.Description,
			"groups":      r.Groups,
			"jwt_token":   flattenRoleJWTTokens(r.JWTTokens),
			"policies":    r.Policies,
		})
	}
	return
}

func flattenSyncWindows(sws argoCDAppv1.SyncWindows) (
	result []map[string]interface{}) {
	for _, sw := range sws {
		result = append(result, map[string]interface{}{
			"applications": sw.Applications,
			"clusters":     sw.Clusters,
			"duration":     sw.Duration,
			"kind":         sw.Kind,
			"manual_sync":  sw.ManualSync,
			"namespaces":   sw.Namespaces,
			"schedule":     sw.Schedule,
		})
	}
	return
}

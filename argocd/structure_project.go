package argocd

import (
	argoCDAppv1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func expandProject(d *schema.ResourceData) (
	metadata metav1.ObjectMeta,
	spec argoCDAppv1.AppProjectSpec,
	err error) {
	metadata = expandMetadata(d.Get("metadata").([]interface{}))
	spec = expandProjectSpec(d.Get("metadata").([]interface{}))
	return
}

func expandProjectSpec(projectSpec []interface{}) (
	spec argoCDAppv1.AppProjectSpec) {
	s := projectSpec[0].(map[string]interface{})
	if v, ok := s["description"]; ok {
		spec.Description = v.(string)
	}
	if v, ok := s["source_repos"]; ok {
		for _, sr := range v.([]string) {
			spec.SourceRepos = append(spec.SourceRepos, sr)
		}
	}
	if v, ok := s["orphaned_resources"]; ok {
		if warn, ok := v.(map[string]bool)["warn"]; ok {
			spec.OrphanedResources.Warn = &warn
		}
	}
	// TODO: refactor
	if v, ok := s["cluster_resource_whitelist"]; ok {
		for _, gk := range v.([]map[string]string) {
			spec.ClusterResourceWhitelist = append(spec.ClusterResourceWhitelist, metav1.GroupKind{
				Group: gk["group"],
				Kind:  gk["kind"],
			})
		}
	}
	// TODO: refactor
	if v, ok := s["namespace_resource_blacklist"]; ok {
		for _, gk := range v.([]map[string]string) {
			spec.NamespaceResourceBlacklist = append(spec.NamespaceResourceBlacklist, metav1.GroupKind{
				Group: gk["group"],
				Kind:  gk["kind"],
			})
		}
	}
	if v, ok := s["destination"]; ok {
		for _, d := range v.([]map[string]string) {
			spec.Destinations = append(
				spec.Destinations,
				argoCDAppv1.ApplicationDestination{
					Server:    d["server"],
					Namespace: d["namespace"],
				},
			)
		}
	}
	if v, ok := s["sync_windows"]; ok {
		for _, sw := range v.([]map[string]interface{}) {
			spec.SyncWindows = append(
				spec.SyncWindows,
				&argoCDAppv1.SyncWindow{
					Applications: sw["applications"].([]string),
					Clusters:     sw["clusters"].([]string),
					Duration:     sw["duration"].(string),
					Kind:         sw["kind"].(string),
					ManualSync:   sw["manual_sync"].(bool),
					Namespaces:   sw["namespaces"].([]string),
					Schedule:     sw["schedule"].(string),
				},
			)
		}
	}
	return spec
}

func flattenProject(p *argoCDAppv1.AppProject, d *schema.ResourceData) map[string]interface{} {
	result := map[string]interface{}{
		"metadata": flattenMetadata(p.ObjectMeta, d),
		"spec": []map[string]interface{}{
			{
				"cluster_resource_whitelist": flattenK8SGroupKinds(
					p.Spec.ClusterResourceWhitelist),
				"namespace_resource_blacklist": flattenK8SGroupKinds(
					p.Spec.NamespaceResourceBlacklist),
				"destination": flattenDestinations(
					p.Spec.Destinations),
				"orphaned_resources": flattenOrphanedResources(
					p.Spec.OrphanedResources),
				"roles": flattenRoles(
					p.Spec.Roles),
				"sync_windows": flattenSyncWindows(
					p.Spec.SyncWindows),
				"description":  p.Spec.Description,
				"source_repos": p.Spec.SourceRepos,
			},
		},
	}
	return result
}

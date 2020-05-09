package argocd

import (
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func expandProjectRoles(roles []interface{}) (
	projectRoles []v1alpha1.ProjectRole,
	err error) {
	for _, _r := range roles {
		r := _r.(map[string]interface{})

		rolePolicies := expandStringList(r["policies"].([]interface{}))
		roleGroups := expandStringList(r["groups"].([]interface{}))

		projectRoles = append(
			projectRoles,
			v1alpha1.ProjectRole{
				Name:        r["name"].(string),
				Description: r["description"].(string),
				Policies:    rolePolicies,
				Groups:      roleGroups,
			},
		)
	}
	return
}

func flattenProjectOrphanedResources(ors *v1alpha1.OrphanedResourcesMonitorSettings) (
	result map[string]bool) {
	if ors != nil {
		result = map[string]bool{
			"warn": *ors.Warn,
		}
	}
	return
}

func flattenProjectRoles(rs []v1alpha1.ProjectRole) (
	result []map[string]interface{}) {
	for _, r := range rs {
		result = append(result, map[string]interface{}{
			"name":        r.Name,
			"description": r.Description,
			"groups":      r.Groups,
			"policies":    r.Policies,
		})
	}
	return
}

func expandProjectSpec(d *schema.ResourceData) (
	spec v1alpha1.AppProjectSpec,
	err error) {

	s := d.Get("spec.0").(map[string]interface{})

	if v, ok := s["description"]; ok {
		spec.Description = v.(string)
	}
	if v, ok := s["source_repos"]; ok {
		for _, sr := range v.([]interface{}) {
			spec.SourceRepos = append(spec.SourceRepos, sr.(string))
		}
	}
	if v, ok := s["orphaned_resources"]; ok {
		if _warn, ok := v.(map[string]interface{})["warn"]; ok {
			warn := _warn.(bool)
			spec.OrphanedResources = &v1alpha1.OrphanedResourcesMonitorSettings{
				Warn: &warn,
			}
		}
	}
	if v, ok := s["cluster_resource_whitelist"]; ok {
		spec.ClusterResourceWhitelist = expandK8SGroupKind(v.(*schema.Set))
	}
	if v, ok := s["namespace_resource_blacklist"]; ok {
		spec.NamespaceResourceBlacklist = expandK8SGroupKind(v.(*schema.Set))
	}
	if v, ok := s["destination"]; ok {
		spec.Destinations = expandApplicationDestination(v.(*schema.Set))
	}
	if v, ok := s["sync_window"]; ok {
		spec.SyncWindows = expandSyncWindows(v.([]interface{}))
	}
	if v, ok := s["role"]; ok {
		spec.Roles, err = expandProjectRoles(v.([]interface{}))
		if err != nil {
			return spec, err
		}
		for _, r := range spec.Roles {
			for _, p := range r.Policies {
				if err := validatePolicy(d.Get("metadata.0.name").(string), r.Name, p); err != nil {
					return spec, err
				}
			}
		}
	}
	return spec, nil
}

func flattenProjectSpec(s v1alpha1.AppProjectSpec) (
	[]map[string]interface{},
	error) {
	spec := map[string]interface{}{
		"cluster_resource_whitelist":   flattenK8SGroupKinds(s.ClusterResourceWhitelist),
		"namespace_resource_blacklist": flattenK8SGroupKinds(s.NamespaceResourceBlacklist),
		"destination":                  flattenApplicationDestinations(s.Destinations),
		"orphaned_resources":           flattenProjectOrphanedResources(s.OrphanedResources),
		"role":                         flattenProjectRoles(s.Roles),
		"sync_window":                  flattenSyncWindows(s.SyncWindows),
		"description":                  s.Description,
		"source_repos":                 s.SourceRepos,
	}
	return []map[string]interface{}{spec}, nil
}

func expandProject(d *schema.ResourceData) (
	metadata v1.ObjectMeta,
	spec v1alpha1.AppProjectSpec,
	err error) {
	metadata = expandMetadata(d)
	spec, err = expandProjectSpec(d)
	return
}

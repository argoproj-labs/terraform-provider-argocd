package argocd

import (
	"encoding/json"
	"fmt"

	application "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Expand

func expandProject(d *schema.ResourceData) (
	metadata meta.ObjectMeta,
	spec application.AppProjectSpec,
	err error) {
	metadata = expandMetadata(d)
	spec, err = expandProjectSpec(d)
	return
}

func expandProjectRoles(roles []interface{}) (
	projectRoles []application.ProjectRole,
	err error) {
	for _, _r := range roles {
		r := _r.(map[string]interface{})

		rolePolicies := expandStringList(r["policies"].([]interface{}))
		roleGroups := expandStringList(r["groups"].([]interface{}))

		projectRoles = append(
			projectRoles,
			application.ProjectRole{
				Name:        r["name"].(string),
				Description: r["description"].(string),
				Policies:    rolePolicies,
				Groups:      roleGroups,
			},
		)
	}
	return
}

func expandProjectSpec(d *schema.ResourceData) (
	spec application.AppProjectSpec,
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
	if v, ok := s["signature_keys"]; ok {
		for _, sk := range v.([]interface{}) {
			spec.SignatureKeys = append(spec.SignatureKeys, application.SignatureKey{
				KeyID: sk.(string),
			})
		}
	}
	if v, ok := s["orphaned_resources"]; ok {
		orphanedResources := v.([]interface{})
		if len(orphanedResources) > 0 {
			spec.OrphanedResources = &application.OrphanedResourcesMonitorSettings{}

			if orphanedResources[0] != nil {
				if _warn, _ok := orphanedResources[0].(map[string]interface{})["warn"]; _ok {
					warn := _warn.(bool)
					spec.OrphanedResources.Warn = &warn
				}
				if _ignore, _ok := orphanedResources[0].(map[string]interface{})["ignore"]; _ok {
					ignore := expandOrphanedResourcesIgnore(_ignore.(*schema.Set))
					spec.OrphanedResources.Ignore = ignore
				}
			}
		}
	}
	if v, ok := s["cluster_resource_blacklist"]; ok {
		spec.ClusterResourceBlacklist = expandK8SGroupKind(v.(*schema.Set))
	}
	if v, ok := s["cluster_resource_whitelist"]; ok {
		spec.ClusterResourceWhitelist = expandK8SGroupKind(v.(*schema.Set))
	}
	if v, ok := s["namespace_resource_blacklist"]; ok {
		spec.NamespaceResourceBlacklist = expandK8SGroupKind(v.(*schema.Set))
	}
	if v, ok := s["namespace_resource_whitelist"]; ok {
		spec.NamespaceResourceWhitelist = expandK8SGroupKind(v.(*schema.Set))
	}
	if v, ok := s["destination"]; ok {
		spec.Destinations = expandApplicationDestinations(v.(*schema.Set))
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

func expandOrphanedResourcesIgnore(ignore *schema.Set) (
	result []application.OrphanedResourceKey) {
	for _, _i := range ignore.List() {
		i := _i.(map[string]interface{})
		result = append(result, application.OrphanedResourceKey{
			Group: i["group"].(string),
			Kind:  i["kind"].(string),
			Name:  i["name"].(string),
		})
	}
	return
}

// Flatten

func flattenProject(p *application.AppProject, d *schema.ResourceData) error {
	fMetadata := flattenMetadata(p.ObjectMeta, d)
	fSpec := flattenProjectSpec(p.Spec)

	if err := d.Set("spec", fSpec); err != nil {
		e, _ := json.MarshalIndent(fSpec, "", "\t")
		return fmt.Errorf("error persisting spec: %s\n%s", err, e)
	}
	if err := d.Set("metadata", fMetadata); err != nil {
		e, _ := json.MarshalIndent(fMetadata, "", "\t")
		return fmt.Errorf("error persisting metadata: %s\n%s", err, e)
	}
	return nil
}

func flattenProjectSpec(s application.AppProjectSpec) []map[string]interface{} {
	spec := map[string]interface{}{
		"cluster_resource_blacklist":   flattenK8SGroupKinds(s.ClusterResourceBlacklist),
		"cluster_resource_whitelist":   flattenK8SGroupKinds(s.ClusterResourceWhitelist),
		"namespace_resource_blacklist": flattenK8SGroupKinds(s.NamespaceResourceBlacklist),
		"namespace_resource_whitelist": flattenK8SGroupKinds(s.NamespaceResourceWhitelist),
		"destination":                  flattenApplicationDestinations(s.Destinations),
		"orphaned_resources":           flattenProjectOrphanedResources(s.OrphanedResources),
		"role":                         flattenProjectRoles(s.Roles),
		"sync_window":                  flattenSyncWindows(s.SyncWindows),
		"description":                  s.Description,
		"source_repos":                 s.SourceRepos,
		"signature_keys":               flattenProjectSignatureKeys(s.SignatureKeys),
	}
	return []map[string]interface{}{spec}
}

func flattenProjectSignatureKeys(keys []application.SignatureKey) (
	result []string) {
	for _, key := range keys {
		result = append(result, key.KeyID)
	}
	return
}

func flattenProjectOrphanedResources(ors *application.OrphanedResourcesMonitorSettings) (
	result []map[string]interface{}) {
	r := make(map[string]interface{}, 0)
	if ors != nil {
		if ors.Warn != nil {
			r["warn"] = *ors.Warn
		}
		if ors.Ignore != nil {
			r["ignore"] = flattenProjectOrphanedResourcesIgnore(ors.Ignore)
		}
		result = append(result, r)
	}
	return
}

func flattenProjectOrphanedResourcesIgnore(ignore []application.OrphanedResourceKey) (
	result []map[string]string) {
	for _, i := range ignore {
		result = append(result, map[string]string{
			"group": i.Group,
			"kind":  i.Kind,
			"name":  i.Name,
		})
	}
	return
}

func flattenProjectRoles(rs []application.ProjectRole) (
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

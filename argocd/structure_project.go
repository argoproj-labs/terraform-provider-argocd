package argocd

import (
	"fmt"
	argoCDAppv1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func expandProject(d *schema.ResourceData) (
	metadata metav1.ObjectMeta,
	spec argoCDAppv1.AppProjectSpec,
	err error) {
	metadata = expandMetadata(d.Get("metadata").([]interface{}))
	spec, err = expandProjectSpec(d.Get("spec").([]interface{}))
	return
}

func expandProjectJWTTokens(jwts []interface{}) (
	jwtTokens []argoCDAppv1.JWTToken,
	err error) {

	for _, _jwt := range jwts {
		jwt := _jwt.(map[string]interface{})
		jwtToken := argoCDAppv1.JWTToken{}

		switch _iat, ok := jwt["iat"].(string); ok {
		case false:
			return jwtTokens, fmt.Errorf("iat is missing from role jwt_tokens: %#v", jwts)
		default:
			if _iat == "" {
				_iat = "0"
			}
			iat, err := convertStringToInt64(_iat)
			if err != nil {
				return jwtTokens, err
			}
			jwtToken.IssuedAt = iat
		}

		if _exp, ok := jwt["exp"].(string); ok {
			if _exp == "" {
				_exp = "0"
			}
			exp, err := convertStringToInt64(_exp)
			if err != nil {
				return jwtTokens, err
			}
			jwtToken.ExpiresAt = exp
		}

		jwtTokens = append(jwtTokens, jwtToken)
	}
	return
}

func expandProjectRoles(roles []interface{}) (
	projectRoles []argoCDAppv1.ProjectRole,
	err error) {
	for _, _r := range roles {
		r := _r.(map[string]interface{})

		rolePolicies := expandStringList(r["policies"].(*schema.Set).List())
		roleGroups := expandStringList(r["groups"].(*schema.Set).List())
		roleJWTTokens, err := expandProjectJWTTokens(r["jwt_token"].([]interface{}))
		if err != nil {
			return projectRoles, err
		}
		projectRoles = append(
			projectRoles,
			argoCDAppv1.ProjectRole{
				Name:        r["name"].(string),
				Description: r["description"].(string),
				Policies:    rolePolicies,
				Groups:      roleGroups,
				JWTTokens:   roleJWTTokens,
			},
		)
	}
	return
}

func expandProjectSpec(projectSpec []interface{}) (
	spec argoCDAppv1.AppProjectSpec,
	err error) {
	s := projectSpec[0].(map[string]interface{})
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
			spec.OrphanedResources = &argoCDAppv1.OrphanedResourcesMonitorSettings{
				Warn: &warn,
			}
		}
	}
	// TODO: refactor as expandProjectClusterResourceWhitelist
	if v, ok := s["cluster_resource_whitelist"]; ok {
		for _, _gk := range v.(*schema.Set).List() {
			gk := _gk.(map[string]interface{})
			spec.ClusterResourceWhitelist = append(spec.ClusterResourceWhitelist, metav1.GroupKind{
				Group: gk["group"].(string),
				Kind:  gk["kind"].(string),
			})
		}
	}
	// TODO: refactor as expandProjectNamespaceResourceBlacklist
	// ( == expandApplicationClusterResourceWhitelist)
	if v, ok := s["namespace_resource_blacklist"]; ok {
		for _, _gk := range v.(*schema.Set).List() {
			gk := _gk.(map[string]interface{})
			spec.NamespaceResourceBlacklist = append(spec.NamespaceResourceBlacklist, metav1.GroupKind{
				Group: gk["group"].(string),
				Kind:  gk["kind"].(string),
			})
		}
	}
	// TODO: refactor as expandApplicationDestination
	if v, ok := s["destination"]; ok {
		for _, _dest := range v.(*schema.Set).List() {
			dest := _dest.(map[string]interface{})
			spec.Destinations = append(
				spec.Destinations,
				argoCDAppv1.ApplicationDestination{
					Server:    dest["server"].(string),
					Namespace: dest["namespace"].(string),
				},
			)
		}
	}
	if v, ok := s["role"]; ok {
		spec.Roles, err = expandProjectRoles(v.([]interface{}))
		if err != nil {
			return spec, err
		}
	}
	// TODO: refactor as expandProjectSyncWindow
	if v, ok := s["sync_window"]; ok {
		for _, _sw := range v.([]interface{}) {
			sw := _sw.(map[string]interface{})
			spec.SyncWindows = append(
				spec.SyncWindows,
				&argoCDAppv1.SyncWindow{
					Applications: expandStringList(sw["applications"].([]interface{})),
					Clusters:     expandStringList(sw["clusters"].([]interface{})),
					Duration:     sw["duration"].(string),
					Kind:         sw["kind"].(string),
					ManualSync:   sw["manual_sync"].(bool),
					Namespaces:   expandStringList(sw["namespaces"].([]interface{})),
					Schedule:     sw["schedule"].(string),
				},
			)
		}
	}
	return spec, nil
}

func flattenProjectSpec(s argoCDAppv1.AppProjectSpec, d *schema.ResourceData) (
	result []map[string]interface{},
	err error) {

	roles := s.Roles

	if allow, ok := d.GetOk("allow_external_jwt_tokens"); ok && allow.(bool) {
		roles, err = strategicMergePatchJWTs(roles, d)
		if err != nil {
			return nil, err
		}
	}

	spec := map[string]interface{}{
		"cluster_resource_whitelist":   flattenK8SGroupKinds(s.ClusterResourceWhitelist),
		"namespace_resource_blacklist": flattenK8SGroupKinds(s.NamespaceResourceBlacklist),
		"destination":                  flattenDestinations(s.Destinations),
		"orphaned_resources":           flattenOrphanedResources(s.OrphanedResources),
		"role":                         flattenRoles(roles),
		"sync_window":                  flattenSyncWindows(s.SyncWindows),
		"description":                  s.Description,
		"source_repos":                 s.SourceRepos,
	}

	return []map[string]interface{}{spec}, nil
}

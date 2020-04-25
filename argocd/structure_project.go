package argocd

import (
	"encoding/json"
	"fmt"
	argoCDAppv1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
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

// Returns a role that matches the provided name,
// along with a boolean indicating whether there was a match.
func GetProjectRoleOk(name string, roles []argoCDAppv1.ProjectRole) (
	result argoCDAppv1.ProjectRole,
	ok bool) {
	for _, r := range roles {
		if r.Name == name {
			return r, true
		}
	}
	return
}

func strategicMergePatchJWTTokenSlice(overwrite bool, original, modified, current []argoCDAppv1.JWTToken) (
	result []argoCDAppv1.JWTToken,
	err error) {

	// Convert roles to []byte for use in strategic merge patch
	ob, err := json.Marshal(original)
	if err != nil {
		return nil, err
	}
	mb, err := json.Marshal(modified)
	if err != nil {
		return nil, err
	}
	cb, err := json.Marshal(current)
	if err != nil {
		return nil, err
	}

	patchMetaFromStruct, err := strategicpatch.NewPatchMetaFromStruct(argoCDAppv1.JWTToken{})
	if err != nil {
		return nil, err
	}
	lpmeta, pmeta, err := patchMetaFromStruct.LookupPatchMetadataForStruct("iat")
	if err != nil {
		return nil, err
	}
	fmt.Print(pmeta)

	// TODO: investigate json marshalling error
	// TODO: on a perfectly valid json document
	//
	// TODO: found: need to iterate over each JWT document
	// TODO: or better, find the method to handle slices merges (if it exists..)
	patch, err := strategicpatch.CreateThreeWayMergePatch(ob, mb, cb, lpmeta, overwrite)
	if err != nil {
		return result, fmt.Errorf("%s\n\nmodified: %#v\n\nmodifiedb: %s ", err, modified, mb)
		//return nil, err
	}

	rolesBytes, err := strategicpatch.StrategicMergePatch(
		ob,
		patch,
		argoCDAppv1.JWTToken{},
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(rolesBytes, &result); err != nil {
		eors, _ := json.MarshalIndent(original, "", "\t")
		emrs, _ := json.MarshalIndent(modified, "", "\t")
		ecrs, _ := json.MarshalIndent(current, "", "\t")
		return result, fmt.Errorf("%s\n\npatchmeta: %s\n\nsmpatch: %s\n\npatched: %s\n\noriginal: %s\n\nmodified: %s\n\ncurrent: %s ", err, patchMetaFromStruct, patch, rolesBytes, eors, emrs, ecrs)
	}
	return
}

func flattenProjectSpec(s argoCDAppv1.AppProjectSpec, d *schema.ResourceData) (
	[]map[string]interface{},
	error) {

	currentRoles := s.Roles

	// Allow for external JWTs to coexist with managed JWTs
	// by not persisting external JWTs to the state.
	allow := d.Get("allow_external_jwt_tokens").(bool)
	if allow {
		var patchedRoles []argoCDAppv1.ProjectRole

		oldRoles, modifiedRoles := d.GetChange("spec.0.role")
		ors, err := expandProjectRoles(oldRoles.([]interface{}))
		if err != nil {
			return nil, err
		}
		mrs, err := expandProjectRoles(modifiedRoles.([]interface{}))
		if err != nil {
			return nil, err
		}

		// Iterate over the modified roles slice,
		// since if old roles are missing from that slice,
		// it means they will be removed anyways.
		for _, mr := range mrs {
			or, _ := GetProjectRoleOk(mr.Name, ors)
			// No need to patch if the role does not currently exist.
			if cr, ok := GetProjectRoleOk(mr.Name, currentRoles); ok {
				roleJWTs, err := strategicMergePatchJWTTokenSlice(
					false,
					or.JWTTokens,
					mr.JWTTokens,
					cr.JWTTokens,
				)
				if err != nil {
					return nil, err
				}
				mr.JWTTokens = roleJWTs
			}
			patchedRoles = append(patchedRoles, mr)
		}
		currentRoles = patchedRoles
	}

	spec := map[string]interface{}{
		"cluster_resource_whitelist":   flattenK8SGroupKinds(s.ClusterResourceWhitelist),
		"namespace_resource_blacklist": flattenK8SGroupKinds(s.NamespaceResourceBlacklist),
		"destination":                  flattenDestinations(s.Destinations),
		"orphaned_resources":           flattenOrphanedResources(s.OrphanedResources),
		"role":                         flattenRoles(currentRoles),
		"sync_window":                  flattenSyncWindows(s.SyncWindows),
		"description":                  s.Description,
		"source_repos":                 s.SourceRepos,
	}
	//crs, _ := json.MarshalIndent(roles, "", "\t")
	//crcs, _ := json.MarshalIndent(s.Roles, "", "\t")
	//return nil, fmt.Errorf("computed: %s\n\ncurrent: %s ", crs, crcs)

	return []map[string]interface{}{spec}, nil
}

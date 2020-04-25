package argocd

import (
	"encoding/json"
	"fmt"
	argoCDAppv1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"strconv"
)

func convertStringToInt64(s string) (i int64, err error) {
	i, err = strconv.ParseInt(s, 10, 64)
	return
}

func convertInt64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}

func isKeyInMap(key string, d map[string]interface{}) bool {
	if d == nil {
		return false
	}
	for k := range d {
		if k == key {
			return true
		}
	}
	return false
}

func expandStringMap(m map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		result[k] = v.(string)
	}
	return result
}

func expandStringList(l []interface{}) (
	result []string) {
	for _, p := range l {
		result = append(result, p.(string))
	}
	return
}

func strategicMergePatchJWTs(currentRoles []argoCDAppv1.ProjectRole, d *schema.ResourceData) (
	result []argoCDAppv1.ProjectRole,
	err error) {

	originalRoles, modifiedRoles := d.GetChange("spec.0.role")
	ors, err := expandProjectRoles(originalRoles.([]interface{}))
	if err != nil {
		return result, err
	}
	mrs, err := expandProjectRoles(modifiedRoles.([]interface{}))
	if err != nil {
		return result, err
	}

	// Convert roles to []byte for use in strategic merge patch
	originalRolesBytes, err := json.Marshal(ors)
	if err != nil {
		return result, err
	}
	modifiedRolesBytes, err := json.Marshal(mrs)
	if err != nil {
		return result, err
	}
	currentRolesBytes, err := json.Marshal(currentRoles)
	if err != nil {
		return result, err
	}

	patchMeta, err := strategicpatch.NewPatchMetaFromStruct(argoCDAppv1.ProjectRole{})
	if err != nil {
		return result, err
	}
	patch, err := strategicpatch.CreateThreeWayMergePatch(
		originalRolesBytes,
		modifiedRolesBytes,
		currentRolesBytes,
		patchMeta,
		false,
	)
	rolesBytes, err := strategicpatch.StrategicMergePatch(
		originalRolesBytes,
		patch,
		argoCDAppv1.ProjectRole{},
	)
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal(rolesBytes, &result); err != nil {
		return result, fmt.Errorf("%s %s", err, rolesBytes)
	}
	return
}

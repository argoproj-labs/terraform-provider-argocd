package argocd

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/argoproj/argo-cd/v2/server/rbacpolicy"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/oboukili/terraform-provider-argocd/internal/features"
)

func convertStringToInt64(s string) (i int64, err error) {
	i, err = strconv.ParseInt(s, 10, 64)
	return
}

func convertInt64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}

func convertInt64PointerToString(i *int64) string {
	return strconv.FormatInt(*i, 10)
}

func convertStringToInt64Pointer(s string) (*int64, error) {
	i, err := convertStringToInt64(s)
	if err != nil {
		return nil, fmt.Errorf("not a valid int64: %s", s)
	}

	return &i, nil
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

func expandStringList(l []interface{}) (result []string) {
	for _, p := range l {
		result = append(result, p.(string))
	}

	return
}

func sliceOfString(slice []interface{}) []string {
	result := make([]string, len(slice))

	for i, s := range slice {
		result[i] = s.(string)
	}

	return result
}

func isValidPolicyAction(action string) bool {
	validActions := map[string]bool{
		rbacpolicy.ActionGet:      true,
		rbacpolicy.ActionCreate:   true,
		rbacpolicy.ActionUpdate:   true,
		rbacpolicy.ActionDelete:   true,
		rbacpolicy.ActionSync:     true,
		rbacpolicy.ActionOverride: true,
		"*":                       true,
	}
	validActionPatterns := []*regexp.Regexp{
		regexp.MustCompile("action/.*"),
	}

	if validActions[action] {
		return true
	}

	for i := range validActionPatterns {
		if validActionPatterns[i].MatchString(action) {
			return true
		}
	}

	return false
}

func validatePolicy(project string, role string, policy string) error {
	policyComponents := strings.Split(policy, ",")
	if len(policyComponents) != 6 || strings.Trim(policyComponents[0], " ") != "p" {
		return fmt.Errorf("invalid policy rule '%s': must be of the form: 'p, sub, res, act, obj, eft'", policy)
	}

	// subject
	subject := strings.Trim(policyComponents[1], " ")
	expectedSubject := fmt.Sprintf("proj:%s:%s", project, role)

	if subject != expectedSubject {
		return fmt.Errorf("invalid policy rule '%s': policy subject must be: '%s', not '%s'", policy, expectedSubject, subject)
	}

	// resource
	// https://github.com/argoproj/argo-cd/blob/c99669e088b5f25c8ce8faff6df25797a8beb5ba/pkg/apis/application/v1alpha1/types.go#L1554
	validResources := map[string]bool{
		rbacpolicy.ResourceApplications: true,
		rbacpolicy.ResourceRepositories: true,
		rbacpolicy.ResourceClusters:     true,
		rbacpolicy.ResourceExec:         true,
		rbacpolicy.ResourceLogs:         true,
	}

	resource := strings.Trim(policyComponents[2], " ")
	if !validResources[resource] {
		return fmt.Errorf("invalid policy rule '%s': resource '%s' not recognised", policy, resource)
	}

	// action
	action := strings.Trim(policyComponents[3], " ")
	if !isValidPolicyAction(action) {
		return fmt.Errorf("invalid policy rule '%s': invalid action '%s'", policy, action)
	}

	// object
	object := strings.Trim(policyComponents[4], " ")

	objectRegexp, err := regexp.Compile(fmt.Sprintf(`^%s/[*\w-.]+$`, project))
	if err != nil || !objectRegexp.MatchString(object) {
		return fmt.Errorf("invalid policy rule '%s': object must be of form '%s/*' or '%s/<APPNAME>', not '%s'", policy, project, project, object)
	}

	// effect
	effect := strings.Trim(policyComponents[5], " ")
	if effect != "allow" && effect != "deny" {
		return fmt.Errorf("invalid policy rule '%s': effect must be: 'allow' or 'deny'", policy)
	}

	return nil
}

func persistToState(key string, data interface{}, d *schema.ResourceData) error {
	if err := d.Set(key, data); err != nil {
		return fmt.Errorf("error persisting %s: %s", key, err)
	}

	return nil
}

func argoCDAPIError(action, resource, id string, err error) diag.Diagnostics {
	return []diag.Diagnostic{
		{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("failed to %s %s %s", action, resource, id),
			Detail:   err.Error(),
		},
	}
}

func errorToDiagnostics(summary string, err error) diag.Diagnostics {
	d := diag.Diagnostic{
		Severity: diag.Error,
		Summary:  summary,
	}

	if err != nil {
		d.Detail = err.Error()
	}

	return []diag.Diagnostic{d}
}

func featureNotSupported(feature features.Feature) diag.Diagnostics {
	f := features.ConstraintsMap[feature]

	return []diag.Diagnostic{
		{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("%s is only supported from ArgoCD %s onwards", f.Name, f.MinVersion.String()),
		},
	}
}

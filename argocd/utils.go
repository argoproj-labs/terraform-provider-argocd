package argocd

import (
	"fmt"
	"strconv"

	"github.com/argoproj-labs/terraform-provider-argocd/internal/features"

	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

func expandBoolMap(m map[string]interface{}) map[string]bool {
	result := make(map[string]bool)

	for k, v := range m {
		result[k] = v.(bool)
	}

	return result
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

// pluginSDKDiags converts diagnostics from `terraform-plugin-framework/diag` to
// `terraform-plugin-sdk/v2/diag`
func pluginSDKDiags(ds fwdiag.Diagnostics) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, d := range ds {
		_diag := diag.Diagnostic{
			Detail:  d.Detail(),
			Summary: d.Summary(),
		}

		switch d.Severity() {
		case fwdiag.SeverityError:
			_diag.Severity = diag.Error
		default:
			_diag.Severity = diag.Warning
		}

		diags = append(diags, _diag)
	}

	return diags
}

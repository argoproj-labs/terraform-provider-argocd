package diagnostics

import (
	"fmt"

	"github.com/argoproj-labs/terraform-provider-argocd/internal/features"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func ArgoCDAPIError(action, resource, id string, err error) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.AddError(fmt.Sprintf("failed to %s %s %s", action, resource, id), err.Error())

	return diags
}

func Error(summary string, err error) diag.Diagnostics {
	var diags diag.Diagnostics

	var detail string

	if err != nil {
		detail = err.Error()
	}

	diags.AddError(summary, detail)

	return diags
}

func FeatureNotSupported(f features.Feature) diag.Diagnostics {
	var diags diag.Diagnostics

	fc := features.ConstraintsMap[f]

	diags.AddError(fmt.Sprintf("%s is only supported from ArgoCD %s onwards", fc.Name, fc.MinVersion.String()), "")

	return diags
}

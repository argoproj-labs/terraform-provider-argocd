package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Bool = enableOCIValidator{}

type enableOCIValidator struct{}

func (v enableOCIValidator) Description(_ context.Context) string {
	return "enable_oci can only be set to true when type is 'helm'"
}

func (v enableOCIValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v enableOCIValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	// If the value is null or unknown, no validation needed
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// Only validate if enable_oci is true
	if !req.ConfigValue.ValueBool() {
		return
	}

	// Get the type attribute value
	var typeValue attr.Value
	diags := req.Config.GetAttribute(ctx, path.Root("type"), &typeValue)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If type is unknown, we can't validate yet (will be validated during apply)
	if typeValue.IsUnknown() {
		return
	}

	// If type is null, it will default to "git", which is invalid for enable_oci=true
	if typeValue.IsNull() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Configuration",
			"enable_oci can only be set to true when type is 'helm'",
		)
		return
	}

	// Check if type is "helm"
	typeStr, ok := typeValue.(interface{ ValueString() string })
	if !ok {
		// This shouldn't happen, but handle it gracefully
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Configuration",
			"Unable to validate enable_oci: type attribute has unexpected type",
		)
		return
	}

	if typeStr.ValueString() != "helm" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Configuration",
			"enable_oci can only be set to true when type is 'helm', but type is '"+typeStr.ValueString()+"'",
		)
	}
}

// EnableOCIRequiresHelmType returns a validator that ensures enable_oci is only true when type is "helm"
func EnableOCIRequiresHelmType() validator.Bool {
	return enableOCIValidator{}
}

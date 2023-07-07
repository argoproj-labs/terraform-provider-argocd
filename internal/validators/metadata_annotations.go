package validators

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"k8s.io/apimachinery/pkg/util/validation"
)

var _ validator.Map = (*metadataAnnotationsValidator)(nil)

type metadataAnnotationsValidator struct{}

func MetadataAnnotations() metadataAnnotationsValidator {
	return metadataAnnotationsValidator{}
}

// Description returns a plain text description of the validator's behavior, suitable for a practitioner to understand its impact.
func (v metadataAnnotationsValidator) Description(ctx context.Context) string {
	return "ensures that all keys in the supplied map are valid qualified names"
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a practitioner to understand its impact.
func (v metadataAnnotationsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate runs the main validation logic of the validator, reading configuration data out of `req` and updating `resp` with diagnostics.
func (v metadataAnnotationsValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	// If the value is unknown or null, there is nothing to validate.
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	var m map[string]string

	resp.Diagnostics.Append(req.ConfigValue.ElementsAs(ctx, &m, false)...)

	if resp.Diagnostics.HasError() {
		return
	}

	for k := range m {
		errors := validation.IsQualifiedName(strings.ToLower(k))
		for _, err := range errors {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid Annotation Key: not a valid qualified name",
				err)
		}
	}
}

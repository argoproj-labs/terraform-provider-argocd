package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"k8s.io/apimachinery/pkg/api/validation"
)

var _ validator.String = (*isDNSSubdomainValidator)(nil)

type isDNSSubdomainValidator struct{}

func IsDNSSubdomain() isDNSSubdomainValidator {
	return isDNSSubdomainValidator{}
}

// Description returns a plain text description of the validator's behavior, suitable for a practitioner to understand its impact.
func (v isDNSSubdomainValidator) Description(ctx context.Context) string {
	return "ensures that attribute is a valid DNS subdomain"
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a practitioner to understand its impact.
func (v isDNSSubdomainValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate runs the main validation logic of the validator, reading configuration data out of `req` and updating `resp` with diagnostics.
func (v isDNSSubdomainValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If the value is unknown or null, there is nothing to validate.
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	errors := validation.NameIsDNSSubdomain(req.ConfigValue.ValueString(), false)
	for _, err := range errors {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid DNS subdomain",
			err)
	}
}

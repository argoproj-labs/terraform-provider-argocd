package validators

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = positiveIntegerValidator{}

type positiveIntegerValidator struct{}

func (v positiveIntegerValidator) Description(_ context.Context) string {
	return "value must be a positive integer"
}

func (v positiveIntegerValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v positiveIntegerValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		return
	}

	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Integer",
			"The provided value is not a valid integer: "+err.Error(),
		)

		return
	}

	if i <= 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Positive Integer",
			"The provided value must be a positive integer (greater than 0)",
		)

		return
	}
}

func PositiveInteger() validator.String {
	return positiveIntegerValidator{}
}

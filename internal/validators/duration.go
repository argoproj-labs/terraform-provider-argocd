package validators

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// DurationValidator returns a validator which ensures that any configured
// attribute value is a valid duration string.
func DurationValidator() validator.String {
	return durationValidator{}
}

type durationValidator struct{}

func (v durationValidator) Description(ctx context.Context) string {
	return "value must be a valid duration string"
}

func (v durationValidator) MarkdownDescription(ctx context.Context) string {
	return "value must be a valid duration string"
}

func (v durationValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if _, err := time.ParseDuration(value); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Duration",
			fmt.Sprintf("cannot parse duration '%s': %s", value, err.Error()),
		)
	}
}

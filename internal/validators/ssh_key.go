package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"golang.org/x/crypto/ssh"
)

var _ validator.String = (*isSSHPrivateKeyValidator)(nil)

type isSSHPrivateKeyValidator struct{}

func IsSSHPrivateKey() isSSHPrivateKeyValidator {
	return isSSHPrivateKeyValidator{}
}

// Description returns a plain text description of the validator's behavior, suitable for a practitioner to understand its impact.
func (v isSSHPrivateKeyValidator) Description(ctx context.Context) string {
	return "ensures that attribute is a valid SSH private key"
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a practitioner to understand its impact.
func (v isSSHPrivateKeyValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate runs the main validation logic of the validator, reading configuration data out of `req` and updating `resp` with diagnostics.
func (v isSSHPrivateKeyValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If the value is unknown or null, there is nothing to validate.
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	_, err := ssh.ParsePrivateKey([]byte(req.ConfigValue.String()))
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid SSH key",
			err.Error(),
		)
	}
}

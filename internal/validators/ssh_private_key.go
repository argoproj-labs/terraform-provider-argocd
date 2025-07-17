package validators

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = sshPrivateKeyValidator{}

type sshPrivateKeyValidator struct{}

func (v sshPrivateKeyValidator) Description(_ context.Context) string {
	return "value must be a valid SSH private key in PEM format"
}

func (v sshPrivateKeyValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v sshPrivateKeyValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		return
	}

	// Check if it's a valid PEM block
	block, _ := pem.Decode([]byte(value))
	if block == nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid SSH Private Key",
			"The provided value is not a valid PEM-encoded private key",
		)

		return
	}

	// Check if it's a recognized private key type
	validTypes := []string{
		"RSA PRIVATE KEY",
		"PRIVATE KEY",
		"EC PRIVATE KEY",
		"DSA PRIVATE KEY",
		"OPENSSH PRIVATE KEY",
	}

	isValidType := false

	for _, validType := range validTypes {
		if strings.EqualFold(block.Type, validType) {
			isValidType = true
			break
		}
	}

	if !isValidType {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid SSH Private Key Type",
			"The provided PEM block is not a recognized private key type",
		)

		return
	}

	// Additional validation for PKCS#8 and PKCS#1 formats
	if strings.EqualFold(block.Type, "PRIVATE KEY") {
		// PKCS#8 format
		_, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid PKCS#8 Private Key",
				"The provided PKCS#8 private key is invalid: "+err.Error(),
			)

			return
		}
	} else if strings.EqualFold(block.Type, "RSA PRIVATE KEY") {
		// PKCS#1 format
		_, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid PKCS#1 Private Key",
				"The provided PKCS#1 private key is invalid: "+err.Error(),
			)

			return
		}
	}
}

func SSHPrivateKey() validator.String {
	return sshPrivateKeyValidator{}
}

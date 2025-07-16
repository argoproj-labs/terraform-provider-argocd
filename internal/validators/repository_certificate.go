package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ConfigValidator = repositoryCertificateValidator{}

type repositoryCertificateValidator struct{}

func (v repositoryCertificateValidator) Description(_ context.Context) string {
	return "one of `https,ssh` must be specified"
}

func (v repositoryCertificateValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v repositoryCertificateValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var ssh types.List

	var https types.List

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("ssh"), &ssh)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("https"), &https)...)

	if resp.Diagnostics.HasError() {
		return
	}

	sshConfigured := !ssh.IsNull() && len(ssh.Elements()) > 0
	httpsConfigured := !https.IsNull() && len(https.Elements()) > 0

	// Validate that each list contains at most one element
	if sshConfigured && len(ssh.Elements()) > 1 {
		resp.Diagnostics.AddError(
			"Too many SSH certificates",
			"Only one SSH certificate can be specified",
		)

		return
	}

	if httpsConfigured && len(https.Elements()) > 1 {
		resp.Diagnostics.AddError(
			"Too many HTTPS certificates",
			"Only one HTTPS certificate can be specified",
		)

		return
	}

	if !sshConfigured && !httpsConfigured {
		resp.Diagnostics.AddError(
			"Missing required configuration",
			"one of `https,ssh` must be specified",
		)

		return
	}

	if sshConfigured && httpsConfigured {
		resp.Diagnostics.AddError(
			"Conflicting configuration",
			"only one of `https,ssh` can be specified",
		)

		return
	}
	// SSH and HTTPS block fields are required at the schema level, so no additional validation needed
}

func RepositoryCertificate() resource.ConfigValidator {
	return repositoryCertificateValidator{}
}

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
	var https types.Object

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("ssh"), &ssh)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("https"), &https)...)

	if resp.Diagnostics.HasError() {
		return
	}

	sshConfigured := !ssh.IsNull() && len(ssh.Elements()) > 0
	httpsConfigured := !https.IsNull()

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

	// Validate SSH block fields if SSH is configured
	if sshConfigured {
		var sshServerName, sshCertSubtype, sshCertData types.String

		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("ssh").AtListIndex(0).AtName("server_name"), &sshServerName)...)
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("ssh").AtListIndex(0).AtName("cert_subtype"), &sshCertSubtype)...)
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("ssh").AtListIndex(0).AtName("cert_data"), &sshCertData)...)

		if resp.Diagnostics.HasError() {
			return
		}

		if sshServerName.IsNull() || sshServerName.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("ssh").AtListIndex(0).AtName("server_name"),
				"Missing required attribute",
				"ssh.server_name is required when ssh block is specified",
			)
		}

		if sshCertSubtype.IsNull() || sshCertSubtype.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("ssh").AtListIndex(0).AtName("cert_subtype"),
				"Missing required attribute",
				"ssh.cert_subtype is required when ssh block is specified",
			)
		}

		if sshCertData.IsNull() || sshCertData.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("ssh").AtListIndex(0).AtName("cert_data"),
				"Missing required attribute",
				"ssh.cert_data is required when ssh block is specified",
			)
		}
	}

	// Validate HTTPS block fields if HTTPS is configured
	if httpsConfigured {
		var httpsServerName, httpsCertData types.String

		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("https").AtName("server_name"), &httpsServerName)...)
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("https").AtName("cert_data"), &httpsCertData)...)

		if resp.Diagnostics.HasError() {
			return
		}

		if httpsServerName.IsNull() || httpsServerName.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("https").AtName("server_name"),
				"Missing required attribute",
				"https.server_name is required when https block is specified",
			)
		}

		if httpsCertData.IsNull() || httpsCertData.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("https").AtName("cert_data"),
				"Missing required attribute",
				"https.cert_data is required when https block is specified",
			)
		}
	}
}

func RepositoryCertificate() resource.ConfigValidator {
	return repositoryCertificateValidator{}
}
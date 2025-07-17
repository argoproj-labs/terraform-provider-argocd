package provider

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type accountTokenResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Account     types.String `tfsdk:"account"`
	ExpiresIn   types.String `tfsdk:"expires_in"`
	RenewAfter  types.String `tfsdk:"renew_after"`
	RenewBefore types.String `tfsdk:"renew_before"`
	JWT         types.String `tfsdk:"jwt"`
	IssuedAt    types.String `tfsdk:"issued_at"`
	ExpiresAt   types.String `tfsdk:"expires_at"`
}

func accountTokenResourceSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			MarkdownDescription: "The ID of the token.",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"account": rschema.StringAttribute{
			MarkdownDescription: "Account name. Defaults to the current account. I.e. the account configured on the `provider` block.",
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"expires_in": rschema.StringAttribute{
			MarkdownDescription: "Duration before the token will expire. Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`. E.g. `12h`, `7d`. Default: No expiration.",
			Optional:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				stringvalidator.RegexMatches(
					// Duration regex pattern
					regexp.MustCompile(`^(\d+(\.\d*)?[a-zA-Z]+)+$`),
					"must be a valid duration string (e.g., '1h', '30m', '1h30m')",
				),
			},
		},
		"renew_after": rschema.StringAttribute{
			MarkdownDescription: "Duration to control token silent regeneration based on token age. Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`. If set, then the token will be regenerated if it is older than `renew_after`. I.e. if `currentDate - issued_at > renew_after`.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(
					// Duration regex pattern
					regexp.MustCompile(`^(\d+(\.\d*)?[a-zA-Z]+)+$`),
					"must be a valid duration string (e.g., '1h', '30m', '1h30m')",
				),
			},
		},
		"renew_before": rschema.StringAttribute{
			MarkdownDescription: "Duration to control token silent regeneration based on remaining token lifetime. If `expires_in` is set, Terraform will regenerate the token if `expires_at - currentDate < renew_before`. Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(
					// Duration regex pattern
					regexp.MustCompile(`^(\d+(\.\d*)?[a-zA-Z]+)+$`),
					"must be a valid duration string (e.g., '1h', '30m', '1h30m')",
				),
				stringvalidator.AlsoRequires(path.MatchRoot("expires_in")),
			},
		},
		"jwt": rschema.StringAttribute{
			MarkdownDescription: "The raw JWT.",
			Computed:            true,
			Sensitive:           true,
		},
		"issued_at": rschema.StringAttribute{
			MarkdownDescription: "Unix timestamp at which the token was issued.",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"expires_at": rschema.StringAttribute{
			MarkdownDescription: "If `expires_in` is set, Unix timestamp upon which the token will expire.",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	}
}

package provider

import (
	"github.com/argoproj-labs/terraform-provider-argocd/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type projectTokenModel struct {
	ID          types.String `tfsdk:"id"`
	Project     types.String `tfsdk:"project"`
	Role        types.String `tfsdk:"role"`
	ExpiresIn   types.String `tfsdk:"expires_in"`
	RenewAfter  types.String `tfsdk:"renew_after"`
	RenewBefore types.String `tfsdk:"renew_before"`
	Description types.String `tfsdk:"description"`
	JWT         types.String `tfsdk:"jwt"`
	IssuedAt    types.String `tfsdk:"issued_at"`
	ExpiresAt   types.String `tfsdk:"expires_at"`
}

func projectTokenSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Token identifier",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"project": schema.StringAttribute{
			Description: "The project associated with the token.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"role": schema.StringAttribute{
			Description: "The name of the role in the project associated with the token.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"expires_in": schema.StringAttribute{
			Description: "Duration before the token will expire. Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`. E.g. `12h`, `7d`. Default: No expiration.",
			Optional:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				validators.DurationValidator(),
			},
		},
		"renew_after": schema.StringAttribute{
			Description: "Duration to control token silent regeneration based on token age. Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`. If set, then the token will be regenerated if it is older than `renew_after`. I.e. if `currentDate - issued_at > renew_after`.",
			Optional:    true,
			Validators: []validator.String{
				validators.DurationValidator(),
			},
		},
		"renew_before": schema.StringAttribute{
			Description: "Duration to control token silent regeneration based on remaining token lifetime. If `expires_in` is set, Terraform will regenerate the token if `expires_at - currentDate < renew_before`. Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.",
			Optional:    true,
			Validators: []validator.String{
				validators.DurationValidator(),
			},
		},
		"description": schema.StringAttribute{
			Description: "Description of the token.",
			Optional:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"jwt": schema.StringAttribute{
			Description: "The raw JWT.",
			Computed:    true,
			Sensitive:   true,
		},
		"issued_at": schema.StringAttribute{
			Description: "Unix timestamp at which the token was issued.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"expires_at": schema.StringAttribute{
			Description: "If `expires_in` is set, Unix timestamp upon which the token will expire.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	}
}

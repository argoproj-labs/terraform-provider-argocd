package provider

import (
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type accountModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Password     types.String `tfsdk:"password"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	Capabilities types.List   `tfsdk:"capabilities"`
	Tokens       types.List   `tfsdk:"tokens"`
}

type accountDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	Capabilities types.List   `tfsdk:"capabilities"`
	Tokens       types.List   `tfsdk:"tokens"`
}

type accountTokenModel struct {
	ID        types.String `tfsdk:"id"`
	IssuedAt  types.String `tfsdk:"issued_at"`
	ExpiresAt types.String `tfsdk:"expires_at"`
}

func accountResourceSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			MarkdownDescription: "The ID of the account (same as name).",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": rschema.StringAttribute{
			MarkdownDescription: "The name of the ArgoCD account.",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"password": rschema.StringAttribute{
			MarkdownDescription: "The password for the account. When changed, will update the account password using the previous state value as the current password. Note: ArgoCD API requires the current password to update/set a password, even for accounts without existing passwords. Initial passwords must be set through ArgoCD configuration.",
			Optional:            true,
			Sensitive:           true,
		},
		"enabled": rschema.BoolAttribute{
			MarkdownDescription: "Whether the account is enabled.",
			Computed:            true,
		},
		"capabilities": rschema.ListAttribute{
			MarkdownDescription: "List of account capabilities.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"tokens": rschema.ListNestedAttribute{
			MarkdownDescription: "List of active tokens for the account.",
			Computed:            true,
			NestedObject: rschema.NestedAttributeObject{
				Attributes: map[string]rschema.Attribute{
					"id": rschema.StringAttribute{
						MarkdownDescription: "Token ID.",
						Computed:            true,
					},
					"issued_at": rschema.StringAttribute{
						MarkdownDescription: "Unix timestamp when the token was issued.",
						Computed:            true,
					},
					"expires_at": rschema.StringAttribute{
						MarkdownDescription: "Unix timestamp when the token expires, 0 if no expiration.",
						Computed:            true,
					},
				},
			},
		},
	}
}

func accountDataSourceSchemaAttributes() map[string]dsschema.Attribute {
	return map[string]dsschema.Attribute{
		"id": dsschema.StringAttribute{
			MarkdownDescription: "The ID of the account (same as name).",
			Computed:            true,
		},
		"name": dsschema.StringAttribute{
			MarkdownDescription: "The name of the ArgoCD account to retrieve.",
			Required:            true,
		},
		"enabled": dsschema.BoolAttribute{
			MarkdownDescription: "Whether the account is enabled.",
			Computed:            true,
		},
		"capabilities": dsschema.ListAttribute{
			MarkdownDescription: "List of account capabilities.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"tokens": dsschema.ListNestedAttribute{
			MarkdownDescription: "List of active tokens for the account.",
			Computed:            true,
			NestedObject: dsschema.NestedAttributeObject{
				Attributes: map[string]dsschema.Attribute{
					"id": dsschema.StringAttribute{
						MarkdownDescription: "Token ID.",
						Computed:            true,
					},
					"issued_at": dsschema.StringAttribute{
						MarkdownDescription: "Unix timestamp when the token was issued.",
						Computed:            true,
					},
					"expires_at": dsschema.StringAttribute{
						MarkdownDescription: "Unix timestamp when the token expires, 0 if no expiration.",
						Computed:            true,
					},
				},
			},
		},
	}
}

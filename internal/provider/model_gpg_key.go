package provider

import (
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	customtypes "github.com/oboukili/terraform-provider-argocd/internal/types"
)

type gpgKeyModel struct {
	ID          types.String             `tfsdk:"id"`
	PublicKey   customtypes.PGPPublicKey `tfsdk:"public_key"`
	Fingerprint types.String             `tfsdk:"fingerprint"`
	Owner       types.String             `tfsdk:"owner"`
	SubType     types.String             `tfsdk:"sub_type"`
	Trust       types.String             `tfsdk:"trust"`
}

func gpgKeySchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"public_key": schema.StringAttribute{
			MarkdownDescription: "Raw key data of the GPG key to create",
			CustomType:          customtypes.PGPPublicKeyType,
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"fingerprint": schema.StringAttribute{
			MarkdownDescription: "Fingerprint is the fingerprint of the key",
			Computed:            true,
		},
		"id": schema.StringAttribute{
			MarkdownDescription: "GPG key identifier",
			Computed:            true,
		},
		"owner": schema.StringAttribute{
			MarkdownDescription: "Owner holds the owner identification, e.g. a name and e-mail address",
			Computed:            true,
		},
		"sub_type": schema.StringAttribute{
			MarkdownDescription: "SubType holds the key's sub type (e.g. rsa4096)",
			Computed:            true,
		},
		"trust": schema.StringAttribute{
			MarkdownDescription: "Trust holds the level of trust assigned to this key",
			Computed:            true,
		},
	}
}

func newGPGKey(k *v1alpha1.GnuPGPublicKey) *gpgKeyModel {
	return &gpgKeyModel{
		Fingerprint: types.StringValue(k.Fingerprint),
		ID:          types.StringValue(k.KeyID),
		Owner:       types.StringValue(k.Owner),
		PublicKey:   customtypes.PGPPublicKeyValue(k.KeyData),
		SubType:     types.StringValue(k.SubType),
		Trust:       types.StringValue(k.Trust),
	}
}

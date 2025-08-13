package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/argoproj-labs/terraform-provider-argocd/internal/diagnostics"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/account"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &accountDataSource{}

func NewAccountDataSource() datasource.DataSource {
	return &accountDataSource{}
}

// accountDataSource defines the data source implementation.
type accountDataSource struct {
	si *ServerInterface
}

func (d *accountDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account"
}

func (d *accountDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about a specific ArgoCD account.",
		Attributes:          accountDataSourceSchemaAttributes(),
	}
}

func (d *accountDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	si, ok := req.ProviderData.(*ServerInterface)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *ServerInterface, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.si = si
}

func (d *accountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data accountDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	// Initialize API clients
	resp.Diagnostics.Append(d.si.InitClients(ctx)...)

	if resp.Diagnostics.HasError() {
		return
	}

	accountName := data.Name.ValueString()

	tflog.Trace(ctx, fmt.Sprintf("reading account data source for %s", accountName))

	// Get account from ArgoCD
	accountResp, err := d.si.AccountClient.GetAccount(ctx, &account.GetAccountRequest{
		Name: accountName,
	})

	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			resp.Diagnostics.AddError("Account not found", fmt.Sprintf("Account %s not found", accountName))
			return
		}

		resp.Diagnostics.Append(diagnostics.ArgoCDAPIError("read", "account", accountName, err)...)

		return
	}

	if accountResp == nil {
		resp.Diagnostics.AddError("Account not found", fmt.Sprintf("Account %s not found", accountName))
		return
	}

	// Update model with response data
	data.ID = types.StringValue(accountResp.Name)
	data.Name = types.StringValue(accountResp.Name)
	data.Enabled = types.BoolValue(accountResp.Enabled)

	// Convert capabilities to types.List
	capabilities := make([]types.String, len(accountResp.Capabilities))
	for i, cap := range accountResp.Capabilities {
		capabilities[i] = types.StringValue(cap)
	}

	capList, diag := types.ListValueFrom(ctx, types.StringType, capabilities)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	data.Capabilities = capList

	// Convert tokens to types.List
	tokens := make([]accountTokenModel, len(accountResp.Tokens))
	for i, token := range accountResp.Tokens {
		tokens[i] = accountTokenModel{
			ID:        types.StringValue(token.Id),
			IssuedAt:  types.StringValue(strconv.FormatInt(token.IssuedAt, 10)),
			ExpiresAt: types.StringValue(strconv.FormatInt(token.ExpiresAt, 10)),
		}
	}

	tokenList, diag := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":         types.StringType,
			"issued_at":  types.StringType,
			"expires_at": types.StringType,
		},
	}, tokens)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	data.Tokens = tokenList

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

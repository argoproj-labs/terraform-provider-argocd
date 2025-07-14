package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/argoproj-labs/terraform-provider-argocd/internal/diagnostics"
	"github.com/argoproj-labs/terraform-provider-argocd/internal/sync"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/account"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &accountResource{}

func NewAccountResource() resource.Resource {
	return &accountResource{}
}

// accountResource defines the resource implementation.
type accountResource struct {
	si *ServerInterface
}

func (r *accountResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account"
}

func (r *accountResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: "Manages ArgoCD [accounts](https://argo-cd.readthedocs.io/en/latest/operator-manual/user-management/) for user authentication and authorization.\n\n~> **Note** This resource manages account information and password updates. When the password is changed, the previous state value is used as the current password for the API call. For token management, use the `argocd_account_token` resource.",
		Attributes:          accountResourceSchemaAttributes(),
	}
}

func (r *accountResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.si = si
}

func (r *accountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data accountModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	// Initialize API clients
	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	// Check for errors before proceeding
	if resp.Diagnostics.HasError() {
		return
	}

	// ArgoCD accounts are typically created through ArgoCD configuration
	// This resource primarily manages existing accounts
	accountName := data.Name.ValueString()

	tflog.Trace(ctx, fmt.Sprintf("creating account resource for %s", accountName))

	// Verify the account exists by reading it
	diags := r.readAccount(ctx, &data, accountName)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Set the ID to the account name
	data.ID = types.StringValue(accountName)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *accountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data accountModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	// Initialize API clients
	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	if resp.Diagnostics.HasError() {
		return
	}

	accountName := data.Name.ValueString()

	// Read account from ArgoCD
	diags := r.readAccount(ctx, &data, accountName)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *accountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state accountModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	// Initialize API clients
	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	if resp.Diagnostics.HasError() {
		return
	}

	accountName := data.Name.ValueString()

	// Check if password has changed
	if !data.Password.Equal(state.Password) {
		oldPassword := state.Password.ValueString()
		newPassword := data.Password.ValueString()

		if newPassword == "" {
			resp.Diagnostics.AddError("Password cannot be empty", "The password field cannot be set to an empty value.")
			return
		}

		tflog.Debug(ctx, fmt.Sprintf("updating password for account %s", accountName))

		updateReq := &account.UpdatePasswordRequest{
			Name:            accountName,
			CurrentPassword: oldPassword,
			NewPassword:     newPassword,
		}

		sync.AccountsMutex.Lock()
		_, err := r.si.AccountClient.UpdatePassword(ctx, updateReq)
		sync.AccountsMutex.Unlock()

		if err != nil {
			resp.Diagnostics.Append(diagnostics.ArgoCDAPIError("update", "account password", accountName, err)...)
			return
		}
	}

	// Read updated account from ArgoCD
	diags := r.readAccount(ctx, &data, accountName)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *accountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// ArgoCD accounts are typically managed through ArgoCD configuration
	// This resource doesn't actually delete accounts from ArgoCD
	// It only removes the account from Terraform state
	tflog.Trace(ctx, "account resource delete - removing from state only")
}

func (r *accountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Use the ID as the account name
	accountName := req.ID

	// Set both the ID and name attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), accountName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), accountName)...)
}

func (r *accountResource) readAccount(ctx context.Context, data *accountModel, accountName string) diag.Diagnostics {
	var diags diag.Diagnostics

	// Get account from ArgoCD
	accountResp, err := r.si.AccountClient.GetAccount(ctx, &account.GetAccountRequest{
		Name: accountName,
	})

	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			diags.AddError("Account not found", fmt.Sprintf("Account %s not found", accountName))
			return diags
		}
		diags.Append(diagnostics.ArgoCDAPIError("read", "account", accountName, err)...)
		return diags
	}

	if accountResp == nil {
		diags.AddError("Account not found", fmt.Sprintf("Account %s not found", accountName))
		return diags
	}

	// Update model with response data
	data.Name = types.StringValue(accountResp.Name)
	data.Enabled = types.BoolValue(accountResp.Enabled)

	// Convert capabilities to types.List
	capabilities := make([]attr.Value, len(accountResp.Capabilities))
	for i, cap := range accountResp.Capabilities {
		capabilities[i] = types.StringValue(cap)
	}
	capList, diag := types.ListValue(types.StringType, capabilities)
	if diag.HasError() {
		diags.Append(diag...)
		return diags
	}
	data.Capabilities = capList

	// Convert tokens to types.List
	tokenAttrTypes := map[string]attr.Type{
		"id":         types.StringType,
		"issued_at":  types.StringType,
		"expires_at": types.StringType,
	}

	tokens := make([]attr.Value, len(accountResp.Tokens))
	for i, token := range accountResp.Tokens {
		tokenAttrs := map[string]attr.Value{
			"id":         types.StringValue(token.Id),
			"issued_at":  types.StringValue(strconv.FormatInt(token.IssuedAt, 10)),
			"expires_at": types.StringValue(strconv.FormatInt(token.ExpiresAt, 10)),
		}
		tokenObj, tokenDiag := types.ObjectValue(tokenAttrTypes, tokenAttrs)
		if tokenDiag.HasError() {
			diags.Append(tokenDiag...)
			return diags
		}
		tokens[i] = tokenObj
	}

	tokenList, diag := types.ListValue(types.ObjectType{AttrTypes: tokenAttrTypes}, tokens)
	if diag.HasError() {
		diags.Append(diag...)
		return diags
	}
	data.Tokens = tokenList

	return diags
}

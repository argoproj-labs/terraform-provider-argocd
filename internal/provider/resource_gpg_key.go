package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/argoproj-labs/terraform-provider-argocd/internal/diagnostics"
	"github.com/argoproj-labs/terraform-provider-argocd/internal/sync"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/gpgkey"
	"github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &gpgKeyResource{}

func NewGPGKeyResource() resource.Resource {
	return &gpgKeyResource{}
}

// gpgKeyResource defines the resource implementation.
type gpgKeyResource struct {
	si *ServerInterface
}

func (r *gpgKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gpg_key"
}

func (r *gpgKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages [GPG keys](https://argo-cd.readthedocs.io/en/stable/user-guide/gpg-verification/) within ArgoCD.",
		Attributes:          gpgKeySchemaAttributes(),
	}
}

func (r *gpgKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *gpgKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data gpgKeyModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	// Initialize API clients
	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	// Check for errors before proceeding
	if resp.Diagnostics.HasError() {
		return
	}

	// Create GPG key
	sync.GPGKeysMutex.Lock()

	keys, err := r.si.GPGKeysClient.Create(ctx, &gpgkey.GnuPGPublicKeyCreateRequest{
		Publickey: &v1alpha1.GnuPGPublicKey{KeyData: data.PublicKey.String()},
	})

	sync.GPGKeysMutex.Unlock()

	if err != nil {
		resp.Diagnostics.Append(diagnostics.ArgoCDAPIError("create", "GPG key", "", err)...)
		return
	}

	if keys.Created == nil || len(keys.Created.Items) == 0 {
		resp.Diagnostics.AddError("unexpected response when creating ArgoCD GPG Key - no keys created", "")
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created GPG key %s", keys.Created.Items[0].KeyID))

	// Parse response and store state
	resp.Diagnostics.Append(resp.State.Set(ctx, newGPGKey(&keys.Created.Items[0]))...)
}

func (r *gpgKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data gpgKeyModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	// Initialize API clients
	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	// Check for errors before proceeding
	if resp.Diagnostics.HasError() {
		return
	}

	// Read key from API
	key, diags := readGPGKey(ctx, r.si, data.ID.ValueString())
	resp.Diagnostics.Append(diags...)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, key)...)
}

func (r *gpgKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data gpgKeyModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	// Initialize API clients
	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	// Check for errors before proceeding
	if resp.Diagnostics.HasError() {
		return
	}

	// In general, this resource will be recreated rather than updated. However,
	// `Update` will be called on the first apply after an import so we need to
	// ensure that we set the state of the computed data by reading the key from
	// the API.
	key, diags := readGPGKey(ctx, r.si, data.ID.ValueString())
	resp.Diagnostics.Append(diags...)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, key)...)
}

func (r *gpgKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data gpgKeyModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	// Initialize API clients
	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	// Check for errors before proceeding
	if resp.Diagnostics.HasError() {
		return
	}

	sync.GPGKeysMutex.Lock()

	_, err := r.si.GPGKeysClient.Delete(ctx, &gpgkey.GnuPGPublicKeyQuery{
		KeyID: data.ID.ValueString(),
	})

	sync.GPGKeysMutex.Unlock()

	if err != nil && !strings.Contains(err.Error(), "NotFound") {
		resp.Diagnostics.Append(diagnostics.ArgoCDAPIError("delete", "GPG key", data.ID.ValueString(), err)...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted GPG key %s", data.ID.ValueString()))
}

func (r *gpgKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func readGPGKey(ctx context.Context, si *ServerInterface, id string) (*gpgKeyModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	sync.GPGKeysMutex.RLock()

	k, err := si.GPGKeysClient.Get(ctx, &gpgkey.GnuPGPublicKeyQuery{
		KeyID: id,
	})

	sync.GPGKeysMutex.RUnlock()

	if err != nil {
		if !strings.Contains(err.Error(), "NotFound") {
			diags.Append(diagnostics.ArgoCDAPIError("read", "GPG key", id, err)...)
		}

		return nil, diags
	}

	return newGPGKey(k), diags
}

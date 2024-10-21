package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/argoproj-labs/terraform-provider-argocd/internal/diagnostics"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &applicationDataSource{}

func NewArgoCDApplicationDataSource() datasource.DataSource {
	return &applicationDataSource{}
}

// applicationDataSource defines the data source implementation.
type applicationDataSource struct {
	si *ServerInterface
}

func (d *applicationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (d *applicationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing ArgoCD application.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ArgoCD application identifier",
				Computed:            true,
			},
			"metadata": objectMetaSchemaAttribute("applications.argoproj.io", true),
			"spec":     applicationSpecSchemaAttribute(true, true),
			"status":   applicationStatusSchemaAttribute(),
		},
	}
}

func (d *applicationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	si, ok := req.ProviderData.(*ServerInterface)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data",
			fmt.Sprintf("Expected *ServerInterface, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.si = si
}

func (d *applicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data applicationModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	// Initialize API clients
	resp.Diagnostics.Append(d.si.InitClients(ctx)...)

	// Check for errors before proceeding
	if resp.Diagnostics.HasError() {
		return
	}

	id := fmt.Sprintf("%s:%s", data.Metadata.Name.ValueString(), data.Metadata.Namespace.ValueString())
	data.ID = types.StringValue(id)

	// Read application
	resp.Diagnostics.Append(readApplication(ctx, d.si, &data)...)

	tflog.Trace(ctx, "read ArgoCD application")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func readApplication(ctx context.Context, si *ServerInterface, data *applicationModel) (diags diag.Diagnostics) {
	ids := strings.Split(data.ID.ValueString(), ":")
	appName := ids[0]
	namespace := ids[1]

	apps, err := si.ApplicationClient.List(ctx, &application.ApplicationQuery{
		Name:         &appName,
		AppNamespace: &namespace,
	})
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			data.ID = types.StringUnknown()
			return diags
		}

		diags.Append(diagnostics.ArgoCDAPIError("read", "application", appName, err)...)

		return diags
	}

	l := len(apps.Items)

	switch {
	case l < 1:
		data.ID = types.StringUnknown()
		return diags
	case l == 1:
		break
	case l > 1:
		diags.AddError(fmt.Sprintf("found multiple applications matching name '%s' and namespace '%s'", appName, namespace), "")
		return diags
	}

	app := apps.Items[0]

	data.Metadata = newObjectMeta(app.ObjectMeta)
	data.Spec = newApplicationSpec(app.Spec)
	data.Status = newApplicationStatus(app.Status)

	return diags
}

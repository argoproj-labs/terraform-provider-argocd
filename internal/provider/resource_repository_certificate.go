package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/argoproj-labs/terraform-provider-argocd/internal/diagnostics"
	"github.com/argoproj-labs/terraform-provider-argocd/internal/sync"
	"github.com/argoproj-labs/terraform-provider-argocd/internal/validators"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/certificate"
	"github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const sshCertType = "ssh"

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &repositoryCertificateResource{}
var _ resource.ResourceWithImportState = &repositoryCertificateResource{}
var _ resource.ResourceWithConfigValidators = &repositoryCertificateResource{}

func NewRepositoryCertificateResource() resource.Resource {
	return &repositoryCertificateResource{}
}

// repositoryCertificateResource defines the resource implementation.
type repositoryCertificateResource struct {
	si *ServerInterface
}

func (r *repositoryCertificateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository_certificate"
}

func (r *repositoryCertificateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages [custom TLS certificates](https://argo-cd.readthedocs.io/en/stable/user-guide/private-repositories/#self-signed-untrusted-tls-certificates) used by ArgoCD for connecting Git repositories.",
		Attributes:          repositoryCertificateSchemaAttributes(),
		Blocks:              repositoryCertificateSchemaBlocks(),
	}
}

func (r *repositoryCertificateResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		validators.RepositoryCertificate(),
	}
}

func (r *repositoryCertificateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *repositoryCertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data repositoryCertificateModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	// Initialize API clients
	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	// Check for errors before proceeding
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to API model
	cert := data.toAPIModel()

	// Check if HTTPS certificate already exists
	if cert.CertType == "https" {
		sync.CertificateMutex.Lock()
		existing, err := r.si.CertificateClient.ListCertificates(ctx, &certificate.RepositoryCertificateQuery{
			HostNamePattern: cert.ServerName,
			CertType:        cert.CertType,
			CertSubType:     cert.CertSubType,
		})
		sync.CertificateMutex.Unlock()

		if err != nil {
			resp.Diagnostics.Append(diagnostics.ArgoCDAPIError("list", "repository certificates", cert.ServerName, err)...)
			return
		}

		if len(existing.Items) > 0 {
			resp.Diagnostics.AddError(
				"Repository certificate already exists",
				fmt.Sprintf("https certificate for '%s' already exists", cert.ServerName),
			)

			return
		}
	}

	// Create certificate
	certs := v1alpha1.RepositoryCertificateList{
		Items: []v1alpha1.RepositoryCertificate{*cert},
	}

	sync.CertificateMutex.Lock()
	createdCerts, err := r.si.CertificateClient.CreateCertificate(
		ctx,
		&certificate.RepositoryCertificateCreateRequest{
			Certificates: &certs,
			Upsert:       false,
		},
	)
	sync.CertificateMutex.Unlock()

	if err != nil {
		resp.Diagnostics.Append(diagnostics.ArgoCDAPIError("create", "repository certificate", cert.ServerName, err)...)
		return
	}

	// Handle the response - use the created certificate or the original if empty
	var resultCert *v1alpha1.RepositoryCertificate
	if len(createdCerts.Items) > 0 {
		resultCert = &createdCerts.Items[0]
	} else {
		resultCert = cert
	}

	// Generate ID and update model
	data.ID = types.StringValue(data.generateID())

	// Update the model with the created certificate data
	result := data // Start with the original data to preserve all fields
	result.ID = types.StringValue(data.generateID())

	// Update computed fields from API response
	if len(data.SSH) > 0 && resultCert.CertType == sshCertType {
		result.SSH[0].CertInfo = types.StringValue(resultCert.CertInfo)
	}

	if len(data.HTTPS) > 0 && resultCert.CertType == "https" {
		result.HTTPS[0].CertInfo = types.StringValue(resultCert.CertInfo)
		result.HTTPS[0].CertSubType = types.StringValue(resultCert.CertSubType)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)

	tflog.Trace(ctx, fmt.Sprintf("created repository certificate %s", data.ID.ValueString()))
}

func (r *repositoryCertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data repositoryCertificateModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	// Initialize API clients
	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	// Check for errors before proceeding
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse certificate ID to get query parameters
	certType, certSubType, serverName, err := r.parseID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse certificate ID", err.Error())
		return
	}

	// Read certificate from API
	cert, diags := r.readCertificate(ctx, certType, certSubType, serverName)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If certificate was not found, remove from state
	if cert == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update the model with the read certificate data
	result := data // Start with the original data to preserve all fields
	result.ID = data.ID

	// Update computed fields from API response
	if len(data.SSH) > 0 && cert.CertType == sshCertType {
		result.SSH[0].CertInfo = types.StringValue(cert.CertInfo)
	}

	if len(data.HTTPS) > 0 && cert.CertType == "https" {
		result.HTTPS[0].CertInfo = types.StringValue(cert.CertInfo)
		result.HTTPS[0].CertSubType = types.StringValue(cert.CertSubType)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *repositoryCertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Repository certificates don't support updates - all attributes are ForceNew
	resp.Diagnostics.AddError(
		"Repository certificates cannot be updated",
		"Repository certificates are immutable. To change a certificate, it must be deleted and recreated.",
	)
}

func (r *repositoryCertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data repositoryCertificateModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	// Initialize API clients
	resp.Diagnostics.Append(r.si.InitClients(ctx)...)

	// Check for errors before proceeding
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse certificate ID to get query parameters
	certType, certSubType, serverName, err := r.parseID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse certificate ID", err.Error())
		return
	}

	// Delete certificate
	query := certificate.RepositoryCertificateQuery{
		HostNamePattern: serverName,
		CertType:        certType,
		CertSubType:     certSubType,
	}

	sync.CertificateMutex.Lock()
	_, err = r.si.CertificateClient.DeleteCertificate(ctx, &query)
	sync.CertificateMutex.Unlock()

	if err != nil {
		if !strings.Contains(err.Error(), "NotFound") {
			resp.Diagnostics.Append(diagnostics.ArgoCDAPIError("delete", "repository certificate", serverName, err)...)
			return
		}
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted repository certificate %s", data.ID.ValueString()))
}

func (r *repositoryCertificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *repositoryCertificateResource) readCertificate(ctx context.Context, certType, certSubType, serverName string) (*v1alpha1.RepositoryCertificate, diag.Diagnostics) {
	var diags diag.Diagnostics

	sync.CertificateMutex.RLock()
	defer sync.CertificateMutex.RUnlock()

	certs, err := r.si.CertificateClient.ListCertificates(ctx, &certificate.RepositoryCertificateQuery{
		HostNamePattern: serverName,
		CertType:        certType,
		CertSubType:     certSubType,
	})

	if err != nil {
		diags.Append(diagnostics.ArgoCDAPIError("read", "repository certificate", serverName, err)...)
		return nil, diags
	}

	if certs == nil || len(certs.Items) == 0 {
		// Certificate has been deleted out-of-band
		return nil, diags
	}

	// Find the specific certificate by generating its ID
	targetID := r.generateID(certType, certSubType, serverName)

	for _, cert := range certs.Items {
		certID := r.generateIDFromCert(&cert)
		if certID == targetID {
			return &cert, diags
		}
	}

	// Certificate not found
	return nil, diags
}

func (r *repositoryCertificateResource) parseID(id string) (certType, certSubType, serverName string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) < 2 {
		return "", "", "", fmt.Errorf("invalid certificate ID format: %s", id)
	}

	certType = parts[0]
	switch certType {
	case sshCertType:
		if len(parts) < 3 {
			return "", "", "", fmt.Errorf("invalid SSH certificate ID format: %s", id)
		}

		return parts[0], parts[1], parts[2], nil
	case "https":
		if len(parts) < 2 {
			return "", "", "", fmt.Errorf("invalid HTTPS certificate ID format: %s", id)
		}

		return parts[0], "", parts[1], nil
	default:
		return "", "", "", fmt.Errorf("unknown certificate type: %s", certType)
	}
}

func (r *repositoryCertificateResource) generateID(certType, certSubType, serverName string) string {
	if certType == sshCertType {
		return fmt.Sprintf("%s/%s/%s", certType, certSubType, serverName)
	}

	return fmt.Sprintf("%s/%s", certType, serverName)
}

func (r *repositoryCertificateResource) generateIDFromCert(cert *v1alpha1.RepositoryCertificate) string {
	if cert.CertType == sshCertType {
		return fmt.Sprintf("%s/%s/%s", cert.CertType, cert.CertSubType, cert.ServerName)
	}

	return fmt.Sprintf("%s/%s", cert.CertType, cert.ServerName)
}

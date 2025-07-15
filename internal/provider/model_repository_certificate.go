package provider

import (
	"github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type repositoryCertificateModel struct {
	ID    types.String                     `tfsdk:"id"`
	SSH   *repositoryCertificateSSHModel   `tfsdk:"ssh"`
	HTTPS *repositoryCertificateHTTPSModel `tfsdk:"https"`
}

type repositoryCertificateSSHModel struct {
	ServerName  types.String `tfsdk:"server_name"`
	CertSubType types.String `tfsdk:"cert_subtype"`
	CertData    types.String `tfsdk:"cert_data"`
	CertInfo    types.String `tfsdk:"cert_info"`
}

type repositoryCertificateHTTPSModel struct {
	ServerName  types.String `tfsdk:"server_name"`
	CertData    types.String `tfsdk:"cert_data"`
	CertSubType types.String `tfsdk:"cert_subtype"`
	CertInfo    types.String `tfsdk:"cert_info"`
}

func repositoryCertificateSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "Repository certificate identifier",
			Computed:            true,
		},
	}
}

func repositoryCertificateSchemaBlocks() map[string]schema.Block {
	return map[string]schema.Block{
		"ssh": schema.SingleNestedBlock{
			MarkdownDescription: "SSH certificate configuration",
			Attributes: map[string]schema.Attribute{
				"server_name": schema.StringAttribute{
					MarkdownDescription: "DNS name of the server this certificate is intended for",
					Optional:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"cert_subtype": schema.StringAttribute{
					MarkdownDescription: "The sub type of the cert, i.e. `ssh-rsa`",
					Optional:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"cert_data": schema.StringAttribute{
					MarkdownDescription: "The actual certificate data, dependent on the certificate type",
					Optional:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"cert_info": schema.StringAttribute{
					MarkdownDescription: "Additional certificate info, dependent on the certificate type (e.g. SSH fingerprint, X509 CommonName)",
					Computed:            true,
				},
			},
		},
		"https": schema.SingleNestedBlock{
			MarkdownDescription: "HTTPS certificate configuration",
			Attributes: map[string]schema.Attribute{
				"server_name": schema.StringAttribute{
					MarkdownDescription: "DNS name of the server this certificate is intended for",
					Optional:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"cert_data": schema.StringAttribute{
					MarkdownDescription: "The actual certificate data, dependent on the certificate type",
					Optional:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"cert_subtype": schema.StringAttribute{
					MarkdownDescription: "The sub type of the cert, i.e. `ssh-rsa`",
					Computed:            true,
				},
				"cert_info": schema.StringAttribute{
					MarkdownDescription: "Additional certificate info, dependent on the certificate type (e.g. SSH fingerprint, X509 CommonName)",
					Computed:            true,
				},
			},
		},
	}
}

func (m *repositoryCertificateModel) toAPIModel() *v1alpha1.RepositoryCertificate {
	cert := &v1alpha1.RepositoryCertificate{}

	if m.SSH != nil {
		cert.CertType = "ssh"
		cert.ServerName = m.SSH.ServerName.ValueString()
		cert.CertSubType = m.SSH.CertSubType.ValueString()
		cert.CertData = []byte(m.SSH.CertData.ValueString())
	} else if m.HTTPS != nil {
		cert.CertType = "https"
		cert.ServerName = m.HTTPS.ServerName.ValueString()
		cert.CertData = []byte(m.HTTPS.CertData.ValueString())
	}

	return cert
}

func newRepositoryCertificateModel(cert *v1alpha1.RepositoryCertificate) *repositoryCertificateModel {
	model := &repositoryCertificateModel{}

	// Generate ID based on certificate type
	switch cert.CertType {
	case "ssh":
		model.ID = types.StringValue(cert.CertType + "/" + cert.CertSubType + "/" + cert.ServerName)
		model.SSH = &repositoryCertificateSSHModel{
			ServerName:  types.StringValue(cert.ServerName),
			CertSubType: types.StringValue(cert.CertSubType),
			CertData:    types.StringValue(string(cert.CertData)),
			CertInfo:    types.StringValue(cert.CertInfo),
		}
	case "https":
		model.ID = types.StringValue(cert.CertType + "/" + cert.ServerName)
		model.HTTPS = &repositoryCertificateHTTPSModel{
			ServerName:  types.StringValue(cert.ServerName),
			CertData:    types.StringValue(string(cert.CertData)),
			CertSubType: types.StringValue(cert.CertSubType),
			CertInfo:    types.StringValue(cert.CertInfo),
		}
	}

	return model
}

func (m *repositoryCertificateModel) generateID() string {
	if m.SSH != nil {
		return "ssh/" + m.SSH.CertSubType.ValueString() + "/" + m.SSH.ServerName.ValueString()
	} else if m.HTTPS != nil {
		return "https/" + m.HTTPS.ServerName.ValueString()
	}

	return ""
}

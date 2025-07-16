package provider

import (
	"github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type repositoryCertificateModel struct {
	ID    types.String                      `tfsdk:"id"`
	SSH   []repositoryCertificateSSHModel   `tfsdk:"ssh"`
	HTTPS []repositoryCertificateHTTPSModel `tfsdk:"https"`
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
		"ssh": schema.ListNestedBlock{
			MarkdownDescription: "SSH certificate configuration",
			NestedObject: schema.NestedBlockObject{
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
		},
		"https": schema.ListNestedBlock{
			MarkdownDescription: "HTTPS certificate configuration",
			NestedObject: schema.NestedBlockObject{
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
		},
	}
}

func (m *repositoryCertificateModel) toAPIModel() *v1alpha1.RepositoryCertificate {
	cert := &v1alpha1.RepositoryCertificate{}

	if len(m.SSH) > 0 {
		ssh := m.SSH[0]
		cert.CertType = "ssh"
		cert.ServerName = ssh.ServerName.ValueString()
		cert.CertSubType = ssh.CertSubType.ValueString()
		cert.CertData = []byte(ssh.CertData.ValueString())
	} else if len(m.HTTPS) > 0 {
		https := m.HTTPS[0]
		cert.CertType = "https"
		cert.ServerName = https.ServerName.ValueString()
		cert.CertData = []byte(https.CertData.ValueString())
	}

	return cert
}

func (m *repositoryCertificateModel) generateID() string {
	if len(m.SSH) > 0 {
		ssh := m.SSH[0]
		return "ssh/" + ssh.CertSubType.ValueString() + "/" + ssh.ServerName.ValueString()
	} else if len(m.HTTPS) > 0 {
		https := m.HTTPS[0]
		return "https/" + https.ServerName.ValueString()
	}

	return ""
}

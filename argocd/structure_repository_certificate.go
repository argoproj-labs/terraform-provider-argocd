package argocd

import (
	"context"

	application "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandRepositoryCertificate(d *schema.ResourceData) *application.RepositoryCertificate {
	certificate := &application.RepositoryCertificate{}

	if _, ok := d.GetOk("ssh"); ok {
		certificate.CertType = "ssh"
		if v, ok := d.GetOk("ssh.0.server_name"); ok {
			certificate.ServerName = v.(string)
		}

		if v, ok := d.GetOk("ssh.0.cert_subtype"); ok {
			certificate.CertSubType = v.(string)
		}

		if v, ok := d.GetOk("ssh.0.cert_data"); ok {
			certificate.CertData = []byte(v.(string))
		}
	} else if _, ok := d.GetOk("https"); ok {
		certificate.CertType = "https"
		if v, ok := d.GetOk("https.0.server_name"); ok {
			certificate.ServerName = v.(string)
		}

		if v, ok := d.GetOk("https.0.cert_data"); ok {
			certificate.CertData = []byte(v.(string))
		}
	}

	return certificate
}

func flattenRepositoryCertificate(certificate *application.RepositoryCertificate, d *schema.ResourceData, ctx context.Context) error {
	var r map[string]interface{}

	if certificate.CertType == "ssh" {
		r = map[string]interface{}{
			"ssh": []map[string]string{
				{
					"server_name":  certificate.ServerName,
					"cert_subtype": certificate.CertSubType,
					"cert_info":    certificate.CertInfo,
					"cert_data":    getCertDataFromContextOrState(ctx, d, "ssh.0.cert_data"),
				},
			},
		}
	} else if certificate.CertType == "https" {
		r = map[string]interface{}{
			"https": []map[string]string{
				{
					"server_name":  certificate.ServerName,
					"cert_subtype": certificate.CertSubType,
					"cert_info":    certificate.CertInfo,
					"cert_data":    getCertDataFromContextOrState(ctx, d, "https.0.cert_data"),
				},
			},
		}
	}

	for k, v := range r {
		if err := persistToState(k, v, d); err != nil {
			return err
		}
	}

	return nil
}

// Since ArgoCD API does not return sensitive data, to avoid data drift :
// get cert_data from context if it has just been created
// or from current state if we are in a subsequent "Read"
func getCertDataFromContextOrState(ctx context.Context, d *schema.ResourceData, statePath string) (certData string) {
	if v, ok := ctx.Value("cert_data").([]byte); ok {
		certData = string(v)
	} else {
		certData = d.Get(statePath).(string)
	}

	return
}

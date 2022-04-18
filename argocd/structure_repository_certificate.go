package argocd

import (
	"fmt"

	application "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Expand

func expandRepositoryCertificate(d *schema.ResourceData) (*application.RepositoryCertificate, error) {
	certificate := &application.RepositoryCertificate{}
	if v, ok := d.GetOk("server_name"); ok {
		certificate.ServerName = v.(string)
	}
	if v, ok := d.GetOk("cert_type"); ok {
		certType := v.(string)
		if certType != "ssh" && certType != "https" {
			return nil, fmt.Errorf("invalid certificate type '%s': must be either ssh or https", certType)
		}
		certificate.CertType = v.(string)
	}
	if v, ok := d.GetOk("cert_subtype"); ok {
		certificate.CertSubType = v.(string)
	}
	if v, ok := d.GetOk("cert_data"); ok {
		certificate.CertData = []byte(v.(string))
	}
	if v, ok := d.GetOk("cert_info"); ok {
		certificate.CertInfo = v.(string)
	}
	return certificate, nil
}

// Flatten

func flattenRepositoryCertificate(certificate *application.RepositoryCertificate, d *schema.ResourceData) error {
	r := map[string]interface{}{
		"server_name":  certificate.ServerName,
		"cert_type":    certificate.CertType,
		"cert_subtype": certificate.CertSubType,
		"cert_info":    certificate.CertInfo,
		// "cert_data": certificate.CertData,  ArgoCD API does not return sensitive data!
	}
	for k, v := range r {
		if err := persistToState(k, v, d); err != nil {
			return err
		}
	}
	return nil
}

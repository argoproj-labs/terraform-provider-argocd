package argocd

import (
	"fmt"

	application "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Expand

func expandRepositoryCredentials(d *schema.ResourceData) *application.RepoCreds {
	repoCreds := &application.RepoCreds{}
	if v, ok := d.GetOk("url"); ok {
		repoCreds.URL = v.(string)
	}
	if v, ok := d.GetOk("username"); ok {
		repoCreds.Username = v.(string)
	}
	if v, ok := d.GetOk("password"); ok {
		repoCreds.Password = v.(string)
	}
	if v, ok := d.GetOk("ssh_private_key"); ok {
		repoCreds.SSHPrivateKey = v.(string)
	}
	if v, ok := d.GetOk("tls_client_cert_data"); ok {
		repoCreds.TLSClientCertData = v.(string)
	}
	if v, ok := d.GetOk("tls_client_cert_key"); ok {
		repoCreds.TLSClientCertKey = v.(string)
	}
	return repoCreds
}

// Flatten

func flattenRepositoryCredentials(repository application.RepoCreds, d *schema.ResourceData) diag.Diagnostics {
	r := map[string]interface{}{
		"url":      repository.URL,
		"username": repository.Username,
		// TODO: ArgoCD API does not return sensitive data!
		//"password":             repository.Password,
		//"ssh_private_key":      repository.SSHPrivateKey,
		//"tls_client_cert_key":  repository.TLSClientCertKey,
		"tls_client_cert_data": repository.TLSClientCertData,
	}
	for k, v := range r {
		if err := persistToState(k, v, d); err != nil {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("credentials key (%s) and value for repository %s could not be persisted to state", k, repository.URL),
					Detail:   err.Error(),
				},
			}
		}
	}
	return nil
}

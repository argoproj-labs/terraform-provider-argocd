package argocd

import (
	application "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Expand

func expandRepository(d *schema.ResourceData) *application.Repository {
	repository := &application.Repository{}
	if v, ok := d.GetOk("repo"); ok {
		repository.Repo = v.(string)
	}
	if v, ok := d.GetOk("enable_lfs"); ok {
		repository.EnableLFS = v.(bool)
	}
	if v, ok := d.GetOk("inherited_creds"); ok {
		repository.InheritedCreds = v.(bool)
	}
	if v, ok := d.GetOk("insecure"); ok {
		repository.Insecure = v.(bool)
	}
	if v, ok := d.GetOk("name"); ok {
		repository.Name = v.(string)
	}
	if v, ok := d.GetOk("username"); ok {
		repository.Username = v.(string)
	}
	if v, ok := d.GetOk("password"); ok {
		repository.Password = v.(string)
	}
	if v, ok := d.GetOk("ssh_private_key"); ok {
		repository.SSHPrivateKey = v.(string)
	}
	if v, ok := d.GetOk("tls_client_cert_data"); ok {
		repository.TLSClientCertData = v.(string)
	}
	if v, ok := d.GetOk("tls_client_cert_key"); ok {
		repository.TLSClientCertKey = v.(string)
	}
	if v, ok := d.GetOk("type"); ok {
		repository.Type = v.(string)
	}
	return repository
}

// Flatten

func flattenRepository(repository *application.Repository, d *schema.ResourceData) error {
	r := map[string]interface{}{
		"repo":                    repository.Repo,
		"connection_state_status": repository.ConnectionState.Status,
		"enable_lfs":              repository.EnableLFS,
		"inherited_creds":         repository.InheritedCreds,
		"insecure":                repository.Insecure,
		"name":                    repository.Name,
		// TODO: in case of repositoryCredentials existence, will perma-diff
		//"username":                repository.Username,
		// TODO: ArgoCD API does not return sensitive data!
		//"password":                repository.Password,
		//"ssh_private_key":         repository.SSHPrivateKey,
		//"tls_client_cert_key":     repository.TLSClientCertKey,
		"tls_client_cert_data": repository.TLSClientCertData,
		"type":                 repository.Type,
	}
	for k, v := range r {
		if err := persistToState(k, v, d); err != nil {
			return err
		}
	}
	return nil
}

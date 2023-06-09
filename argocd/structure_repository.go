package argocd

import (
	application "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandRepository(d *schema.ResourceData) (*application.Repository, error) {
	var err error

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

	if v, ok := d.GetOk("project"); ok {
		repository.Project = v.(string)
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

	if v, ok := d.GetOk("enable_oci"); ok {
		repository.EnableOCI = v.(bool)
	}

	if v, ok := d.GetOk("type"); ok {
		repository.Type = v.(string)
	}

	if v, ok := d.GetOk("githubapp_id"); ok {
		repository.GithubAppId, err = convertStringToInt64(v.(string))
		if err != nil {
			return nil, err
		}
	}

	if v, ok := d.GetOk("githubapp_installation_id"); ok {
		repository.GithubAppInstallationId, err = convertStringToInt64(v.(string))
		if err != nil {
			return nil, err
		}
	}

	if v, ok := d.GetOk("githubapp_enterprise_base_url"); ok {
		repository.GitHubAppEnterpriseBaseURL = v.(string)
	}

	if v, ok := d.GetOk("githubapp_private_key"); ok {
		repository.GithubAppPrivateKey = v.(string)
	}

	return repository, nil
}

func flattenRepository(repository *application.Repository, d *schema.ResourceData) error {
	r := map[string]interface{}{
		"repo":                    repository.Repo,
		"connection_state_status": repository.ConnectionState.Status,
		"enable_lfs":              repository.EnableLFS,
		"inherited_creds":         repository.InheritedCreds,
		"insecure":                repository.Insecure,
		"name":                    repository.Name,
		"project":                 repository.Project,
		"type":                    repository.Type,

		// ArgoCD API does not return sensitive data so we can't track the state of these attributes.
		// "password":              repository.Password,
		// "ssh_private_key":       repository.SSHPrivateKey,
		// "tls_client_cert_key":   repository.TLSClientCertKey,
		// "githubapp_private_key": repository.GithubAppPrivateKey,
	}

	if !repository.InheritedCreds {
		// To prevent perma-diff in case of existence of repository credentials
		// existence, we only track the state of these values when the
		// repository is not inheriting credentials
		r["githubapp_enterprise_base_url"] = repository.GitHubAppEnterpriseBaseURL
		r["tls_client_cert_data"] = repository.TLSClientCertData
		r["username"] = repository.Username

		if repository.GithubAppId > 0 {
			r["githubapp_id"] = convertInt64ToString(repository.GithubAppId)
		}

		if repository.GithubAppInstallationId > 0 {
			r["githubapp_installation_id"] = convertInt64ToString(repository.GithubAppInstallationId)
		}
	}

	for k, v := range r {
		if err := persistToState(k, v, d); err != nil {
			return err
		}
	}

	return nil
}

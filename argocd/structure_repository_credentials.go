package argocd

import (
	"fmt"

	application "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandRepositoryCredentials(d *schema.ResourceData) (*application.RepoCreds, error) {
	var err error

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

	if v, ok := d.GetOk("enable_oci"); ok {
		repoCreds.EnableOCI = v.(bool)
	}

	if v, ok := d.GetOk("githubapp_id"); ok {
		repoCreds.GithubAppId, err = convertStringToInt64(v.(string))
		if err != nil {
			return nil, err
		}
	}

	if v, ok := d.GetOk("githubapp_installation_id"); ok {
		repoCreds.GithubAppInstallationId, err = convertStringToInt64(v.(string))
		if err != nil {
			return nil, err
		}
	}

	if v, ok := d.GetOk("githubapp_enterprise_base_url"); ok {
		repoCreds.GitHubAppEnterpriseBaseURL = v.(string)
	}

	if v, ok := d.GetOk("githubapp_private_key"); ok {
		repoCreds.GithubAppPrivateKey = v.(string)
	}

	return repoCreds, nil
}

func flattenRepositoryCredentials(repoCreds application.RepoCreds, d *schema.ResourceData) diag.Diagnostics {
	r := map[string]interface{}{
		"url":      repoCreds.URL,
		"username": repoCreds.Username,
	}

	// Note: We are only able to retrieve URL and Username from the ArgoCD API
	// at this point since `repocreds` does not implement a `Get` endpoint which
	// would alow us to retrieve additional details. See
	// https://github.com/argoproj/argo-cd/blob/7be094f38d06859b594b98eb75c7c70d39b80b1e/server/repocreds/repocreds.go#L58-L61

	for k, v := range r {
		if err := persistToState(k, v, d); err != nil {
			return errorToDiagnostics(fmt.Sprintf("credentials key (%s) and value for repository %s could not be persisted to state", k, repoCreds.URL), err)
		}
	}

	return nil
}

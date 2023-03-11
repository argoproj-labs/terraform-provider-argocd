# argocd_repository_credentials

Creates ArgoCD repository credentials, for use with future or existing private repositories.

**Note**: due to restrictions in the ArgoCD API the provider is unable to track
drift in this resource to fields other than `username`. I.e. the provider is
unable to detect changes to repository credentials that are made outside of
Terraform (e.g. manual updates to the underlying Kubernetes Secrets). 

## Example Usage

```hcl
// Private repository credentials
resource "argocd_repository_credentials" "private" {
  url             = "git@private-git-repository.local"
  username        = "git"
  ssh_private_key = "-----BEGIN OPENSSH PRIVATE KEY-----\nfoo\nbar\n-----END OPENSSH PRIVATE KEY-----"
}

// Uses previously defined argocd_repository_credentials credentials
resource "argocd_repository" "private" {
  repo = "git@private-git-repository.local:somerepo.git"
}
```

## Argument Reference

* `url` - (Required), string, URL that these credentials matches to.
* `username` - (Optional), string, username to authenticate against the repository server.
* `password` - (Optional), string, password to authenticate against the repository server.
* `ssh_private_key` - (Optional), string, SSH private key data to authenticate against the repository server. **Only for Git repositories**.
* `tls_client_cert_data` - (Optional), TLS client cert data to authenticate against the repository server.
* `tls_client_cert_key` - (Optional), TLS client cert key to authenticate against the repository server.
* `enable_oci` - (Optional), specify whether the repository server should be viewed as OCI compliant.
* `githubapp_id` - (Optional), string, GitHub App id for authenticating at the repo server only for GitHub repos
* `githubapp_installation_id` - (Optional), string, GitHub App installation id for authenticating at the repo server only for GitHub repos
* `githubapp_enterprise_base_url` - (Optional), string, If using GitHub App for a GitHub Enterprise repository the host url is required
* `githubapp_private_key` - (Optional), string, SSH private key data for GitHub App authentication.

## Import

ArgoCD repository credentials can be imported using an id consisting of `{url}`, e.g.
```
$ terraform import argocd_repository_credentials.myrepocreds git@private-git-repository.local:somerepo.git
```

**NOTE**: as ArgoCD API does not return any sensitive information, a subsequent _terraform apply_ should be executed to make the password, ssh_private_key and tls_client_cert_key attributes converge to their expected values defined within the plan.
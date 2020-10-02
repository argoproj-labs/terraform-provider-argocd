# argocd_repository

Creates an ArgoCD repository.

## Example Usage

```hcl
// Helm repository
resource "argocd_repository" "public_nginx_helm" {
  repo = "https://helm.nginx.com/stable"
  name = "nginx-stable"
  type = "helm"
}

// Public Git repository
resource "argocd_repository" "public_git" {
  repo = "git@github.com:user/somerepo.git"
}

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

* `repo` - (Required), string, URL of the repository.
* `type` - (Optional), string, type of the repo, may be "git or "helm. Defaults to `git`.
* `insecure` - (Optional), boolean, whether to verify the repository TLS certificate.
* `name` - (Optional), string, only for Helm repositories.
* `enable_lfs` - (Optional), boolean, whether git-lfs support should be enabled for this repository.
* `username` - (Optional), string, username to authenticate against the repository server.
* `password` - (Optional), string, password to authenticate against the repository server.
* `ssh_private_key` - (Optional), string, SSH private key data to authenticate against the repository server. **Only for Git repositories**.
* `tls_client_cert_data` - (Optional), TLS client cert data to authenticate against the repository server.
* `tls_client_cert_key` - (Optional), TLS client cert key to authenticate against the repository server.

# Exported Attributes

* `connection_state_status` - string, repository connection state status.
* `inherited_creds` - boolean, whether credentials wre inherited fron a credential set.

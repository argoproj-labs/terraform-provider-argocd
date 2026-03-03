# Public Helm repository
resource "argocd_repository" "public_nginx_helm" {
  repo = "https://helm.nginx.com/stable"
  name = "nginx-stable"
  type = "helm"
}

# Public Git repository
resource "argocd_repository" "public_git" {
  repo = "git@github.com:user/somerepo.git"
}

# Private Git repository
resource "argocd_repository" "private" {
  repo            = "git@private-git-repository.local:somerepo.git"
  username        = "git"
  ssh_private_key = "-----BEGIN OPENSSH PRIVATE KEY-----\nfoo\nbar\n-----END OPENSSH PRIVATE KEY-----"
  insecure        = true
}

# Repository with proxy configuration
resource "argocd_repository" "with_proxy" {
  repo     = "https://github.com/example/repo.git"
  username = "git"
  password = "my-token"
  proxy    = "http://proxy.example.com:8080"
  no_proxy = "*.internal.example.com,localhost"
}

# OCI repository (e.g., for Helm charts stored in OCI registries)
resource "argocd_repository" "oci_registry" {
  repo     = "oci://ghcr.io/argoproj/argo-helm/argo-cd"
  name     = "argocd-oci"
  type     = "oci"
  username = "my-username"
  password = "my-token"
}

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

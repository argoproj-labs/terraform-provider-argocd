resource "argocd_repository_credentials" "private" {
  url             = "git@private-git-repository.local"
  username        = "git"
  ssh_private_key = "-----BEGIN OPENSSH PRIVATE KEY-----\nfoo\nbar\n-----END OPENSSH PRIVATE KEY-----"
}

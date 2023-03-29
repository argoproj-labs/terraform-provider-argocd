resource "argocd_repository_credentials" "this" {
  url      = "https://github.com/foo"
  username = "foo"
  password = "bar"
}

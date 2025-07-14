data "argocd_application" "foo" {
  metadata = {
    name      = "foo"
    namespace = "argocd"
  }
}

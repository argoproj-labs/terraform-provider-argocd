resource "argocd_project_token" "long" {
  project     = argocd_project.foo.metadata.0.name
  role        = "foo"
  description = "long lived token"
}

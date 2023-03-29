resource "argocd_project_token" "long" {
  project     = argocd_project.foo.metadata.0.name
  role        = "foo"
  description = "long lived token"
}

resource "argocd_project_token" "renew_before" {
  project      = argocd_project.foo.metadata.0.name
  role         = "foo"
  description  = "auto-renewing short lived token"
  expires_in   = "24h"
  renew_before = "10h"
}

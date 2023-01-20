resource "argocd_project_token" "secret" {
  project      = "someproject"
  role         = "foobar"
  description  = "short lived token"
  expires_in   = "1h"
  renew_before = "30m"
}

resource "argocd_account_token" "admin" {
  renew_after = "30s"
}

resource "argocd_account_token" "test" {
  account      = "test"
  expires_in   = "1m"
  renew_before = "45s"
}

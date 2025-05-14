# Token for account configured on the `provider`
resource "argocd_account_token" "this" {
  renew_after = "168h" # renew after 7 days
}

# Token for ac count `foo`
resource "argocd_account_token" "foo" {
  account      = "foo"
  expires_in   = "168h" # expire in 7 days
  renew_before = "84h"  # renew when less than 3.5 days remain until expiry
}

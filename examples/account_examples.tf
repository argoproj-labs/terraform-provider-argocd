# Example: Managing ArgoCD Accounts

# Resource: argocd_account
# Manages an ArgoCD account and allows password updates
resource "argocd_account" "example" {
  name = "example-user"

  # Optional: Set password (when changed, uses previous state as current password)
  password = "secure-password"
}

# Resource: argocd_account_token (existing)
# Creates a token for an account
resource "argocd_account_token" "example" {
  account    = argocd_account.example.name
  expires_in = "24h"
}

# Data Source: argocd_account
# Retrieves a specific account
data "argocd_account" "admin" {
  name = "admin"
}

# Use the account data
output "admin_account_info" {
  value = {
    name         = data.argocd_account.admin.name
    enabled      = data.argocd_account.admin.enabled
    capabilities = data.argocd_account.admin.capabilities
    tokens       = data.argocd_account.admin.tokens
  }
}

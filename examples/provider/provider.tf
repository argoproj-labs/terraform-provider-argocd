# Exposed ArgoCD API - authenticated using authentication token.
provider "argocd" {
  server_addr = "argocd.local:443"
  auth_token  = "1234..."
}

# Exposed ArgoCD API - authenticated using `username`/`password`
provider "argocd" {
  server_addr = "argocd.local:443"
  username    = "foo"
  password    = local.password
}

# Exposed ArgoCD API - (pre)authenticated using local ArgoCD config (e.g. when
# you have previously logged in using SSO).
provider "argocd" {
  use_local_config = true
  # context = "foo" # Use explicit context from ArgoCD config instead of `current-context`.
}

# Unexposed ArgoCD API - using the current Kubernetes context and
# port-forwarding to temporarily expose ArgoCD API and authenticating using
# `auth_token`.
provider "argocd" {
  auth_token   = "1234..."
  port_forward = true
}

# Unexposed ArgoCD API - using port-forwarding to temporarily expose ArgoCD API
# whilst overriding the current context in kubeconfig.
provider "argocd" {
  auth_token                  = "1234..."
  port_forward_with_namespace = "custom-argocd-namespace"
  kubernetes {
    config_context = "kind-argocd"
  }
}

# Unexposed ArgoCD API - using `core` to run ArgoCD server locally and
# communicate directly with the Kubernetes API.
provider "argocd" {
  core = true
}

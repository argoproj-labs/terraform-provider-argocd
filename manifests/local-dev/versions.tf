terraform {
  required_providers {
    argocd = {
      source  = "oboukili/argocd"
      version = ">= 5.0.0"
    }
    kind = {
      source  = "unicell/kind"
      version = "0.0.2-u2"
    }
  }
}

provider "argocd" {
  server_addr = "127.0.0.1:8080"
  insecure    = true
  username    = "admin"
  password    = "acceptancetesting"
}

# provider "argocd" {
#   context          = "localhost:8080"
#   use_local_config = true # Note: you will need to log in via the ArgoCD CLI first (`argocd login localhost:8080 --username admin --password acceptancetesting --insecure`) for this to work
# }

# provider "argocd" {
#   port_forward = true
#   username     = "admin"
#   password     = "acceptancetesting"
#   kubernetes {
#     config_context = "kind-argocd"
#   }
# }

# provider "argocd" {
#   core = true
# }

provider "kind" {}

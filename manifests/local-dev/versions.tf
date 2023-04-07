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

provider "kind" {}

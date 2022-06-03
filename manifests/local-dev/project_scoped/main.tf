terraform {
  required_providers {
    argocd = {
      source  = "oboukili/argocd"
      version = "1.0.0"
//      version = "0.4.8"
    }
  }
}

provider "argocd" {
  server_addr = "127.0.0.1:8080"
  insecure    = true
  auth_token = ""
}

resource "argocd_repository" "myrepo" {
  name = "nginx"
  repo = "https://github.com/oboukili/terraform-provider-argocd.git"
  type = "git"
  project_name = "foo"
}
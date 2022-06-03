terraform {
  required_providers {
    argocd = {
      source  = "myregistry/oboukili/argocd"
      version = "1.0.0"
//      version = "0.4.8"
    }
  }
}

provider "argocd" {
  server_addr = "127.0.0.1:8080"
  insecure    = true
  username    = "admin"
  password    = "acceptancetesting"
}

resource "argocd_project" "foo" {
  metadata {
    name      = "foo"
    namespace = "argocd"
    labels = {
      acceptance = "true"
    }
  }

  spec {
    description  = "simple project"
    source_repos = ["*"]

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "foo"
    }
    cluster_resource_whitelist {
      group = "rbac.authorization.k8s.io"
      kind  = "ClusterRoleBinding"
    }
    cluster_resource_whitelist {
      group = "rbac.authorization.k8s.io"
      kind  = "ClusterRole"
    }
    namespace_resource_blacklist {}
    namespace_resource_whitelist {
      group = "*"
      kind  = "*"
    }
    orphaned_resources {
      warn = true
    }
    role {
      name = "foo-role"
      policies = [
        "p, proj:foo:foo-role, applications, *, foo/*, allow",
      ]
    }
    sync_window {
      kind         = "allow"
      applications = ["api-*"]
      clusters     = ["*"]
      namespaces   = ["*"]
      duration     = "3600s"
      schedule     = "10 1 * * *"
      manual_sync  = true
    }
    sync_window {
      kind         = "deny"
      applications = ["foo"]
      clusters     = ["in-cluster"]
      namespaces   = ["default"]
      duration     = "12h"
      schedule     = "22 1 5 * *"
      manual_sync  = false
    }
  }
}

resource "argocd_project_token" "foo_token" {
  project      = "foo"
  role         = "foo-role"
  description  = "short lived token"
}
resource "argocd_project" "foo" {
  metadata {
    name      = "foo"
    namespace = "argocd"
    labels = {
      acceptance = "true"
    }
    annotations = {
      "this.is.a.really.long.nested.key" = "yes, really!"
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

    destination {
      server    = argocd_cluster.kind_secondary.server
      namespace = "default"
    }

    cluster_resource_whitelist {
      group = "rbac.authorization.k8s.io"
      kind  = "ClusterRoleBinding"
    }

    cluster_resource_whitelist {
      group = "rbac.authorization.k8s.io"
      kind  = "ClusterRole"
    }

    namespace_resource_blacklist {
      group = "networking.k8s.io"
      kind  = "Ingress"
    }

    namespace_resource_whitelist {
      group = "*"
      kind  = "*"
    }

    role {
      name = "foo"
      policies = [
        "p, proj:foo:foo, applications, get, foo/*, allow",
        "p, proj:foo:foo, applications, sync, foo/*, deny",
      ]
    }

    orphaned_resources {
      warn = true
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
      timezone     = "Europe/London"
    }
  }
}

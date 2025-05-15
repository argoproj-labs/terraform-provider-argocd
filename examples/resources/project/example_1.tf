resource "argocd_project" "myproject" {
  metadata {
    name      = "myproject"
    namespace = "argocd"
    labels = {
      acceptance = "true"
    }
    annotations = {
      "this.is.a.really.long.nested.key" = "yes, really!"
    }
  }

  spec {
    description = "simple project"

    source_namespaces = ["argocd"]
    source_repos      = ["*"]

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "foo"
    }
    destination {
      name      = "anothercluster"
      namespace = "bar"
    }

    cluster_resource_blacklist {
      group = "*"
      kind  = "*"
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

    orphaned_resources {
      warn = true

      ignore {
        group = "apps/v1"
        kind  = "Deployment"
        name  = "ignored1"
      }

      ignore {
        group = "apps/v1"
        kind  = "Deployment"
        name  = "ignored2"
      }
    }

    role {
      name = "testrole"
      policies = [
        "p, proj:myproject:testrole, applications, override, myproject/*, allow",
        "p, proj:myproject:testrole, applications, sync, myproject/*, allow",
        "p, proj:myproject:testrole, clusters, get, myproject/*, allow",
        "p, proj:myproject:testrole, repositories, create, myproject/*, allow",
        "p, proj:myproject:testrole, repositories, delete, myproject/*, allow",
        "p, proj:myproject:testrole, repositories, update, myproject/*, allow",
        "p, proj:myproject:testrole, logs, get, myproject/*, allow",
        "p, proj:myproject:testrole, exec, create, myproject/*, allow",
      ]
    }
    role {
      name = "anotherrole"
      policies = [
        "p, proj:myproject:testrole, applications, get, myproject/*, allow",
        "p, proj:myproject:testrole, applications, sync, myproject/*, deny",
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
      timezone     = "Europe/London"
    }

    signature_keys = [
      "4AEE18F83AFDEB23",
      "07E34825A909B250"
    ]
  }
}

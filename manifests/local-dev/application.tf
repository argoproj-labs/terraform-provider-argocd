resource "argocd_application" "foo" {
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
    project = argocd_project.foo.metadata[0].name

    source {
      repo_url        = "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami"
      chart           = "redis"
      target_revision = "16.9.11"

      helm {
        release_name = "testing"

        parameter {
          name  = "image.tag"
          value = "6.2.5"
        }

        parameter {
          name  = "architecture"
          value = "standalone"
        }
      }
    }

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }

    sync_policy {
      automated {
        prune       = true
        self_heal   = true
        allow_empty = false
      }

      sync_options = [
        "PrunePropagationPolicy=foreground",
        "ApplyOutOfSyncOnly=true"
      ]

      retry {
        limit = 5
        backoff {
          duration     = "3m"
          factor       = "2"
          max_duration = "30m"
        }
      }
    }
  }

  # wait = true
}

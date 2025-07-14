# Kustomize application
resource "argocd_application" "kustomize" {
  metadata {
    name      = "kustomize-app"
    namespace = "argocd"
    labels = {
      test = "true"
    }
  }

  cascade = false # disable cascading deletion
  wait    = true

  spec {
    project = "myproject"

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "foo"
    }

    source {
      repo_url        = "https://github.com/kubernetes-sigs/kustomize"
      path            = "examples/helloWorld"
      target_revision = "master"
      kustomize {
        name_prefix = "foo-"
        name_suffix = "-bar"
        images      = ["hashicorp/terraform:light"]
        common_labels = {
          "this.is.a.common" = "la-bel"
          "another.io/one"   = "true"
        }
      }
    }

    sync_policy {
      automated {
        prune       = true
        self_heal   = true
        allow_empty = true
      }
      # Only available from ArgoCD 1.5.0 onwards
      sync_options = ["Validate=false"]
      retry {
        limit = "5"
        backoff {
          duration     = "30s"
          max_duration = "2m"
          factor       = "2"
        }
      }
    }

    ignore_difference {
      group         = "apps"
      kind          = "Deployment"
      json_pointers = ["/spec/replicas"]
    }

    ignore_difference {
      group = "apps"
      kind  = "StatefulSet"
      name  = "someStatefulSet"
      json_pointers = [
        "/spec/replicas",
        "/spec/template/spec/metadata/labels/bar",
      ]
      # Only available from ArgoCD 2.1.0 onwards
      jq_path_expressions = [
        ".spec.replicas",
        ".spec.template.spec.metadata.labels.bar",
      ]
    }
  }
}

# Helm application
resource "argocd_application" "helm" {
  metadata {
    name      = "helm-app"
    namespace = "argocd"
    labels = {
      test = "true"
    }
  }

  spec {
    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }

    source {
      repo_url        = "https://some.chart.repo.io"
      chart           = "mychart"
      target_revision = "1.2.3"
      helm {
        release_name = "testing"
        parameter {
          name  = "image.tag"
          value = "1.2.3"
        }
        parameter {
          name  = "someotherparameter"
          value = "true"
        }
        value_files = ["values-test.yml"]
        values = yamlencode({
          someparameter = {
            enabled   = true
            someArray = ["foo", "bar"]
          }
        })
      }
    }
  }
}

# Multiple Application Sources with Helm value files from external Git repository
resource "argocd_application" "multiple_sources" {
  metadata {
    name      = "helm-app-with-external-values"
    namespace = "argocd"
  }

  spec {
    project = "default"

    source {
      repo_url        = "https://charts.helm.sh/stable"
      chart           = "wordpress"
      target_revision = "9.0.3"
      helm {
        value_files = ["$values/helm-dependency/values.yaml"]
      }
    }

    source {
      repo_url        = "https://github.com/argoproj/argocd-example-apps.git"
      target_revision = "HEAD"
      ref             = "values"
    }

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
  }
}

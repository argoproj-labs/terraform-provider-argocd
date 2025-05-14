resource "argocd_application_set" "clusters" {
  metadata {
    name = "clusters"
  }

  spec {
    generator {
      clusters {}
    }

    template {
      metadata {
        name = "appset-clusters-{{name}}"
      }

      spec {
        project = "default"

        source {
          repo_url        = "https://github.com/argoproj/argo-cd/"
          target_revision = "HEAD"
          chart           = "test/e2e/testdata/guestbook"
        }

        destination {
          server    = "{{server}}"
          namespace = "default"
        }
      }
    }
  }
}

resource "argocd_application_set" "cluster_decision_resource" {
  metadata {
    name = "cluster-decision-resource"
  }

  spec {
    generator {
      cluster_decision_resource {
        config_map_ref = "my-configmap"
        name           = "quak"

        label_selector {
          match_labels = {
            duck = "spotted"
          }

          match_expressions {
            key      = "duck"
            operator = "In"
            values = [
              "spotted",
              "canvasback"
            ]
          }
        }
      }
    }

    template {
      metadata {
        name = "appset-cdr-{{name}}"
      }

      spec {
        source {
          repo_url        = "https://github.com/argoproj/argo-cd/"
          target_revision = "HEAD"
          path            = "test/e2e/testdata/guestbook"
        }

        destination {
          server    = "{{server}}"
          namespace = "default"
        }
      }
    }
  }
}

resource "argocd_application_set" "git" {
  metadata {
    name = "git"
  }

  spec {
    generator {
      git {
        repo_url = "https://github.com/argoproj/argo-cd.git"
        revision = "HEAD"

        # directory {
        #   path = "applicationset/examples/git-generator-directory/excludes/cluster-addons/*"
        # }

        # directory {
        #   exclude = true
        #   path    = "applicationset/examples/git-generator-directory/excludes/cluster-addons/exclude-helm-guestbook"
        # }

        file {
          path = "applicationset/examples/git-generator-files-discovery/cluster-config/**/config.json"
        }
      }
    }

    template {
      metadata {
        name = "appset-git-{{path.basename}}"
      }

      spec {
        project = "default"

        source {
          repo_url        = "https://github.com/argoproj/argo-cd.git"
          target_revision = "HEAD"
          path            = "{{path}}"
        }

        destination {
          server    = "https://kubernetes.default.svc"
          namespace = "{{path.basename}}"
        }
      }
    }
  }
}

resource "argocd_application_set" "list" {
  metadata {
    name = "list"
  }

  spec {
    generator {
      list {
        elements = [
          {
            cluster = "engineering-dev"
            url     = "https://kubernetes.default.svc"
          },
          {
            cluster = argocd_cluster.kind_secondary.name
            url     = argocd_cluster.kind_secondary.server
          }
        ]

        template {
          metadata {}
          spec {
            project = "default"
            source {
              target_revision = "HEAD"
              repo_url        = "https://github.com/argoproj/argo-cd.git"
              # New path value is generated here:
              path = "applicationset/examples/template-override/{{cluster}}-override"
            }
            destination {}
          }
        }
      }
    }

    template {
      metadata {
        name = "appset-list-{{cluster}}"
      }

      spec {
        project = "default"

        source {
          repo_url        = "https://github.com/argoproj/argo-cd.git"
          target_revision = "HEAD"
          path            = "applicationset/examples/list-generator/guestbook/{{cluster}}"
        }

        destination {
          server    = "{{url}}"
          namespace = "guestbook"
        }
      }
    }
  }
}

resource "argocd_application_set" "matrix" {
  metadata {
    name = "matrix"
  }

  spec {
    generator {
      matrix {
        generator {
          matrix {
            generator {
              list {
                elements = [
                  {
                    cluster = "in-cluster"
                    url     = "https://kubernetes.default.svc"
                  }
                ]
              }
            }

            generator {
              git {
                repo_url = "https://github.com/argoproj/argo-cd.git"
                revision = "HEAD"

                file {
                  path = "applicationset/examples/git-generator-files-discovery/cluster-config/**/config.json"
                }
              }
            }
          }
        }

        generator {
          clusters {}
        }
      }
    }

    template {
      metadata {
        name = "appset-matrix-{{name}}"
      }

      spec {
        project = "default"

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
          server    = "{{server}}"
          namespace = "default"
        }
      }
    }
  }
}

resource "argocd_application_set" "merge" {
  metadata {
    name = "merge"
  }

  spec {
    generator {
      merge {
        merge_keys = [
          "server"
        ]

        generator {
          clusters {
            values = {
              kafka = true
              redis = false
            }
          }
        }

        generator {
          clusters {
            selector {
              match_labels = {
                use-kafka = "false"
              }
            }

            values = {
              kafka = "false"
            }
          }
        }

        generator {
          list {
            elements = [
              {
                server         = "https://2.4.6.8"
                "values.redis" = "true"
              },
            ]
          }
        }
      }
    }

    template {
      metadata {
        name = "appset-merge-{{name}}"
      }

      spec {
        project = "default"

        source {
          repo_url        = "https://github.com/argoproj/argo-cd.git"
          path            = "app"
          target_revision = "HEAD"

          helm {
            parameter {
              name  = "kafka"
              value = "{{values.kafka}}"
            }

            parameter {
              name  = "redis"
              value = "{{values.redis}}"
            }
          }
        }

        destination {
          server    = "{{server}}"
          namespace = "default"
        }
      }
    }
  }
}

resource "argocd_application_set" "pull_request" {
  metadata {
    name = "pull-request"
  }

  spec {
    generator {
      pull_request {
        github {
          api             = "https://git.example.com/"
          owner           = "myorg"
          repo            = "myrepository"
          app_secret_name = "github-app-repo-creds"

          token_ref {
            secret_name = "github-token"
            key         = "token"
          }

          labels = [
            "preview"
          ]
        }
      }
    }

    template {
      metadata {
        name = "appset-opull-request-{{branch}}-{{number}}"
      }

      spec {
        project = "default"

        source {
          repo_url        = "https://github.com/myorg/myrepo.git"
          path            = "kubernetes/"
          target_revision = "{{head_sha}}"

          helm {
            parameter {
              name  = "image.tag"
              value = "pull-{{head_sha}}"
            }
          }
        }

        destination {
          server    = "https://kubernetes.default.svc"
          namespace = "default"
        }
      }
    }
  }
}

resource "argocd_application_set" "scm_provider" {
  metadata {
    name = "scm-provider"
  }

  spec {
    generator {
      scm_provider {
        github {
          all_branches    = true
          api             = "https://git.example.com/"
          app_secret_name = "gh-app-repo-creds"
          organization    = "myorg"

          token_ref {
            secret_name = "github-token"
            key         = "token"
          }
        }
      }
    }

    template {
      metadata {
        name = "appset-scm-provider-{{repository}}"
      }

      spec {
        project = "default"

        source {
          repo_url        = "{{url}}"
          path            = "kubernetes/"
          target_revision = "{{branch}}"
        }

        destination {
          server    = "https://kubernetes.default.svc"
          namespace = "default"
        }
      }
    }
  }
}

resource "argocd_application_set" "progressive_sync" {
  metadata {
    name = "progressive-sync"
  }

  spec {
    generator {
      list {
        elements = [
          {
            cluster = "engineering-dev"
            url     = "https://1.2.3.4"
            env     = "env-dev"
          },
          {
            cluster = "engineering-qa"
            url     = "https://2.4.6.8"
            env     = "env-qa"
          },
          {
            cluster = "engineering-prod"
            url     = "https://9.8.7.6/"
            env     = "env-prod"
          }
        ]
      }
    }

    strategy {
      type = "RollingSync"
      rolling_sync {
        step {
          match_expressions {
            key      = "envLabel"
            operator = "In"
            values = [
              "env-dev"
            ]
          }

          # max_update = "100%"  # if undefined, all applications matched are updated together (default is 100%)
        }

        step {
          match_expressions {
            key      = "envLabel"
            operator = "In"
            values = [
              "env-qa"
            ]
          }

          max_update = "0"
        }

        step {
          match_expressions {
            key      = "envLabel"
            operator = "In"
            values = [
              "env-prod"
            ]
          }

          max_update = "10%"
        }
      }
    }

    go_template = true

    template {
      metadata {
        name = "appset-progressive-sync-{{.cluster}}"
        labels = {
          envLabel = "{{.env}}"
        }
      }

      spec {
        project = "default"

        source {
          repo_url        = "https://github.com/infra-team/cluster-deployments.git"
          path            = "guestbook/{{.cluster}}"
          target_revision = "HEAD"
        }

        destination {
          server    = "{{.url}}"
          namespace = "guestbook"
        }
      }
    }
  }
}

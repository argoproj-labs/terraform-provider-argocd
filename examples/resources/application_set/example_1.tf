# Clusters Generator
resource "argocd_application_set" "clusters_selector" {
  metadata {
    name = "clusters-selector"
  }

  spec {
    generator {
      clusters {
        selector {
          match_labels = {
            "argocd.argoproj.io/secret-type" = "cluster"
          }
        }
      }
    }

    template {
      metadata {
        name = "{{name}}-clusters-selector"
      }

      spec {
        source {
          repo_url        = "https://github.com/argoproj/argocd-example-apps/"
          target_revision = "HEAD"
          path            = "guestbook"
        }

        destination {
          server    = "{{server}}"
          namespace = "default"
        }
      }
    }
  }
}

# Cluster Decision Resource Generator
resource "argocd_application_set" "cluster_decision_resource" {
  metadata {
    name = "cluster-decision-resource"
  }

  spec {
    generator {
      cluster_decision_resource {
        config_map_ref = "my-configmap"
        name           = "quak"
      }
    }

    template {
      metadata {
        name = "{{name}}-guestbook"
      }

      spec {
        source {
          repo_url        = "https://github.com/argoproj/argocd-example-apps/"
          target_revision = "HEAD"
          path            = "guestbook"
        }

        destination {
          server    = "{{server}}"
          namespace = "default"
        }
      }
    }
  }
}


# Git Generator - Directories
resource "argocd_application_set" "git_directories" {
  metadata {
    name = "git-directories"
  }

  spec {
    generator {
      git {
        repo_url = "https://github.com/argoproj/argo-cd.git"
        revision = "HEAD"

        directory {
          path = "applicationset/examples/git-generator-directory/cluster-addons/*"
        }

        directory {
          path    = "applicationset/examples/git-generator-directory/excludes/cluster-addons/exclude-helm-guestbook"
          exclude = true
        }
      }
    }

    template {
      metadata {
        name = "{{path.basename}}-git-directories"
      }

      spec {
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

# Git Generator - Files
resource "argocd_application_set" "git_files" {
  metadata {
    name = "git-files"
  }

  spec {
    generator {
      git {
        repo_url = "https://github.com/argoproj/argo-cd.git"
        revision = "HEAD"

        file {
          path = "applicationset/examples/git-generator-files-discovery/cluster-config/**/config.json"
        }
      }
    }

    template {
      metadata {
        name = "{{cluster.name}}-git-files"
      }

      spec {
        source {
          repo_url        = "https://github.com/argoproj/argo-cd.git"
          target_revision = "HEAD"
          path            = "applicationset/examples/git-generator-files-discovery/apps/guestbook"
        }

        destination {
          server    = "{{cluster.address}}"
          namespace = "guestbook"
        }
      }
    }
  }
}

# List Generator
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
            cluster = "engineering-prod"
            url     = "https://kubernetes.default.svc"
            foo     = "bar"
          }
        ]
      }
    }

    template {
      metadata {
        name = "{{cluster}}-guestbook"
      }

      spec {
        project = "my-project"

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

# Matrix Generator
resource "argocd_application_set" "matrix" {
  metadata {
    name = "matrix"
  }

  spec {
    generator {
      matrix {
        generator {
          git {
            repo_url = "https://github.com/argoproj/argo-cd.git"
            revision = "HEAD"

            directory {
              path = "applicationset/examples/matrix/cluster-addons/*"
            }
          }
        }

        generator {
          clusters {
            selector {
              match_labels = {
                "argocd.argoproj.io/secret-type" = "cluster"
              }
            }
          }
        }
      }
    }

    template {
      metadata {
        name = "{{path.basename}}-{{name}}"
      }

      spec {
        project = "default"

        source {
          repo_url        = "https://github.com/argoproj/argo-cd.git"
          target_revision = "HEAD"
          path            = "{{path}}"
        }

        destination {
          server    = "{{server}}"
          namespace = "{{path.basename}}"
        }
      }
    }
  }
}

# Merge Generator
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
        name = "{{name}}"
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

# Pull Request Generator - GitHub
resource "argocd_application_set" "pr_github" {
  metadata {
    name = "pr-github"
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
        name = "myapp-{{branch}}-{{number}}"
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

# SCM Provider Generator - GitHub
resource "argocd_application_set" "scm_github" {
  metadata {
    name = "scm-github"
  }

  spec {
    generator {
      scm_provider {
        github {
          app_secret_name = "gh-app-repo-creds"
          organization    = "myorg"

          # all_branches = true
          # api          = "https://git.example.com/"

          # token_ref {
          #   secret_name = "github-token"
          #   key         = "token"
          # }
        }
      }
    }

    template {
      metadata {
        name = "{{repository}}"
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

# Progressive Sync - Rolling Update
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
        name = "{{.cluster}}-guestbook"
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

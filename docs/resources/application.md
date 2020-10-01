# argocd_application

Creates an ArgoCD application.

## Example Usage

```hcl
resource "argocd_application" "kustomize" {
  metadata {
    name      = "kustomize-app"
    namespace = "argocd"
    labels = {
      test = "true"
    }
  }

  spec {
    project = argocd_project.myproject.metadata.0.name

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

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "foo"
    }

    sync_policy {
      automated = {
        prune     = true
        self_heal = true
      }
      # Only available from ArgoCD 1.5.0 onwards
      sync_options = ["Validate=false"]
    }

    ignore_difference {
      group         = "apps"
      kind          = "Deployment"
      json_pointers = ["/spec/replicas"]
    }

    ignore_difference {
      group         = "apps"
      kind          = "StatefulSet"
      name          = "someStatefulSet"
      json_pointers = [
        "/spec/replicas",
        "/spec/template/spec/metadata/labels/bar",
      ]
    }
  }
}

resource "argocd_application" "helm" {
  metadata {
    name      = "helm-app"
    namespace = "argocd"
    labels = {
      test = "true"
    }
  }

  spec {
    source {
      repo_url        = "https://some.chart.repo.io"
      chart           = "mychart"
      target_revision = "1.2.3"
      helm {
        parameter {
          name  = "image.tag"
          value = "1.2.3"
        }
        parameter {
          name  = "someotherparameter"
          value = "true"
        }
        value_files = ["values-test.yml"]
        values      = <<EOT
someparameter:
  enabled: true
  someArray:
  - foo
  - bar    
EOT
        release_name = "testing"
      }
    }

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
  }
}
```

## Argument Reference

* `metadata` - (Required) Standard Kubernetes API service's metadata. For more info see the [Kubernetes reference](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata).
* `spec` - (Required) The application specification, the nested attributes are documented below.

The `metadata` block can have the following attributes:

* `name` - (Required) The project name, must be unique, cannot be updated.
* `annotations` - (Optional) An unstructured key value map stored with the config map that may be used to store arbitrary metadata. **By default, the provider ignores any annotations whose key names end with kubernetes.io. This is necessary because such annotations can be mutated by server-side components and consequently cause a perpetual diff in the Terraform plan output. If you explicitly specify any such annotations in the configuration template then Terraform will consider these as normal resource attributes and manage them as expected (while still avoiding the perpetual diff problem)**. For more info see [Kubernetes reference](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/).
* `labels` - (Optional) Map of string keys and values that can be used to organize and categorize (scope and select) the config map. May match selectors of replication controllers and services. **By default, the provider ignores any labels whose key names end with kubernetes.io. This is necessary because such labels can be mutated by server-side components and consequently cause a perpetual diff in the Terraform plan output. If you explicitly specify any such labels in the configuration template then Terraform will consider these as normal resource attributes and manage them as expected (while still avoiding the perpetual diff problem).** For more info see [Kubernetes reference](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/).

The `spec` block can have the following attributes:

* `destination` - (Required) The allowed cluster/namespace project destination. Structure is documented below.
* `source` - (Required) Contains information about Git repository, path within repository and target application environment. Structure is documented below.
* `project` - (Optional) The ArgoCD project where the application will reside. Empty name means that application belongs to `default` project. Defaults to `default`. 
* `sync_policy` - (Optional) Controls when a sync will be performed. Structure is documented below.
* `ignore_difference` - (Optional) Controls resources fields which should be ignored during comparison. Can be repeated multiple times. Structure is documented below.
* `info` - (Optional) Contains a list of useful information (URLs, email addresses, and plain text) that relates to the application. Can be repeated multiple times. Structure is documented below.
* `revision_history_limit` -  (Optional) This limits the number of items kept in the apps revision history. This should only be changed in exceptional circumstances. Setting to zero will store no history. This will reduce storage used. Increasing will increase the space used to store the history, so we do not recommend increasing it. Default is `10`.

Each `info` block can have the following attributes:
* `name` - (Optional).
* `value` - (Optional).

The `destination` block has the following attributes:
* `server` - (Required) The cluster URL to deploy the application to.
* `namespace` - (Required) The namespace to deploy the application to.

The `sync_policy` block has the following attributes:
* `automated` - (Optional)
* `sync_options` - (Optional) (Only available from ArgoCD 1.5.0 onwards), only `["Validate=false"]` is supported.

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
    project = "myproject"

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
        prune       = true
        self_heal   = true
        allow_empty = true
      }
      # Only available from ArgoCD 1.5.0 onwards
      sync_options = ["Validate=false"]
      retry {
        limit   = "5"
        backoff = {
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
      group         = "apps"
      kind          = "StatefulSet"
      name          = "someStatefulSet"
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

resource "argocd_application" "helm" {
  metadata {
    name      = "helm-app"
    namespace = "argocd"
    labels = {
      test = "true"
    }
  }

  wait = true

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
* `wait` - (Optional) boolean, wait for application to be synced and healthy upon creation and updates, also waits for Kubernetes resources to be truly deleted upon deletion. Wait timeouts are controlled by Terraform Create, Update and Delete resource timeouts (all default to 5 minutes). Default is `false`.

The `metadata` block can have the following attributes:

* `name` - (Required) The project name, must be unique, cannot be updated.
* `annotations` - (Optional) An unstructured key value map stored with the config map that may be used to store arbitrary metadata. **By default, the provider ignores any annotations whose key names end with kubernetes.io. This is necessary because such annotations can be mutated by server-side components and consequently cause a perpetual diff in the Terraform plan output. If you explicitly specify any such annotations in the configuration template then Terraform will consider these as normal resource attributes and manage them as expected (while still avoiding the perpetual diff problem)**. For more info see [Kubernetes reference](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/).
* `labels` - (Optional) Map of string keys and values that can be used to organize and categorize (scope and select) the config map. May match selectors of replication controllers and services. **By default, the provider ignores any labels whose key names end with kubernetes.io. This is necessary because such labels can be mutated by server-side components and consequently cause a perpetual diff in the Terraform plan output. If you explicitly specify any such labels in the configuration template then Terraform will consider these as normal resource attributes and manage them as expected (while still avoiding the perpetual diff problem).** For more info see [Kubernetes reference](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/).

The `spec` block can have the following attributes:

* `destination` - (Required) The allowed cluster/namespace project destination. Structure is documented below.
* `source` - (Required) Contains information about Git repository, path within repository and target application environment. Structure is documented below.
* `project` - (Optional) The ArgoCD project where the application will reside. Empty name means that application belongs to `default` project. Defaults to `default`. 
* `sync_policy` - (Optional) Controls when a sync will be performed. Structure is documented below.
* `ignore_difference` - (Optional) Controls resources' fields which should be ignored during comparison with live state. Can be repeated multiple times. Structure is documented below.
* `info` - (Optional) Contains a list of useful information (URLs, email addresses, and plain text) that relates to the application. Can be repeated multiple times. Structure is documented below.
* `revision_history_limit` -  (Optional) This limits the number of items kept in the apps revision history. This should only be changed in exceptional circumstances. Setting to zero will store no history. This will reduce storage used. Increasing will increase the space used to store the history, so we do not recommend increasing it. Default is `10`.

Each `info` block can have the following attributes:
* `name` - (Optional).
* `value` - (Optional).

The `destination` block has the following attributes:
* `server` - (Optional) The cluster URL to deploy the application to. At most one of `server` or `name` is required.
* `namespace` - (Required) The namespace to deploy the application to.
* `name` - (Optional) Name of the destination cluster which can be used instead of server (url) field. At most one of `server` or `name` is required.

The `sync_policy` block has the following attributes:
* `automated` - (Optional) map(string) of strings, will keep an application synced to the target revision. Structure is documented below
* `sync_options` - (Optional) list of sync options, allow you to specify whole app sync-options (only available from ArgoCD 1.5.0 onwards).
* `retry` - (Optional) controls failed sync retry behavior, structure is documented below

The `sync_policy/automated` map has the following attributes:
* `prune` - (Optional), boolean, will prune resources automatically as part of automated sync. Defaults to `false`.
* `self_heal` - (Optional), boolean, enables auto-syncing if the live resources differ from the targeted revision manifests. Defaults to `false`.
* `allow_empty` - (Optional), boolean, allows apps to have zero live resources. Defaults to `false`.

The `sync_policy/retry` block has the following attributes:
* `limit` - (Optional), max number of allowed sync retries, as a string.
* `backoff` - (Optional), retry backoff strategy, structure is documented below

The `sync_policy/retry/backoff` map has the following attributes:
* `duration` - (Optional), Duration is the amount to back off. Default unit is seconds, but could also be a duration (e.g. "2m", "1h"), as a string.
* `factor` - (Optional), Factor is a factor to multiply the base duration after each failed retry, as a string.
* `max_duration` - (Optional), is the maximum amount of time allowed for the backoff strategy. Default unit is seconds, but could also be a duration (e.g. "2m", "1h"), as a string.

Each `ignore_difference` block can have the following attributes:
* `group` - (Optional) The targeted Kubernetes resource kind.
* `kind` - (Optional) The targeted Kubernetes resource kind.
* `name` - (Optional) The targeted Kubernetes resource name.
* `namespace` - (Optional) The targeted Kubernetes resource namespace.
* `json_pointers` - (Optional) List of JSONPaths strings targeting the field(s) to ignore.
* `jq_path_expressions` - (Optional) List of jq path expression strings targeting the field(s) to ignore (only available from ArgoCD 2.1.0 onwards).

The `source` block has the following attributes:
* `repo_url` - (Required) string, repository URL of the application manifests.
* `path` - (Optional) string, directory path within the Git repository.
* `target_revision` - (Optional) string, defines the commit, tag, or branch in which to sync the application to.
* `chart` - (Optional) Helm chart name, only applicable when the application manifests are a Helm chart.
Only one of the following `source` attributes can be defined at a time:
* `helm` - (Optional) holds Helm specific options. Structure is documented below.
* `kustomize` - (Optional) holds Kustomize specific options. Structure is documented below.
* `ksonnet` - (Optional) holds Ksonnet specific options. Structure is documented below.
* `directory` - (Optional) holds path/directory specific options (native Kubernetes manifests or **Jsonnet** manifests). Structure is documented below.
* `plugin` - (Optional) holds config management plugin specific options. Structure is documented below.

The `helm` block has the following attributes:
* `value_files` - (Optional) list of Helm value files to use when generating a template.
* `values` - (Optional) Helm values, typically defined as a block.
* `release_name` - (Optional) the Helm release name. If omitted it will use the application name.
* `parameter` - (Optional) parameter to the Helm template. Can be repeated multiple times. Structure is documented below.

Each `helm/parameter` block has the following attributes:
* `name` - (Optional) string, name of the helm parameter.
* `value` - (Optional) string, value of the helm parameter.
* `force_string` - (Optional) boolean, determines whether to tell Helm to interpret booleans and numbers as strings.

The `kustomize` block has the following attributes:
* `name_prefix` - (Optional) string, prefix appended to resources for kustomize apps.
* `name_suffix` - (Optional) string, suffix appended to resources for kustomize apps.
* `version` - (Optional) string, contains optional Kustomize version.
* `images` - (Optional) set of strings, kustomize image overrides.
* `common_labels` - (Optional) map(string) of strings, adds additional kustomize commonLabels.
* `common_annotations` - (Optional) map(string) of strings, adds additional kustomize commonAnnotations.

The `ksonnet` block has the following attributes:
* `environment` - (Optional) string, Ksonnet application environment name.
* `parameters` - (Optional) Set of ksonnet component parameter overrides. Can be repeated multiple times. Structure is documented below.

Each `ksonnet/parameters` block has the following attributes:
* `name` - (Optional) string, name of the Ksonnet parameter.
* `value` - (Optional) string, value of the Ksonnet parameter.
* `component` - (Optional) string, value of the component parameter.

The `directory` block has the following attributes:
* `recurse` - (Optional) boolean, determines whether to recursively look for manifests in specified path.
* `jsonnet` - (Optional), Jsonnet parameters. Structure is documented below.

The `directory/jsonnet` block can have the following attributes:
* `ext_var` - (Optional) Jsonnet External Variable. Can be repeated multiple times. Structure is documented below.
* `tla` - (Optional) Jsonnet Top-level Arguments. Can be repeated multiple times. Structure is documented below.

Each `directory/jsonnet/ext_var` and `directory/jsonnet/tla` can have the following attributes:
* `name` - (Optional) string.
* `value` - (Optional) string.
* `Code` - (Optional) boolean.

The `plugin` block has the following attributes:
* `name` - (Optional) string.
* `env` - (Optional) Can be repeated multiple times. Structure is documented below.

Each `plugin/env` block has the following attributes:
* `name` - (Optional) string.
* `value` - (Optional) string.

## Import

ArgoCD applications can be imported using an id consisting of `{name}`, e.g.
```
$ terraform import argocd_application.myapp myapp
```

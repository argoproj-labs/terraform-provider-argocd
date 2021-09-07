# argocd_project

Creates an ArgoCD project.

## Example Usage

```hcl
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
    }
    signature_keys = [
      "4AEE18F83AFDEB23",
      "07E34825A909B250"
    ]
  }
}

```

## Argument Reference

* `metadata` - (Required) Standard Kubernetes API service's metadata. For more info see the [Kubernetes reference](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata).
* `spec` - (Required) The project specification, the nested attributes are documented below.

The `metadata` block can have the following attributes:

* `name` - (Required) The project name, must be unique, cannot be updated.
* `annotations` - (Optional) An unstructured key value map stored with the config map that may be used to store arbitrary metadata. **By default, the provider ignores any annotations whose key names end with kubernetes.io. This is necessary because such annotations can be mutated by server-side components and consequently cause a perpetual diff in the Terraform plan output. If you explicitly specify any such annotations in the configuration template then Terraform will consider these as normal resource attributes and manage them as expected (while still avoiding the perpetual diff problem)**. For more info see [Kubernetes reference](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/).
* `labels` - (Optional) Map of string keys and values that can be used to organize and categorize (scope and select) the config map. May match selectors of replication controllers and services. **By default, the provider ignores any labels whose key names end with kubernetes.io. This is necessary because such labels can be mutated by server-side components and consequently cause a perpetual diff in the Terraform plan output. If you explicitly specify any such labels in the configuration template then Terraform will consider these as normal resource attributes and manage them as expected (while still avoiding the perpetual diff problem).** For more info see [Kubernetes reference](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/).

The `spec` block can have the following attributes:

* `destination` - (Required) The allowed cluster/namespace project destination, can be repeated multiple times. 
* `source_repos` - (Required) List of strings containing allowed application repositories URLs for the project. Can be set to `["*"]` to allow all configured repositories configured in ArgoCD.
* `cluster_resource_whitelist` - (Optional) Cluster-scoped resource allowed to be managed by the project applications, can be repeated multiple times. 
* `description` - (Optional)
* `orphaned_resources` - (Optional) A key value map to control orphaned resources monitoring, 
* `namespace_resource_blacklist` - (Optional) Namespaced-scoped resources allowed to be managed by the project applications, can be repeated multiple times. 
* `role` - (Optional) can be repeated multiple times. 
* `sync_window` - (Optional) can be repeated multiple times. 
* `signature_keys` - (Optional) list of PGP key IDs strings that commits to be synced to must be signed with.

Each `cluster_resource_whitelist` block can have the following attributes:
* `group` - (Optional) The Kubernetes resource Group to match for.
* `kind` - (Optional) The Kubernetes resource Kind to match for.

The `orphaned_resources` block can have the following attributes:
* `warn` - (Optional) Boolean, defaults to `false`.
* `ignore` - (Optional), set of map of strings, specifies which Group/Kind/Name resource(s) to ignore. Can be repeated multiple times. Structure is documented below.

Each `orphaned_resources/ignore` block can have the following attributes:
* `group` - (Optional) The Kubernetes resource Group to match for.
* `kind` - (Optional) The Kubernetes resource Kind to match for.
* `name` - (Optional) The Kubernetes resource name to match for.

Each `namespace_resource_blacklist` block can have the following attributes:
* `group` - (Optional) The Kubernetes resource Group to match for.
* `kind` - (Optional) The Kubernetes resource Kind to match for.

Each `role` block can have the following attributes:
* `name` - (Required) Name of the role.
* `policies` - (Required) list of Casbin formated strings that define access policies for the role in the project, For more information, read the [ArgoCD RBAC reference](https://argoproj.github.io/argo-cd/operator-manual/rbac/#rbac-permission-structure).
* `description` - (Optional)
* `groups` - (Optional) List of OIDC group claims bound to this role.

Each `sync_window` block can have the following attributes:
* `applications` - (Optional) List of applications the window will apply to.
* `clusters` - (Optional) List of clusters the window will apply to.
* `duration` - (Optional) amount of time the sync window will be open.
* `kind` - (Optional) Defines if the window allows or blocks syncs, allowed values are `allow` or `deny`.
* `manual_sync` - (Optional) Boolean, enables manual syncs when they would otherwise be blocked.
* `namespaces` - (Optional) List of namespaces that the window will apply to.
* `schedule` - (Optional) Time the window will begin, specified in cron format.


## Import

ArgoCD projects can be imported using an id consisting of `{name}`, e.g.
```
$ terraform import argocd_project.myproject myproject
```

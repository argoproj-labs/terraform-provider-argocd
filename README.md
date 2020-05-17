# Terraform Provider for ArgoCD

![Acceptance Tests](https://github.com/oboukili/terraform-provider-argocd/workflows/Acceptance%20Tests/badge.svg)

---

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.12.24+

---

## Motivations

### *I thought ArgoCD already allowed for 100% declarative configuration?*

While that is true through the use of ArgoCD Kubernetes Custom Resources, 
there are some resources that simply cannot be managed using Kubernetes manifests,
such as project roles JWTs whose respective lifecycles are better handled by a tool like Terraform.
Even more so when you need to export these JWTs to another external system using Terraform, like a CI platform.

### *Wouldn't using a Kubernetes provider to handle ArgoCD configuration be enough?*

Existing Kubernetes providers do not patch arrays of objects, losing project role JWTs when doing small project changes just happen.

ArgoCD Kubernetes admission webhook controller is not as exhaustive as ArgoCD API validation, this can be seen with RBAC policies, where no validation occur when creating/patching a project.

Using Terraform to manage Kubernetes Custom Resource becomes increasingly difficult 
the further you use HCL2 DSL to merge different data structures *and* want to preserve type safety.

Whatever the Kubernetes CRD provider you are using, you will probably end up using `locals` and the `yamlencode` function **which does not preserve the values' type**.
In these cases, not only the readability of your Terraform plan will worsen, but you will also be losing some safeties that Terraform provides in the process.

---

## Installation

* **From binary releases**:
  Get the [latest release](https://github.com/oboukili/terraform-provider-argocd/releases/latest), or adapt and run the following:
  ```shell script
  curl -LO https://github.com/oboukili/terraform-provider-argocd/releases/download/v0.1.0/terraform-provider-argocd_v0.1.0_linux_amd64.gz
  gunzip -N terraform-provider-argocd_v0.1.0_linux_amd64.gz
  mv terraform-provider-argocd_v0.1.0 ~/.terraform.d/plugins/linux_amd64/
  ```

* **From source**: Follow [the 'contributing' build instructions](https://github.com/oboukili/terraform-provider-argocd#building).

## Usage

```hcl
provider "argocd" {
  server_addr = "argocd.local:443" # env ARGOCD_SERVER
  auth_token  = "1234..."          # env ARGOCD_AUTH_TOKEN
  # username  = "admin"            # env ARGOCD_AUTH_USERNAME
  # password  = "foo"              # env ARGOCD_AUTH_PASSWORD
  insecure = false # env ARGOCD_INSECURE
}

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
    orphaned_resources = {
      warn = true
    }
    role {
      name = "testrole"
      policies = [
        "p, proj:%s:testrole, applications, override, %s/*, allow",
        "p, proj:%s:testrole, applications, sync, %s/*, allow",
      ]
    }
    role {
      name = "anotherrole"
      policies = [
        "p, proj:%s:testrole, applications, get, %s/*, allow",
        "p, proj:%s:testrole, applications, sync, %s/*, deny",
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

resource "argocd_project_token" "secret" {
  count        = 20
  project      = argocd_project.myproject.metadata.0.name
  role         = "foobar"
  description  = "short lived token"
  expires_in   = "1h"
  renew_before = "30m"
}
```

---

## Contributing

Contributions are welcome! You'll first need a working installation of [Go 1.14+](http://www.golang.org). Just as a reminder. you will also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH).

### Building

Clone the repository within your `GOPATH`

```sh
mkdir -p $GOPATH/src/github.com/oboukili; cd $GOPATH/src/github.com/oboukili
git clone git@github.com:oboukili/terraform-provider-argocd
```

Then build the provider

```sh
cd $GOPATH/src/github.com/oboukili/terraform-provider-argocd
go build
```

### Running tests

The acceptance tests run against a disposable ArgoCD installation within a [Kind](https://github.com/kubernetes-sigs/kind) cluster. You will only need to have a running Docker daemon running as an additional prerequisite.

```sh
make testacc_prepare_env
make testacc
make testacc_clean_env
```

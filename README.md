# Terraform Provider for ArgoCD

![Acceptance Tests](https://github.com/oboukili/terraform-provider-argocd/workflows/Acceptance%20Tests/badge.svg)

---

## Compatibility promise

This provider is compatible with _at least_ the last 2 major releases of ArgoCD (e.g, ranging from 1.(n).m, to 1.(n-1).0, where `n` is the latest available major version).

Older releases are not supported and some resources may not work as expected.

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

* **From Terraform Public Registry (TF >= 0.13.0)**

  https://registry.terraform.io/providers/oboukili/argocd/latest
  ```hcl
    terraform {
      required_providers {
        argocd = {
          source = "oboukili/argocd"
          version = "0.4.7"
        }
      }
    }

    provider "argocd" {
      # Configuration options
    }
  ```


* **From binary releases (TF >= 0.12.0, < 0.13)**:
  Get the [latest release](https://github.com/oboukili/terraform-provider-argocd/releases/latest), or adapt and run the following:
  ```shell script
  curl -LO https://github.com/oboukili/terraform-provider-argocd/releases/download/v0.1.0/terraform-provider-argocd_v0.1.0_linux_amd64.gz
  gunzip -N terraform-provider-argocd_v0.1.0_linux_amd64.gz
  mv terraform-provider-argocd_v0.1.0 ~/.terraform.d/plugins/linux_amd64/
  chmod +x ~/.terraform.d/plugins/linux_amd64/terraform-provider-argocd_v0.1.0
  ```

* **From source**: Follow [the 'contributing' build instructions](https://github.com/oboukili/terraform-provider-argocd#building).

## Usage

```hcl
provider "argocd" {
  server_addr = "argocd.local:443" # env ARGOCD_SERVER
  auth_token  = "1234..."          # env ARGOCD_AUTH_TOKEN
  # username  = "admin"            # env ARGOCD_AUTH_USERNAME
  # password  = "foo"              # env ARGOCD_AUTH_PASSWORD
  insecure    = false              # env ARGOCD_INSECURE
}

resource "argocd_cluster" "kubernetes" {
  server = "https://1.2.3.4:12345"
  name   = "mycluster"

  config {
    bearer_token = "eyJhbGciOiJSUzI..."

    tls_client_config {
      ca_data = base64encode(file("path/to/ca.pem"))
      // insecure = true
    }
  }
}

resource "argocd_repository_credentials" "private" {
  url             = "git@private-git-repository.local"
  username        = "git"
  ssh_private_key = "-----BEGIN OPENSSH PRIVATE KEY-----\nfoo\nbar\n-----END OPENSSH PRIVATE KEY-----"
}

// Uses previously defined repository credentials
resource "argocd_repository" "private" {
  repo     = "git@private-git-repository.local:somerepo.git"
  // insecure = true
}

resource "argocd_repository" "public_nginx_helm" {
  repo = "https://helm.nginx.com/stable"
  name = "nginx-stable"
  type = "helm"
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
      clusters     = [
        "in-cluster",
        argocd_cluster.cluster.name,
      ]
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

resource "argocd_project_token" "secret" {
  count        = 20
  project      = argocd_project.myproject.metadata.0.name
  role         = "foobar"
  description  = "short lived token"
  expires_in   = "1h"
  renew_before = "30m"
}

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
      target_revision = "release-kustomize-v3.7"
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

**Note:** to speed up testing environment setup, it is highly recommended you pull all needed container images into your local registry first, as the setup tries to sideload the images within the Kind cluster upon cluster creation.

For example if you use Docker as your local container runtime:
```shell
docker pull argoproj/argocd:v1.8.3
docker pull ghcr.io/dexidp/dex:v2.27.0
docker pull redis:6.2.4-alpine
docker pull bitnami/redis:6.2.5
```

#### Troubleshooting during local development

* **"too many open files":** Running all acceptance tests in parallel (the default) may open a lot of files and sockets, therefore ensure your local workstation [open files/sockets limits are tuned accordingly](https://k6.io/docs/misc/fine-tuning-os).

---

## Credits

* Thanks to [JetBrains](https://www.jetbrains.com/?from=terraform-provider-argocd) for providing a GoLand open source license to support the development of this provider.
* Thanks to [Keplr](https://www.welcometothejungle.com/fr/companies/keplr) for allowing me to contribute to this side-project of mine during paid work hours.

![](sponsors/jetbrains.svg?display=inline-block) ![](sponsors/keplr.png?display=inline-block)

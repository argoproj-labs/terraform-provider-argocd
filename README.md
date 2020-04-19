# Terraform Provider for ArgoCD

![Acceptance Tests](https://github.com/oboukili/terraform-provider-argocd/workflows/Acceptance%20Tests/badge.svg)

---

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.12.x
- [Go](https://golang.org/doc/install) 1.14+

---

## Motivations

### *I thought ArgoCD already allowed for 100% declarative configuration?*

While that is true through the use of ArgoCD Kubernetes Custom Resources, 
there are some resources that simply cannot be managed using Kubernetes manifests,
such as project roles JWTs whose respective lifecycles are better handled by a tool like Terraform.
Even more so when you need to export these JWTs to another external system using Terraform, like a CI platform.

### *Wouldn't using a Kubernetes provider to handle ArgoCD configuration be enough?*

Using Terraform to manage Kubernetes Custom Resource becomes increasingly difficult 
the further you use HCL2 DSL to merge different data structures *and* want to preserve type safety.

Whatever the Kubernetes CRD provider you are using, you will probably end up using `locals` and the `yamlencode` function **which does not preserve the values' type**.
In these cases, not only the readability of your Terraform plan will worsen, but you will also be losing some safeties that Terraform provides in the process.

---

## Building

Clone the repository within your `GOPATH`

```sh
mkdir -p $GOPATH/src/github.com/oboukili; cd $GOPATH/src/github.com/oboukili
git clone git@github.com:oboukili/terraform-provider-argocd
```

Then build the provider

```sh
cd $GOPATH/src/github.com/oboukili/terraform-provider-argocd
make build
```

## Usage

```hcl
provider "argocd" {
  server_addr = "argocd.local:443" # env ARGOCD_SERVER
  auth_token  = "1234..."          # env ARGOCD_AUTH_TOKEN
  # username  = "admin"            # env ARGOCD_AUTH_USERNAME
  # password  = "foo"              # env ARGOCD_AUTH_PASSWORD
  insecure    = false              # env ARGOCD_INSECURE                 
}

resource "argocd_project_token" "secret" {
  project     = "myproject"
  role        = "bar"
  description = "short lived token"
  expires_in  = "3600"
}
```

## Developing the Provider

Contributions are welcome! You'll first need a working installation of [Go 1.14+](http://www.golang.org). Just as a reminder. you will also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH).

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
make build
$GOPATH/bin/terraform-provider-argocd
```

### Running tests

The acceptance tests run against a disposable ArgoCD installation within a [Kind](https://github.com/kubernetes-sigs/kind) cluster. You will only need to have a running Docker daemon running as an additional prerequisite.

```sh
make testacc_prepare_env
make testacc
make testacc_clean_env
```

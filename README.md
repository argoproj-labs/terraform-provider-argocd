# Terraform Provider for ArgoCD

![Acceptance Tests](https://github.com/oboukili/terraform-provider-argocd/workflows/Acceptance%20Tests/badge.svg)

---

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.12.x
- [Go](https://golang.org/doc/install) 1.14+

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

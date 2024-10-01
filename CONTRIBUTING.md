# Contributing

Contributions are welcome! 

## Building

1. `git clone` this repository and `cd` into its directory
2. `make build` will trigger the Golang build

The provided `GNUmakefile` defines additional commands generally useful during
development, like for running tests, generating documentation, code formatting
and linting. Taking a look at it's content is recommended.

## Testing

The acceptance tests run against a disposable ArgoCD installation within a
[Kind](https://github.com/kubernetes-sigs/kind) cluster. Other requirements are
having a Docker daemon running and
[Kustomize](https://kubectl.docs.kubernetes.io/installation/kustomize/)
installed.

```sh
make testacc_prepare_env
make testacc
make testacc_clean_env
```

## Generating documentation

This provider uses [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs/)
to generate documentation and store it in the `docs/` directory.
Once a release is cut, the Terraform Registry will download the documentation from `docs/`
and associate it with the release version. Read more about how this works on the
[official page](https://www.terraform.io/registry/providers/docs).

Use `make generate` to ensure the documentation is regenerated with any changes.

## Using a development build

If [running tests and acceptance tests](#testing) isn't enough, it's possible to
set up a local terraform configuration to use a development builds of the
provider. This can be achieved by leveraging the Terraform CLI [configuration
file development
overrides](https://www.terraform.io/cli/config/config-file#development-overrides-for-provider-developers).

First, use `make install` to place a fresh development build of the provider in
your
[`${GOBIN}`](https://pkg.go.dev/cmd/go#hdr-Compile_and_install_packages_and_dependencies)
(defaults to `${GOPATH}/bin` or `${HOME}/go/bin` if `${GOPATH}` is not set).
Repeat this every time you make changes to the provider locally.

Then, setup your environment following [these
instructions](https://www.terraform.io/plugin/debugging#terraform-cli-development-overrides)
to make your local terraform use your local build.

## Troubleshooting during local development

* **"too many open files":** Running all acceptance tests in parallel (the
  default) may open a lot of files and sockets, therefore ensure your local
  workstation [open files/sockets limits are tuned
  accordingly](https://k6.io/docs/misc/fine-tuning-os).

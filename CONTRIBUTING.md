# Contributing

Contributions are welcome! 

## Dependency Management

### K8s version
In our CI we test against a Kubernetes version that is supported by all Argo CD versions we support.

That version can be obtained when looking at [this table](https://argo-cd.readthedocs.io/en/stable/operator-manual/installation/#tested-versions) in the Argo CD documentation.

### Argo CD client-lib

Some dependencies we use are strictly aligned with the Argo CD client-lib that we use and should only be updated together:
- github.com/argoproj/gitops-engine
- k8s.io/*

Please don't update any of these dependencies without having discussed this first!

## Building

1. `git clone` this repository and `cd` into its directory
2. `make build` will trigger the Golang build

The provided `GNUmakefile` defines additional commands generally useful during
development, like for running tests, generating documentation, code formatting
and linting. Taking a look at its content is recommended.

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

## Debugging

Hasicorp Docs: https://developer.hashicorp.com/terraform/plugin/debugging#starting-a-provider-in-debug-mode

### Running the Terraform provider in debug mode

In VS Code open the Debug tab and select the profile "Debug Terraform Provider". Set some breakpoints and then run this task. 

Then head to the debug console and copy the line where it says `TF_REATTACH_PROVIDERS` and copy it.

Now that your provider is running in debug-mode in VS Code, you can head to any terminal where you want to run a Terraform stack and prepend the terraform command with the copied text. The Terraform CLI will then ensure it's using the provider already running inside VS Code.

### Running acceptance tests in debug mode

Open a test file, hover over a test function name and in the Debug tab hit "Debug selected Test". You shouldn't use the builtin "Debug Test" profile that is shown when hovering over a test function since it doesn't contain the necessary configuration to find your Argo CD environment.

## Troubleshooting during local development

* **"too many open files":** Running all acceptance tests in parallel (the
  default) may open a lot of files and sockets, therefore ensure your local
  workstation [open files/sockets limits are tuned
  accordingly](https://k6.io/docs/misc/fine-tuning-os).

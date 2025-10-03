# Contributing

Contributions are welcome! 

[![Contributors](https://img.shields.io/github/contributors/argoproj-labs/terraform-provider-argocd)](https://github.com/argoproj-labs/terraform-provider-argocd)
[![Last commit](https://img.shields.io/github/last-commit/argoproj-labs/terraform-provider-argocd)](https://github.com/argoproj-labs/terraform-provider-argocd)
[![Stars](https://img.shields.io/github/stars/argoproj-labs/terraform-provider-argocd)](https://github.com/argoproj-labs/hera/terraofrm-provider-argocd)

## New Contributor Guide

If you are a new contributor this section aims to show you everything you need to get started.

We especially welcome contributions to issues that are labeled with ["good-first-issue"](https://github.com/argoproj-labs/terraform-provider-argocd/issues?q=is%3Aopen%20is%3Aissue%20label%3A%22good%20first%20issue%22)
or
["help-wanted"](https://github.com/argoproj-labs/terraform-provider-argocd/issues?q=is%3Aopen%20is%3Aissue%20label%3A%22help%20wanted%22).

We also encourage contributions in the form of:
- bug/crash reports
- Answering questions on [Slack](https://cloud-native.slack.com/archives/C07PQF40SF8)
- Posting your use-case for the provider on [Slack](https://cloud-native.slack.com/archives/C07PQF40SF8) / Blog Post

### Setting up

To contribute to this Provider you need the following tools installed locally:
* [Go](https://go.dev/doc/install) (1.24)
* [GNU Make](https://www.gnu.org/software/make/)
* [Kustomize](https://kubectl.docs.kubernetes.io/installation/kustomize/)
* [Container runtime](https://java.testcontainers.org/supported_docker_environment/)
* [Kind](https://kind.sigs.k8s.io) (optional)
* [golangci-lint](https://golangci-lint.run/usage/install/#local-installation) (optional)

#### Codespaces

If you don't want to install tools locally you can use Github Codespaces to contribute to this project. We have a pre-configured codespace that should have all tools installed already:

[![Open in GitHub Codespaces](https://github.com/codespaces/badge.svg)](https://github.com/codespaces/new/argoproj-labs/terraform-provider-argocd)

## Contributing checklist

Please keep in mind the following guidelines and practices when contributing to the Provider:

1. Your commit must be signed (`git commit --signoff`). We use the [DCO application](https://github.com/apps/dco)
   that enforces the Developer Certificate of Origin (DCO) on commits.
1. Use `make fmt` to format the repository code. 
1. Use `make lint` to lint the project.
1. Use `make generate` to generate documentation on schema changes
1. Add unit tests for any new code you write.
1. Add an example, or extend an existing example in the [examples](./examples), with any new features you may add. Use `make generate` to add examples to the docs

## Building 

1. `git clone` this repository and `cd` into its directory
2. `make build` will trigger the Golang build and place it's binary in `<git-repo-path>/bin/terraform-provider-argocd`

The provided `GNUmakefile` defines additional commands generally useful during
development, like for running tests, generating documentation, code formatting
and linting. Taking a look at its content is recommended.

## Testing

The acceptance tests run against a disposable ArgoCD installation within a containerized-K3s cluster. We are using [testcontainers](https://testcontainers.com) for this. If you have a [supported container runtime](https://java.testcontainers.org/supported_docker_environment/) installed you can simply run the tests using:

```sh
make testacc # to run all the Terraform tests
make test # to only run helper unit tests (minority of the testcases)
```

## Documentation

This provider uses [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs/)
to generate documentation and store it in the `docs/` directory.
Once a release is cut, the Terraform Registry will download the documentation from `docs/`
and associate it with the release version. Read more about how this works on the
[official page](https://www.terraform.io/registry/providers/docs).

Use `make generate` to ensure the documentation is regenerated with any changes.

## Debugging

We have some pre-made config to debug and run the provider using VSCode. If you are using another IDE take a look at [Hashicorp's Debug docs](https://developer.hashicorp.com/terraform/plugin/debugging#starting-a-provider-in-debug-mode) for instructions or adapt [.vscode/launch.json](.vscode/launch.json) for your IDE

### Running the Terraform provider in debug mode (VSCode-specific)

To use the preconfigured debug config in VS Code open the Debug tab and select the profile "Debug Terraform Provider". Set some breakpoints and then run this task. 

Head to the debug console and copy the line where it says `TF_REATTACH_PROVIDERS` to the clipboard. 

Open a terminal session and export the `TF_REATTACH_PROVIDERS` variable in this session. Every Terraform CLI command in this terminal session will then ensure it's using the provider already running inside VS Code and attach to it.

Example of such a command:

```console
export TF_REATTACH_PROVIDERS='{"registry.terraform.io/argoproj-labs/argocd":{"Protocol":"grpc","ProtocolVersion":6,"Pid":2065,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/rj/_02y2jmn3k1bxx45wlzt2dkc0000gn/T/plugin193859953"}}}' 
terraform apply -auto-approve # will use the provider running in debug-mode
```

**Note**: if the provider crashes or you restart the debug-session you have to re-export this variable to your terminal for the Terraform CLI to find the already running provider!

### Running acceptance tests in debug mode (VSCode-specific)

Open a test file, **hover** over a test function's name and then in the Debug tab of VSCode select "Debug selected Test". This will run the test you selected with the specific arguments required for Terraform to run the acceptance test.

**Note**: You shouldn't use the builtin "Debug Test" profile that is shown when hovering over a test function since it doesn't contain the necessary configuration to find your Argo CD environment.

## Run Terraform using a local build

It's possible to set up a local terraform configuration to use a development build of the
provider. This can be achieved by leveraging the Terraform CLI [configuration
file development
overrides](https://www.terraform.io/cli/config/config-file#development-overrides-for-provider-developers).

First, use `make install` to place a fresh development build of the provider in
your
[`${GOBIN}`](https://pkg.go.dev/cmd/go#hdr-Compile_and_install_packages_and_dependencies)
(defaults to `${GOPATH}/bin` or `${HOME}/go/bin` if `${GOPATH}` is not set).
Repeat this every time you make changes to the provider locally.

Note: you can also use `make build` to place the binary into `<git-repo-path>/bin/terraform-provider-argocd` instead.


Then write this config to a file:
```hcl filename="../reproduce/.terraformrc"
provider_installation {
  dev_overrides {
    "argoproj-labs/argocd" = "/Users/username/go/bin" # path must be absolute and point to the directoy containing the binary
  }

  direct {}
}
```

And lastly use the following environment variable in a terminal session to tell Terraform to use this file for picking up the development binary:
```console
export TF_CLI_CONFIG_FILE=../.reproduce/.terraformrc
terraform plan # will not use the local provider build 
```

For further reference consult [HashiCorp's article](https://www.terraform.io/plugin/debugging#terraform-cli-development-overrides) about this topic.

## Dependency Management

### K8s version
In our CI we test against a Kubernetes version that is supported by all Argo CD versions we support.

That version can be obtained when looking at [this table](https://argo-cd.readthedocs.io/en/stable/operator-manual/installation/#tested-versions) in the Argo CD documentation.

### Argo CD client-lib

Some dependencies we use are strictly aligned with the Argo CD client-lib that we use and should only be updated together:
- github.com/argoproj/gitops-engine
- k8s.io/*

Please **don't update** any of these dependencies without having discussed this first!


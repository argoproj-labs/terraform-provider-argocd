# Contributing

Contributions are welcome! 

[![Contributors](https://img.shields.io/github/contributors/argoproj-labs/terraform-provider-argocd)](https://github.com/argoproj-labs/terraform-provider-argocd)
[![Last commit](https://img.shields.io/github/last-commit/argoproj-labs/terraform-provider-argocd)](https://github.com/argoproj-labs/terraform-provider-argocd)
[![Stars](https://img.shields.io/github/stars/argoproj-labs/terraform-provider-argocd)](https://github.com/argoproj-labs/terraform-provider-argocd)

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
* [Go](https://go.dev/doc/install) (see the `go` directive in [go.mod](./go.mod) for the required version)
* [GNU Make](https://www.gnu.org/software/make/)
* [Container runtime](https://java.testcontainers.org/supported_docker_environment/) (required for the acceptance tests)
* [Terraform](https://developer.hashicorp.com/terraform/install) (used by the acceptance tests; the test framework uses the one on your `PATH`)
* [golangci-lint](https://golangci-lint.run/docs/welcome/install/local/) (optional)

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

The acceptance tests run against a disposable Argo CD installation within a containerized K3s cluster. We are using [testcontainers](https://testcontainers.com) for this, so all you need is a [supported container runtime](https://java.testcontainers.org/supported_docker_environment/) and network access (images and the upstream Argo CD manifests are pulled). Then run:

```sh
make test # unit tests only; acceptance tests are skipped and no containers are started
make testacc # all tests, including the acceptance tests
make testacc TEST_FILTER=TestAccArgoCDRepository_ProjectChange # only run tests matching the (regex) filter
```

How it works:
* The fixture is set up in `TestMain` (see [internal/testhelpers](./internal/testhelpers)) and only when `TF_ACC` is set (which `make testacc` does). It starts K3s, installs Argo CD from the kustomize resources in [manifests](./manifests) and waits until it is ready. This takes a few minutes.
* Every test package that contains acceptance tests (`argocd` and `internal/provider`) starts its own cluster, so `make testacc` creates two in parallel. Note that `go test` buffers the output of each package until it has finished, so you won't see any output while the fixture is starting.
* The Argo CD and K3s versions are defined in the `GNUmakefile` and can be overridden, e.g. `make testacc ARGOCD_VERSION=v3.4.8`. The connection details (server address, admin credentials) are taken from the fixture, you don't need to configure anything else.
* The fixture container is removed again when the tests finish.

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

This starts the test fixture (see [Testing](#testing)) just like `make testacc`. The Argo CD and K3s versions are duplicated in [.vscode/launch.json](.vscode/launch.json) (and in [.run](.run) for IntelliJ), keep them in sync with the `GNUmakefile`.

**Note**: You shouldn't use the builtin "Debug Test" profile that is shown when hovering over a test function since it doesn't set `TF_ACC` and the versions, so the test fixture is not started.

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
In our CI we test against a Kubernetes version (`K3S_VERSION` in the `GNUmakefile`) that is supported by all Argo CD versions we support. When changing it, also update the copies in [.vscode/launch.json](.vscode/launch.json) and [.run](.run).

A supported version can be obtained when looking at [this table](https://argo-cd.readthedocs.io/en/stable/operator-manual/installation/#tested-versions) in the Argo CD documentation.

### Argo CD client-lib

Some dependencies we use are strictly aligned with the Argo CD client-lib that we use and should only be updated together:
- github.com/argoproj/argo-cd/gitops-engine
- k8s.io/*

Please **don't update** any of these dependencies without having discussed this first!


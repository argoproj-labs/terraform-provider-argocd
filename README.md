<a href="https://terraform.io">
    <img src=".github/tf.png" alt="Terraform logo" title="Terraform" align="left" height="50" />
</a>

<a href="https://argoproj.github.io/cd">
    <img src=".github/argo-cd.png" alt="Terraform logo" title="Terraform" align="right" height="50" />
</a>

# Terraform Provider for ArgoCD

[![Tests](https://github.com/argoproj-labs/terraform-provider-argocd/actions/workflows/tests.yml/badge.svg)](https://github.com/argoproj-labs/terraform-provider-argocd/actions/workflows/tests.yml)

The [ArgoCD Terraform
Provider](https://registry.terraform.io/providers/argoproj-labs/argocd/latest/docs)
provides lifecycle management of
[ArgoCD](https://argo-cd.readthedocs.io/en/stable/) resources.

**NB**: The provider is not concerned with the installation/configuration of
ArgoCD itself. To make use of the provider, you will need to have an existing
ArgoCD deployment and, the ArgoCD API server must be
[accessible](https://argo-cd.readthedocs.io/en/stable/getting_started/#3-access-the-argo-cd-api-server)
from where you are running Terraform.

---

## Documentation

Official documentation on how to use this provider can be found on the
[Terraform
Registry](https://registry.terraform.io/providers/argoproj-labs/argocd/latest/docs).

## Version Compatibility 

The provider supports all versions Argo-CD itself currently supports. See the [Argo CD documentation](https://argo-cd.readthedocs.io/en/stable/operator-manual/installation/#supported-versions) for supported versions.

The following table shows which provider version to use given your Argo CD version:

| Argo CD version | minimum provider version | maximum provider version |
|-----------------|--------------------------|--------------------------|
| 2.14            | 7.5.0                    | not yet available        |
| 2.13            | 7.4.0                    | not yet available        |
| 2.12            | 7.2.0                    | not yet available        |
| 2.11            | 7.1.0                    | 7.4.0                    |
| 2.10            | 6.1.0                    | 7.3.1                    |

Note: these numbers are based on our testing matrix that tests the provider against these versions of Argo CD. You may be able to use other provider versions as the ones listed here and it may still work.

## Motivations

### *I thought ArgoCD already allowed for 100% declarative configuration?*

While that is true through the use of ArgoCD Kubernetes Custom Resources, there
are some resources that simply cannot be managed using Kubernetes manifests,
such as project roles JWTs whose respective lifecycles are better handled by a
tool like Terraform. Even more so when you need to export these JWTs to another
external system using Terraform, like a CI platform.

### *Wouldn't using a Kubernetes provider to handle ArgoCD configuration be enough?*

Existing Kubernetes providers do not patch arrays of objects, losing project
role JWTs when doing small project changes just happen.

ArgoCD Kubernetes admission webhook controller is not as exhaustive as ArgoCD
API validation, this can be seen with RBAC policies, where no validation occur
when creating/patching a project.

Using Terraform to manage Kubernetes Custom Resource becomes increasingly
difficult the further you use HCL2 DSL to merge different data structures *and*
want to preserve type safety.

Whatever the Kubernetes CRD provider you are using, you will probably end up
using `locals` and the `yamlencode` function **which does not preserve the
values' type**. In these cases, not only the readability of your Terraform plan
will worsen, but you will also be losing some safeties that Terraform provides
in the process.

## Upgrading

### Migrate provider source `oboukili` -> `argoproj-labs`

As announced in the releases [v6.2.0] and [v7.0.0], we moved the provider from "github.com/**oboukili**/terraform-provider-argocd/" 
to "github.com/**argoproj-labs**/terraform-provider-argocd". Users need to migrate their Terraform state according to
HashiCorps [replace-provider] docs. In summary, you can do the following:

1. List currently used providers

    ```bash
    $ terraform providers

    Providers required by configuration:
    .
    ├── provider[registry.terraform.io/hashicorp/helm] 2.15.0
    ├── (..)
    └── provider[registry.terraform.io/oboukili/argocd] 6.1.1

    Providers required by state:

        (..)

        provider[registry.terraform.io/oboukili/argocd]

        provider[registry.terraform.io/hashicorp/helm]
    ```

2. **If you see** the provider "registry.terraform.io/**oboukili**/argocd", you can update the provider specification:

    ```diff
    --- a/versions.tf
    +++ b/versions.tf
    @@ -5,7 +5,7 @@ terraform {
        }
        argocd = {
    -      source  = "oboukili/argocd"
    +      source  = "argoproj-labs/argocd"
        version = "6.1.1"
        }
        helm = {
    ```

3. Download the new provider via `terraform init`:

    ```bash
    $ terraform init
    Initializing HCP Terraform...
    Initializing provider plugins...
    - Finding (..)
    - Finding oboukili/argocd versions matching "6.1.1"...
    - Finding latest version of argoproj-labs/argocd...
    - (..)
    - Installing oboukili/argocd v6.1.1...
    - Installed oboukili/argocd v6.1.1 (self-signed, key ID 09A6EABF546E8638)
    - Installing argoproj-labs/argocd v7.0.0...
    - Installed argoproj-labs/argocd v7.0.0 (self-signed, key ID 6421DA8DFD8F48D0)
    (..)

    HCP Terraform has been successfully initialized!

    (..)
    ```

4. Then, execute the migration via `terraform state replace-provider`:

    ```bash
    $ terraform state replace-provider registry.terraform.io/oboukili/argocd registry.terraform.io/argoproj-labs/argocd
    Terraform will perform the following actions:

    ~ Updating provider:
        - registry.terraform.io/oboukili/argocd
        + registry.terraform.io/argoproj-labs/argocd

    Changing 5 resources:

    argocd_project.apps_with_clusterroles
    argocd_application.app_of_apps
    argocd_project.base
    argocd_project.apps_restricted
    argocd_project.core_services_unrestricted

    Do you want to make these changes?
    Only 'yes' will be accepted to continue.

    Enter a value: yes

    Successfully replaced provider for 5 resources.
    ```

5. You have successfully migrated

## Requirements

* [Terraform](https://www.terraform.io/downloads) (>= 1.0)
* [Go](https://go.dev/doc/install) (1.24)
* [GNU Make](https://www.gnu.org/software/make/)
* [golangci-lint](https://golangci-lint.run/usage/install/#local-installation) (optional)

## Credits

* We would like to thank [Olivier Boukili] for creating this awesome Terraform provider and moving the project over to
  [argoproj-labs] on Apr 5th 2024.

[argoproj-labs]: https://github.com/argoproj-labs
[Olivier Boukili]: https://github.com/oboukili
[v6.2.0]: https://github.com/argoproj-labs/terraform-provider-argocd/releases/tag/v6.2.0
[v7.0.0]: https://github.com/argoproj-labs/terraform-provider-argocd/releases/tag/v7.0.0
[replace-provider]: https://developer.hashicorp.com/terraform/cli/commands/state/replace-provider

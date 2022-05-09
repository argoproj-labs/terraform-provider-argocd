# argocd_cluster

Creates an ArgoCD cluster.

## Example Usage - Bearer token

```hcl
resource "argocd_cluster" "kubernetes" {
  server = "https://1.2.3.4:12345"

  config {
    bearer_token = "eyJhbGciOiJSUzI..."

    tls_client_config {
      ca_data = file("path/to/ca.pem")
      // ca_data = "-----BEGIN CERTIFICATE-----\nfoo\nbar\n-----END CERTIFICATE-----"
      // ca_data = base64decode("LS0tLS1CRUdJTiBDRVJUSUZ...")

      // insecure = true
    }
  }
}
```

## Example Usage - GCP GKE cluster

```hcl
data "google_container_cluster" "cluster" {
  name     = "cluster"
  location = "europe-west1"
}

# Create the service account, cluster role + binding, which ArgoCD expects to be present in the targeted cluster
resource "kubernetes_service_account" "argocd_manager" {
  metadata {
    name      = "argocd-manager"
    namespace = "kube-system"
  }
}

resource "kubernetes_cluster_role" "argocd_manager" {
  metadata {
    name = "argocd-manager-role"
  }

  rule {
    api_groups = ["*"]
    resources  = ["*"]
    verbs      = ["*"]
  }

  rule {
    non_resource_urls = ["*"]
    verbs             = ["*"]
  }
}

resource "kubernetes_cluster_role_binding" "argocd_manager" {
  metadata {
    name = "argocd-manager-role-binding"
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = kubernetes_cluster_role.argocd_manager.metadata.0.name
  }

  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account.argocd_manager.metadata.0.name
    namespace = kubernetes_service_account.argocd_manager.metadata.0.namespace
  }
}

data "kubernetes_secret" "argocd_manager" {
  metadata {
    name      = kubernetes_service_account.argocd_manager.default_secret_name
    namespace = kubernetes_service_account.argocd_manager.metadata.0.namespace
  }
}

resource "argocd_cluster" "gke" {
  server = format("https://%s", data.google_container_cluster.cluster.endpoint)
  name   = "gke"

  config {
    bearer_token = data.kubernetes_secret.argocd_manager.data["token"]
    tls_client_config {
      ca_data      = base64decode(data.google_container_cluster.cluster.master_auth.0.cluster_ca_certificate)
    }
  }
}
```

## Example Usage - AWS EKS cluster

```hcl
data "aws_eks_cluster" "cluster" {
  name = "cluster"
}

resource "argocd_cluster" "eks" {
  server     = format("https://%s", data.aws_eks_cluster.cluster.endpoint)
  name       = "eks"
  namespaces = ["default", "optional"]

  config {
    aws_auth_config {
      cluster_name = "myekscluster"
      role_arn     = "arn:aws:iam::<123456789012>:role/<role-name>"
    }
    tls_client_config {
      ca_data = data.aws_eks_cluster.cluster.certificate_authority[0].data
    }
  }
}
```

## Argument Reference

* `server` - (Required) Server is the API server URL of the Kubernetes cluster.
* `name` - (Optional) Name of the cluster. If omitted, will use the server address.
* `shard` - (Optional) Shard contains optional shard number. Calculated on the fly by the application controller if not specified.
* `namespaces` - (Optional) Holds list of namespaces which are accessible in that cluster. Cluster level resources would be ignored if namespace list is not empty..
* `config` - (Optional) The configuration specification, nested attributes are documented below.
* `metadata` - (Optional) Cluster metadata, nested attributes are documented below.
* `project` - (Optional) Scope cluster to ArgoCD project. If omitted, cluster will be global. Requires ArgoCD 2.2.0 onwards.

The `config` block can have the following attributes:

* `aws_auth_config` - (Optional) AWS EKS specific IAM authentication. Structure is documented below.
* `bearer_token` - (Optional) OAuth2 bearer token. ArgoCD client will not attempt to use refresh tokens for an OAuth2 flow.
* `exec_provider_config` - (Optional) configuration used to call an external command to perform cluster authentication See: https://godoc.org/k8s.io/client-go/tools/clientcmd/api#ExecConfig. Structure is documented below.
* `tls_client_config` - (Optional) TLS client configuration. Structure is documented below.
* `username` - (Optional)
* `password` - (Optional)

The `config.aws_auth_config` block can have the following attributes:

* `cluster_name` - (Optional) Name of the EKS cluster.
* `role_arn` - (Optional) RoleARN contains optional role ARN. If set then AWS IAM Authenticator assume a role to perform cluster operations instead of the default AWS credential provider chain.

The `config.exec_provider_config` can have the following attributes:

* `api_version` - (Optional) Preferred input version of the ExecInfo.
* `command` - (Optional) Command to execute.
* `args` - (Optional) list of string. Arguments to pass to the command when executing it.
* `env` - (Optional) map of string. Defines additional environment variables to expose to the process.
* `install_hint` - (Optional) This text is shown to the user when the executable doesn't seem to be present.

The `config.tls_client_config` block can have the following attributes:

* `ca_data` - (Optional) string. Holds PEM-encoded bytes (typically read from a root certificates bundle).
* `cert_data` - (Optional) string. Holds PEM-encoded bytes (typically read from a client certificate file).
* `key_data` - (Optional) string. Holds PEM-encoded bytes (typically read from a client certificate key file).
* `insecure` - (Optional) boolean. For when the server should be accessed without verifying the TLS certificate.
* `server_name` - (Optional) string. Passed to the server for SNI and is used in the client to check server certificates against. If empty, the hostname used to contact the server is used.

The `metadata` block can have the following attributes:

* `annotations` - (Optional) An unstructured key value map stored with the config map that may be used to store arbitrary metadata. **By default, the provider ignores any annotations whose key names end with kubernetes.io. This is necessary because such annotations can be mutated by server-side components and consequently cause a perpetual diff in the Terraform plan output. If you explicitly specify any such annotations in the configuration template then Terraform will consider these as normal resource attributes and manage them as expected (while still avoiding the perpetual diff problem)**. For more info see [Kubernetes reference](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/).
* `labels` - (Optional) Map of string keys and values that can be used to organize and categorize (scope and select) the config map. May match selectors of replication controllers and services. **By default, the provider ignores any labels whose key names end with kubernetes.io. This is necessary because such labels can be mutated by server-side components and consequently cause a perpetual diff in the Terraform plan output. If you explicitly specify any such labels in the configuration template then Terraform will consider these as normal resource attributes and manage them as expected (while still avoiding the perpetual diff problem).** For more info see [Kubernetes reference](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/).

## Attribute Reference

* `info.0.server_version` - The version of the remote Kubernetes cluster.
* `info.0.applications_count` - How many ArgoCD applications the cluster currently holds.
* `info.0.connection_state.0.message`
* `info.0.connection_state.0.status`

## Import

ArgoCD clusters can be imported using an id consisting of `{server}`, e.g.
```
$ terraform import argocd_cluster.mycluster https://mycluster.io:443
```

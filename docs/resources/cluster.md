# argocd_cluster

Creates an ArgoCD cluster.

## Example Usage - Bearer token

```hcl
resource "argocd_cluster" "kubernetes" {
  server = "https://1.2.3.4:12345"

  config {
    bearer_token = "eyJhbGciOiJSUzI..."

    tls_client_config {
      ca_data = base64encode(file("path/to/ca.pem"))
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

resource "argocd_cluster" "gke" {
  server = data.google_container_cluster.cluster.endpoint
  name   = "gke"

  config {
    tls_client_config {
      ca_cert_data = data.google_container_cluster.cluster.master_auth.0.cluster_ca_certificate
      cert_data    = data.google_container_cluster.cluster.master_auth.0.client_certificate
      key_data     = data.google_container_cluster.cluster.master_auth.0.client_key
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
  server     = data.aws_eks_cluster.cluster.endpoint
  name       = "eks"
  namespaces = ["default", "optional"]

  config {
    aws_auth_config {
      cluster_name = "myekscluster"
      role_arn     = "arn:aws:iam::<123456789012>:role/<role-name>"
    }
    tls_client_config {
      ca_cert_data = data.aws_eks_cluster.cluster.certificate_authority[0].data 
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

## Attribute Reference

* `info.0.server_version` - The version of the remote Kubernetes cluster.
* `info.0.applications_count` - How many ArgoCD applications the cluster currently holds.
* `info.0.connection_state.0.message`
* `info.0.connection_state.0.status` 

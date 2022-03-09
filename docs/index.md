# ArgoCD Provider

A Terraform provider for [ArgoCD](https://argoproj.github.io/argo-cd/).

## Example Usage

```hcl
provider "argocd" {
  server_addr = "argocd.local:443"
  auth_token  = "1234..."
}
```

## Argument Reference

* `server_addr` - (Required) ArgoCD server address with port.
* `use_local_config` - (Optional) use the authentication settings found in the local config file. Useful when you have previously logged in using SSO. Conflicts with
`auth_token`, `username` and `password`.
* `config_path` (Optional) - Override the default config path of `$HOME/.config/argocd/config`. Only relevant when using `use_local_config` above.
  Can be set through the `ARGOCD_CONFIG_PATH` environment variable.
* `auth_token` - (Optional) ArgoCD authentication token, takes precedence over `username`/`password`. Can be set through the `ARGOCD_AUTH_TOKEN` environment variable.
* `username` - (Optional) authentication username. Can be set through the `ARGOCD_AUTH_USERNAME` environment variable.
* `password` - (Optional) authentication password. Can be set through the `ARGOCD_AUTH_PASSWORD` environment variable.
* `cert_file` - (Optional) Additional root CA certificates file to add to the client TLS connection pool. 
* `plain_text` - (Optional) Boolean, whether to initiate an unencrypted connection to ArgoCD server. 
* `context` - (Optional) Kubernetes context to load from an existing `.kube/config` file. Can be set through `ARGOCD_CONTEXT` environment variable.
* `user_agent` - (Optional)
* `grpc_web` - (Optional) Whether to use gRPC web proxy client. Useful if Argo CD server is behind proxy which does not support HTTP2.
* `grpc_web_root_path` - (Optional) Use the gRPC web proxy client and set the web root, e.g. `argo-cd`. Useful if the Argo CD server is behind a proxy at a non-root path.
* `port_forward` - (Optional)
* `port_forward_with_namespace` - (Optional)
* `headers` - (Optional) Additional headers to add to each request to the ArgoCD server.
* `insecure` - (Optional) Whether to skip TLS server certificate. Can be set through the `ARGOCD_INSECURE` environment variable.
* `kubernetes` - Kubernetes configuration block.

The `kubernetes` block supports:

* `config_path` - (Optional) Path to the kube config file. Can be sourced from `KUBE_CONFIG_PATH`.
* `config_paths` - (Optional) A list of paths to the kube config files. Can be sourced from `KUBE_CONFIG_PATHS`.
* `host` - (Optional) The hostname (in form of URI) of the Kubernetes API. Can be sourced from `KUBE_HOST`.
* `username` - (Optional) The username to use for HTTP basic authentication when accessing the Kubernetes API. Can be sourced from `KUBE_USER`.
* `password` - (Optional) The password to use for HTTP basic authentication when accessing the Kubernetes API. Can be sourced from `KUBE_PASSWORD`.
* `token` - (Optional) The bearer token to use for authentication when accessing the Kubernetes API. Can be sourced from `KUBE_TOKEN`.
* `insecure` - (Optional) Whether server should be accessed without verifying the TLS certificate. Can be sourced from `KUBE_INSECURE`.
* `client_certificate` - (Optional) PEM-encoded client certificate for TLS authentication. Can be sourced from `KUBE_CLIENT_CERT_DATA`.
* `client_key` - (Optional) PEM-encoded client certificate key for TLS authentication. Can be sourced from `KUBE_CLIENT_KEY_DATA`.
* `cluster_ca_certificate` - (Optional) PEM-encoded root certificates bundle for TLS authentication. Can be sourced from `KUBE_CLUSTER_CA_CERT_DATA`.
* `config_context` - (Optional) Context to choose from the config file. Can be sourced from `KUBE_CTX`.
* `exec` - (Optional) Configuration block to use an [exec-based credential plugin](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#client-go-credential-plugins), e.g. call an external command to receive user credentials.
    * `api_version` - (Required) API version to use when decoding the ExecCredentials resource, e.g. `client.authentication.k8s.io/v1beta1`.
    * `command` - (Required) Command to execute.
    * `args` - (Optional) List of arguments to pass when executing the plugin.
    * `env` - (Optional) Map of environment variables to set when executing the plugin.
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
* `auth_token` - (Optional) ArgoCD authentication token, taked precedence over `username`/`password`. Can be set through the `ARGOCD_AUTH_TOKEN` environment variable.
* `username` - (Optional) authentication username. Can be set through the `ARGOCD_AUTH_USERNAME` environment variable.
* `password` - (Optional) authentication password. Can be set through the `ARGOCD_AUTH_PASSWORD` environment variable.
* `cert_file` - (Optional) Additional root CA certificates file to add to the client TLS connection pool. 
* `plain_text` - (Optional) Boolean, whether to initiate an unencrypted connection to ArgoCD server. 
* `context` - (Optional) Kubernetes context to load from an existing `.kube/config` file. Can be set through `ARGOCD_CONTEXT` environment variable.
* `user_agent` - (Optional)
* `grpc_web` - (Optional) Whether to use gRPC web proxy client.
* `port_forward` - (Optional)
* `port_forward_with_namespace` - (Optional)
* `headers` - (Optional) Additional headers to add to each request to the ArgoCD server.
* `insecure` - (Optional) Whether to skip TLS server certificate. Can be set through the `ARGOCD_INSECURE` environment variable.

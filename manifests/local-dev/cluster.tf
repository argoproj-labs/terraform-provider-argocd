resource "kind_cluster" "secondary" {
  name       = "secondary"
  node_image = "kindest/node:v1.24.7"
}

resource "argocd_cluster" "kind_secondary" {
  name   = "kind-secondary"
  server = kind_cluster.secondary.endpoint

  config {
    tls_client_config {
      ca_data = kind_cluster.secondary.cluster_ca_certificate
      // insecure = true
    }
  }
}

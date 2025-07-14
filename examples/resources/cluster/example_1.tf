## Bearer token Authentication
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

## GCP GKE cluster
data "google_container_cluster" "cluster" {
  name     = "cluster"
  location = "europe-west1"
}

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
      ca_data = base64decode(data.google_container_cluster.cluster.master_auth.0.cluster_ca_certificate)
    }
  }
}

## AWS EKS cluster
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
      ca_data = base64decode(data.aws_eks_cluster.cluster.certificate_authority[0].data)
    }
  }
}

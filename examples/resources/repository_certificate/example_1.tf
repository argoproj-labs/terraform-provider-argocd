# HTTPS certificate
resource "argocd_repository_certificate" "private-git-repository" {
  https {
    server_name = "private-git-repository.local"
    cert_data   = <<EOT
-----BEGIN CERTIFICATE-----\nfoo\nbar\n-----END CERTIFICATE-----
EOT
  }
}

# SSH certificate
resource "argocd_repository_certificate" "private-git-repository" {
  ssh {
    server_name  = "private-git-repository.local"
    cert_subtype = "ssh-rsa"
    cert_data    = <<EOT
AAAAB3NzaC1yc2EAAAADAQABAAABgQCiPZAufKgxwRgxP9qy2Gtub0FI8qJGtL8Ldb7KatBeRUQQPn8QK7ZYjzYDvP1GOutFMaQT0rKIqaGImIBsztNCno...
EOT
  }
}

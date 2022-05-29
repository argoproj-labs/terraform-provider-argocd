# argocd_certificate

Creates an ArgoCD certificate, for use with future or existing private repositories.

## Example Usage

### Example ssh certificate
```hcl
// Private repository ssh certificate
resource "argocd_certificate" "private-git-repository" {
	ssh {
		server_name  = "private-git-repository.local"
		cert_subtype = "ssh-rsa"
		cert_data    = <<EOT
AAAAB3NzaC1yc2EAAAADAQABAAABgQCiPZAufKgxwRgxP9qy2Gtub0FI8qJGtL8Ldb7KatBeRUQQPn8QK7ZYjzYDvP1GOutFMaQT0rKIqaGImIBsztNCno...
EOT
	}
}

resource "argocd_repository" "private" {
  repo = "git@private-git-repository.local:somerepo.git"
}
```

### Example https certificate
```hcl
resource "argocd_certificate" "private-git-repository" {
	https {
		server_name  = "private-git-repository.local"
		cert_data    = <<EOT
-----BEGIN CERTIFICATE-----\nfoo\nbar\n-----END CERTIFICATE-----
EOT
	}
}

resource "argocd_repository" "private" {
  repo = "https://private-git-repository.local/somerepo.git"
}
```

## Argument Reference

* `https` - (Optional), for a https certificate, the nested attributes are documented below.
* `ssh` - (Optional), for a ssh certificate, the nested attributes are documented below.

### https

* `server_name` - (Required), string, specifies the DNS name of the server this certificate is intended for.
* `cert_data` - (Required), string, contains the actual certificate data, dependent on the certificate type.

### ssh

* `server_name` - (Required), string, specifies the DNS name of the server this certificate is intended for.
* `cert_subtype` - (Required), string, specifies the sub type of the cert, i.e. "ssh-rsa".
* `cert_data` - (Required), string, contains the actual certificate data, dependent on the certificate type.

## Attribute Reference

### https
* `https.0.cert_subtype` - contains the sub type of the cert, i.e. "ssh-rsa"
* `https.0.cert_info` - holds additional certificate info (e.g. X509 CommonName, etc).

### ssh
* `ssh.0.cert_info` - holds additional certificate info (e.g. SSH fingerprint, etc).

# Repositories can be imported using the repository URL.

# Note: as the ArgoCD API does not return any sensitive information, a
# subsequent `terraform apply` should be executed to make the `password`,
# `ssh_private_key` and `tls_client_cert_key` attributes converge to their
# expected values defined within the plan.

terraform import argocd_repository.myrepo git@private-git-repository.local:somerepo.git

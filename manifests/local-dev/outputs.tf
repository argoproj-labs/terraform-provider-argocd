output "project_auth_token" {
  description = "Proj token"
  value = argocd_project_token.foo_token.jwt
  sensitive = true
}
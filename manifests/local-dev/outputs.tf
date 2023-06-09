# Uncomment output below to see `status` for application `foo`.
#
# Note: making use of `status` within Terraform code will result in Terraform
# displaying information regarding changes to this field that occur between
# Terraform runs. I.e. we can expect to see the following in the plan details if
# the status changes:
# 
# ```
# Note: Objects have changed outside of Terraform
# 
# Terraform detected the following changes made outside of Terraform since the last "terraform apply" which may have affected this plan:
# ```
# 
# Tip: post apply run the following for better formatting: 
# ``` sh
# terraform output -raw foo_status | jq .
# ```
# 
# output "foo_status" {
#   value = jsonencode(argocd_application.foo.status.0)
# }

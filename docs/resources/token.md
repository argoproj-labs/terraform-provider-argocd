# argocd_project_token

Creates an ArgoCD role project JSON Web Token.

## Example Usage

```hcl
resource "argocd_project_token" "secret" {
  project      = "someproject"
  role         = "foobar"
  description  = "short lived token"
  expires_in   = "1h"
  renew_before = "30m"
}
```

## Argument Reference

* `project` - (Required) The project name associated with the token.
* `role` - (Required) The project role associated with the token, the role must exist beforehand.
* `description` - (Optional)
* `expires_in` - (Optional) An expiration duration, valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
* `renew_before` - (Optional) duration to control token silent regeneration, valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". If `expires_in` is set, Terraform will regenerate the token if `expires_in - renew_before < currentDate`.

## Attribute Reference

* `jwt` - The raw JWT as a string.
* `issued_at` - Unix timestamp upon which the token was issued at, as a string.
* `expires_at` - If `expires_in` is set, Unix timestamp upon which the token will expire, as a string.

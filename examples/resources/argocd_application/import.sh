# ArgoCD applications can be imported using an id consisting of `{name}:{namespace}`.

terraform import argocd_application.myapp myapp:argocd

# The namespace segment may be omitted (or left empty, e.g. `myapp:`), in which
# case the application is looked up by name only, without restricting to a
# specific namespace. See
# https://argo-cd.readthedocs.io/en/stable/operator-manual/app-any-namespace/
# for how ArgoCD resolves an application's namespace.

terraform import argocd_application.myapp myapp
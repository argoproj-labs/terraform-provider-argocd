package validators

import (
	"testing"

	"github.com/argoproj/argo-cd/v3/util/rbac"
	"github.com/stretchr/testify/assert"
)

func TestValidateProjectPolicy(t *testing.T) {
	t.Parallel()

	const (
		project = "myproject"
		role    = "admin"
	)

	tests := map[string]struct {
		policy      string
		expectError bool
	}{
		"valid policy": {
			policy: "p, proj:myproject:admin, applications, get, myproject/*, allow",
		},
		"valid applicationsets policy": {
			policy: "p, proj:myproject:admin, applicationsets, get, myproject/*, allow",
		},
		"valid clusters policy": {
			policy: "p, proj:myproject:admin, clusters, get, myproject/*, allow",
		},
		"valid repositories policy": {
			policy: "p, proj:myproject:admin, repositories, get, myproject/*, allow",
		},
		"valid exec policy": {
			policy: "p, proj:myproject:admin, exec, create, myproject/*, allow",
		},
		"valid logs policy": {
			policy: "p, proj:myproject:admin, logs, get, myproject/*, allow",
		},
		"valid deny effect": {
			policy: "p, proj:myproject:admin, applications, delete, myproject/*, deny",
		},
		"valid wildcard action": {
			policy: "p, proj:myproject:admin, applications, *, myproject/*, allow",
		},
		"valid action subresource": {
			policy: "p, proj:myproject:admin, applications, action/apps/Deployment/restart, myproject/*, allow",
		},
		"valid fine grained delete subresource": {
			policy: "p, proj:myproject:admin, applications, delete/*/Pod/*/*, myproject/*, allow",
		},
		"valid namespaced object with update subresource": {
			policy: "p, proj:myproject:admin, applications, update/*, myproject/*/multi-ns, allow",
		},
		"invalid format - not enough components": {
			policy:      "p, proj:myproject:admin, applications, get",
			expectError: true,
		},
		"invalid subject": {
			policy:      "p, proj:otherproject:admin, applications, get, myproject/*, allow",
			expectError: true,
		},
		"invalid resource": {
			policy:      "p, proj:myproject:admin, invalidResource, get, myproject/*, allow",
			expectError: true,
		},
		// 'projects' is a valid resource in the global RBAC config map but is not
		// project-scoped, so Argo CD rejects it inside a project role policy.
		"projects is not project scoped": {
			policy:      "p, proj:myproject:admin, projects, get, myproject/*, allow",
			expectError: true,
		},
		"accounts is not project scoped": {
			policy:      "p, proj:myproject:admin, accounts, get, myproject/*, allow",
			expectError: true,
		},
		"certificates is not project scoped": {
			policy:      "p, proj:myproject:admin, certificates, get, myproject/*, allow",
			expectError: true,
		},
		"extensions is not project scoped": {
			policy:      "p, proj:myproject:admin, extensions, invoke, myproject/*, allow",
			expectError: true,
		},
		"invalid action": {
			policy:      "p, proj:myproject:admin, applications, invalid, myproject/*, allow",
			expectError: true,
		},
		"invalid object format": {
			policy:      "p, proj:myproject:admin, applications, get, otherproject/*, allow",
			expectError: true,
		},
		"invalid effect": {
			policy:      "p, proj:myproject:admin, applications, get, myproject/*, maybe",
			expectError: true,
		},
		"object with valid app name": {
			policy: "p, proj:myproject:admin, applications, get, myproject/app-01, allow",
		},
		"object with valid ns/app combo": {
			policy: "p, proj:myproject:admin, applications, get, myproject/default/app-01, allow",
		},
		"object with valid ns wildcard": {
			policy: "p, proj:myproject:admin, applications, get, myproject/default/*, allow",
		},
		"object with dash and dot in name": {
			policy: "p, proj:myproject:admin, applications, get, myproject/app-1.2, allow",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := validateProjectPolicy(project, role, test.policy)

			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateProjectPolicyAcceptsEveryProjectScopedResource guards against the
// provider's accepted resource set drifting away from Argo CD's.
func TestValidateProjectPolicyAcceptsEveryProjectScopedResource(t *testing.T) {
	t.Parallel()

	for res := range rbac.ProjectScoped {
		t.Run(res, func(t *testing.T) {
			t.Parallel()

			policy := "p, proj:myproject:admin, " + res + ", get, myproject/*, allow"
			assert.NoError(t, validateProjectPolicy("myproject", "admin", policy))
		})
	}
}

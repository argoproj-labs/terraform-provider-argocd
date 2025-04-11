package argocd

import (
	"testing"
)

func TestValidatePolicy(t *testing.T) {
	project := "myproject"
	role := "admin"

	tests := []struct {
		name        string
		policy      string
		expectError bool
	}{
		{
			name:        "Valid policy",
			policy:      "p, proj:myproject:admin, applications, get, myproject/*, allow",
			expectError: false,
		},
		{
			name:        "Valid applicationsets policy",
			policy:      "p, proj:myproject:admin, applicationsets, get, myproject/*, allow",
			expectError: false,
		},
		{
			name:        "Invalid format - not enough components",
			policy:      "p, proj:myproject:admin, applications, get",
			expectError: true,
		},
		{
			name:        "Invalid subject",
			policy:      "p, proj:otherproject:admin, applications, get, myproject/*, allow",
			expectError: true,
		},
		{
			name:        "Invalid resource",
			policy:      "p, proj:myproject:admin, invalidResource, get, myproject/*, allow",
			expectError: true,
		},
		{
			name:        "Invalid action",
			policy:      "p, proj:myproject:admin, applications, invalid, myproject/*, allow",
			expectError: true,
		},
		{
			name:        "Invalid object format",
			policy:      "p, proj:myproject:admin, applications, get, otherproject/*, allow",
			expectError: true,
		},
		{
			name:        "Invalid effect",
			policy:      "p, proj:myproject:admin, applications, get, myproject/*, maybe",
			expectError: true,
		},
		{
			name:        "Object with valid app name",
			policy:      "p, proj:myproject:admin, applications, get, myproject/app-01, allow",
			expectError: false,
		},
		{
			name:        "Object with dash and dot in name",
			policy:      "p, proj:myproject:admin, applications, get, myproject/app-1.2, allow",
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validatePolicy(project, role, tc.policy)
			if (err != nil) != tc.expectError {
				t.Errorf("validatePolicy() error = %v, expectError = %v", err, tc.expectError)
			}
		})
	}
}

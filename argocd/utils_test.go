package argocd

import (
	"testing"
)

func TestParseNameNamespaceID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		id          string
		expectName  string
		expectNS    string
		expectDiags bool
	}{
		{
			name:        "Valid id",
			id:          "myapp:argocd",
			expectName:  "myapp",
			expectNS:    "argocd",
			expectDiags: false,
		},
		{
			name:        "Id with more than one colon",
			id:          "myapp:my:ns",
			expectDiags: true,
		},
		{
			name:        "Missing namespace - as set by a bare import block id",
			id:          "myapp",
			expectDiags: true,
		},
		{
			name:        "Empty id",
			id:          "",
			expectDiags: true,
		},
		{
			name:        "Missing name",
			id:          ":argocd",
			expectDiags: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			name, namespace, diags := parseNameNamespaceID("application", tc.id)
			if (diags != nil) != tc.expectDiags {
				t.Fatalf("parseNameNamespaceID() diags = %v, expectDiags = %v", diags, tc.expectDiags)
			}

			if tc.expectDiags {
				return
			}

			if name != tc.expectName || namespace != tc.expectNS {
				t.Errorf("parseNameNamespaceID() = (%q, %q), want (%q, %q)", name, namespace, tc.expectName, tc.expectNS)
			}
		})
	}
}

package argocd

import (
	"context"
	"testing"

	"github.com/argoproj/argo-cd/v3/pkg/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSetPortForwardingOpts_ServerName(t *testing.T) {
	tests := []struct {
		name       string
		serverName types.String
		envValue   string
		expected   string
	}{
		{
			name:       "defaults to argocd-server when unset",
			serverName: types.StringNull(),
			expected:   "argocd-server",
		},
		{
			name:       "uses explicit provider configuration",
			serverName: types.StringValue("argocd-argo-cd-server"),
			expected:   "argocd-argo-cd-server",
		},
		{
			name:       "falls back to ARGOCD_SERVER_NAME environment variable",
			serverName: types.StringNull(),
			envValue:   "argocd-argo-cd-server",
			expected:   "argocd-argo-cd-server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				t.Setenv("ARGOCD_SERVER_NAME", tt.envValue)
			}

			p := ArgoCDProviderConfig{
				PortForward: types.BoolValue(true),
				ServerName:  tt.serverName,
			}

			opts := &apiclient.ClientOptions{PortForward: true}

			enabled, diags := p.setPortForwardingOpts(context.Background(), opts)
			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags.Errors())
			}

			if !enabled {
				t.Fatal("expected port forwarding to be enabled")
			}

			if opts.ServerName != tt.expected {
				t.Errorf("expected ServerName %q, got %q", tt.expected, opts.ServerName)
			}
		})
	}
}

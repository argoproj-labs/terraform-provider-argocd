package argocd

import (
	"fmt"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/version"
	"github.com/oboukili/terraform-provider-argocd/internal/features"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	semverEquals = iota
	semverGreater
	semverLess
)

func serverInterfaceTestData(t *testing.T, argocdVersion string, semverOperator int) *ServerInterface {
	v, err := semver.NewVersion(argocdVersion)
	require.NoError(t, err)
	require.True(t, v.Major() >= 1)

	switch semverOperator {
	case semverEquals:
	case semverGreater:
		inc := v.IncMajor()
		v = &inc

		assert.NoError(t, err)
	case semverLess:
		v, err = semver.NewVersion(
			fmt.Sprintf("%d.%d.%d",
				v.Major()-1,
				v.Minor(),
				v.Patch(),
			))
		assert.NoError(t, err)
	default:
		t.Error("unsupported semver test semverOperator")
	}

	vm := &version.VersionMessage{
		Version: v.String(),
	}

	return &ServerInterface{
		ApiClient:            nil,
		ServerVersion:        v,
		ServerVersionMessage: vm,
	}
}

func TestServerInterface_isFeatureSupported(t *testing.T) {
	t.Parallel()

	type args struct {
		feature features.Feature
	}

	tests := []struct {
		name string
		si   *ServerInterface
		args args
		want bool
	}{
		{
			name: "featureExecLogsPolicy-2.7.2",
			si:   serverInterfaceTestData(t, "2.7.2", semverEquals),
			args: args{feature: features.ExecLogsPolicy},
			want: true,
		},
		{
			name: "featureExecLogsPolicy-2.7.2+",
			si:   serverInterfaceTestData(t, "2.7.2", semverGreater),
			args: args{feature: features.ExecLogsPolicy},
			want: true,
		},
		{
			name: "featureExecLogsPolicy-2.7.2-",
			si:   serverInterfaceTestData(t, "2.7.2", semverLess),
			args: args{feature: features.ExecLogsPolicy},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.si.isFeatureSupported(tt.args.feature)

			if got != tt.want {
				t.Errorf("isFeatureSupported() got = %v, want %v, version %s",
					got,
					tt.want,
					tt.si.ServerVersion.String(),
				)
			}
		})
	}
}

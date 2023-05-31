package argocd

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/version"
	"github.com/stretchr/testify/assert"
	"modernc.org/mathutil"
)

const (
	semverEquals = iota
	semverGreater
	semverLess
)

func serverInterfaceTestData(t *testing.T, argocdVersion string, semverOperator int) *ServerInterface {
	v, err := semver.NewVersion(argocdVersion)
	assert.NoError(t, err)

	incPatch := rand.Int63n(100)
	incMinor := rand.Int63n(100)
	incMajor := rand.Int63n(100)

	switch semverOperator {
	case semverEquals:
	case semverGreater:
		v, err = semver.NewVersion(
			fmt.Sprintf("%d.%d.%d",
				v.Major()+incMajor,
				v.Minor()+incMinor,
				v.Patch()+incPatch,
			))
		assert.NoError(t, err)
	case semverLess:
		v, err = semver.NewVersion(
			fmt.Sprintf("%d.%d.%d",
				mathutil.MaxInt64(v.Major()-incMajor, 0),
				mathutil.MaxInt64(v.Minor()-incMinor, 0),
				mathutil.MaxInt64(v.Patch()-incPatch, 0),
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
		feature int
	}

	tests := []struct {
		name    string
		si      *ServerInterface
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "featureExecLogsPolicy-2.7.2",
			si:      serverInterfaceTestData(t, "2.7.2", semverEquals),
			args:    args{feature: featureExecLogsPolicy},
			want:    true,
			wantErr: false,
		},
		{
			name:    "featureExecLogsPolicy-2.7.2+",
			si:      serverInterfaceTestData(t, "2.7.2", semverGreater),
			args:    args{feature: featureExecLogsPolicy},
			want:    true,
			wantErr: false,
		},
		{
			name:    "featureExecLogsPolicy-2.7.2-",
			si:      serverInterfaceTestData(t, "2.7.2", semverLess),
			args:    args{feature: featureExecLogsPolicy},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.si.isFeatureSupported(tt.args.feature)
			if (err != nil) != tt.wantErr {
				t.Errorf("isFeatureSupported() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}

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

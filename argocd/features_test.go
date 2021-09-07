package argocd

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/version"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"modernc.org/mathutil"
	"testing"
)

const (
	semverEquals = iota
	semverGreater
	semverLess
)

func serverInterfaceTestData(t *testing.T, argocdVersion string, semverOperator int) ServerInterface {
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
	return ServerInterface{
		ApiClient:            nil,
		ServerVersion:        v,
		ServerVersionMessage: vm,
	}
}

func TestServerInterface_isFeatureSupported(t *testing.T) {
	type args struct {
		feature int
	}
	tests := []struct {
		name    string
		fields  ServerInterface
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "featureTokenID-1.5.3",
			fields:  serverInterfaceTestData(t, "1.5.3", semverEquals),
			args:    args{feature: featureTokenIDs},
			want:    true,
			wantErr: false,
		},
		{
			name:    "featureTokenID-1.5.3+",
			fields:  serverInterfaceTestData(t, "1.5.3", semverGreater),
			args:    args{feature: featureTokenIDs},
			want:    true,
			wantErr: false,
		},
		{
			name:    "featureTokenID-1.5.3-",
			fields:  serverInterfaceTestData(t, "1.5.3", semverLess),
			args:    args{feature: featureTokenIDs},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ServerInterface{
				ApiClient:            tt.fields.ApiClient,
				ServerVersion:        tt.fields.ServerVersion,
				ServerVersionMessage: tt.fields.ServerVersionMessage,
			}
			got, err := p.isFeatureSupported(tt.args.feature)
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
					tt.fields.ServerVersion.String(),
				)
			}
		})
	}
}

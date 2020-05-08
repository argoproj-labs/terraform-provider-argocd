package argocd

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/argoproj/argo-cd/pkg/apiclient/version"
	"math/rand"
	"testing"
)

const (
	semverEquals = iota
	semverGreater
	semverLess
)

func serverInterfaceTestData(argocdVersion string, semverOperator int) ServerInterface {

	v, err := semver.NewVersion(argocdVersion)
	if err != nil {
		panic(err)
	}
	incPatch := rand.Int63n(100)
	incMinor := rand.Int63n(100)
	incMajor := rand.Int63n(100)

	switch semverOperator {
	case semverEquals:
	case semverGreater:
		if v, err = semver.NewVersion(
			fmt.Sprintf("%d.%d.%d",
				v.Major()+incMajor,
				v.Minor()+incMinor,
				v.Patch()+incPatch,
			)); err != nil {
			panic(err)
		}

	case semverLess:
		if v, err = semver.NewVersion(
			fmt.Sprintf("%d.%d.%d",
				v.Major()-incMajor%v.Major(),
				v.Minor()-incMinor%v.Minor(),
				v.Patch()-incPatch%v.Patch(),
			)); err != nil {
			panic(err)
		}
	default:
		panic("unsupported semver test semverOperator")
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
			fields:  serverInterfaceTestData("1.5.3", semverEquals),
			args:    args{feature: featureTokenIDs},
			want:    true,
			wantErr: false,
		},
		{
			name:    "featureTokenID-1.5.3+",
			fields:  serverInterfaceTestData("1.5.3", semverGreater),
			args:    args{feature: featureTokenIDs},
			want:    true,
			wantErr: false,
		},
		{
			name:    "featureTokenID-1.5.3-",
			fields:  serverInterfaceTestData("1.5.3", semverLess),
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
				t.Errorf("isFeatureSupported() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("isFeatureSupported() got = %v, want %v", got, tt.want)
			}
		})
	}
}

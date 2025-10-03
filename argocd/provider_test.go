package argocd

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/argoproj-labs/terraform-provider-argocd/internal/features"
	"github.com/argoproj-labs/terraform-provider-argocd/internal/provider"
	"github.com/argoproj-labs/terraform-provider-argocd/internal/testhelpers"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var testAccProviders map[string]func() (*schema.Provider, error)
var testAccProtoV6ProviderFactories map[string]func() (tfprotov6.ProviderServer, error)

func init() {
	testAccProviders = map[string]func() (*schema.Provider, error){
		"argocd": func() (*schema.Provider, error) { //nolint:unparam
			return Provider(), nil
		},
	}

	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"argocd": func() (tfprotov6.ProviderServer, error) {
			ctx := context.Background()

			upgradedSdkServer, err := tf5to6server.UpgradeServer(
				ctx,
				Provider().GRPCProvider,
			)
			if err != nil {
				return nil, err
			}

			providers := []func() tfprotov6.ProviderServer{
				providerserver.NewProtocol6(provider.New("test")),
				func() tfprotov6.ProviderServer {
					return upgradedSdkServer
				},
			}

			muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)
			if err != nil {
				return nil, err
			}

			return muxServer.ProviderServer(), nil
		},
	}
}

func TestMain(m *testing.M) {
	testhelpers.TestMain(m)
}

func TestProvider(t *testing.T) {
	t.Parallel()

	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_headers(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf("%s %s", `
                provider "argocd" {
                    headers = [
                        "Hello: HiThere",
                    ]
                }`, testAccArgoCDApplicationSimple(acctest.RandomWithPrefix("test-acc"), "0.33.0", false),
				),
			},
		},
	})
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("ARGOCD_AUTH_USERNAME"); v == "" {
		t.Fatal("ARGOCD_AUTH_USERNAME must be set for acceptance tests")
	}

	if v := os.Getenv("ARGOCD_AUTH_PASSWORD"); v == "" {
		t.Fatal("ARGOCD_AUTH_PASSWORD must be set for acceptance tests")
	}

	if v := os.Getenv("ARGOCD_SERVER"); v == "" {
		t.Fatal("ARGOCD_SERVER must be set for acceptance tests")
	}

	if v := os.Getenv("ARGOCD_INSECURE"); v == "" {
		t.Fatal("ARGOCD_INSECURE should be set for acceptance tests")
	}
}

// Skip test if feature is not supported
func testAccPreCheckFeatureSupported(t *testing.T, feature features.Feature) {
	v := os.Getenv("ARGOCD_VERSION")
	if v == "" {
		t.Skip("ARGOCD_VERSION must be set set for feature supported acceptance tests")
	}

	serverVersion, err := semver.NewVersion(v)
	if err != nil {
		t.Fatalf("could not parse ARGOCD_VERSION as semantic version: %s", v)
	}

	fc, ok := features.ConstraintsMap[feature]
	if !ok {
		t.Fatal("feature constraint is not handled by the provider")
	}

	if i := fc.MinVersion.Compare(serverVersion); i == 1 {
		t.Skipf("version %s does not support feature", v)
	}
}

// Skip test if feature is supported
// Note: unused at present but left in the code in case it is needed again in future
// func testAccPreCheckFeatureNotSupported(t *testing.T, feature int) {
// 	v := os.Getenv("ARGOCD_VERSION")
// 	if v == "" {
// 		t.Skip("ARGOCD_VERSION must be set for feature supported acceptance tests")
// 	}

// 	serverVersion, err := semver.NewVersion(v)
// 	if err != nil {
// 		t.Fatalf("could not parse ARGOCD_VERSION as semantic version: %s", v)
// 	}

// 	versionConstraint, ok := featureVersionConstraintsMap[feature]
// 	if !ok {
// 		t.Fatal("feature constraint is not handled by the provider")
// 	}

// 	if i := versionConstraint.Compare(serverVersion); i != 1 {
// 		t.Skipf("not running test if feature is already supported (%s)", v)
// 	}
// }

package provider

import (
	"os"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/argoproj-labs/terraform-provider-argocd/internal/features"
	"github.com/argoproj-labs/terraform-provider-argocd/internal/testhelpers"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"argocd": providerserver.NewProtocol6WithError(New("test")),
}

func TestMain(m *testing.M) {
	testhelpers.TestMain(m)
}

func TestProvider_headers(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "argocd" {
						headers = [
							"Hello: HiThere",
						]
					}`,
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

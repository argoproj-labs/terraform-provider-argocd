package argocd

import (
	"fmt"
	"os"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/oboukili/terraform-provider-argocd/internal/features"
)

var testAccProviders map[string]func() (*schema.Provider, error)

func init() {
	testAccProviders = map[string]func() (*schema.Provider, error){
		"argocd": func() (*schema.Provider, error) { //nolint:unparam
			return Provider(), nil
		},
	}
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
                }`, testAccArgoCDApplicationSimple(acctest.RandomWithPrefix("test-acc"), "9.4.1", false),
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
// 		t.Skip("ARGOCD_VERSION must be set set for feature supported acceptance tests")
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

package argocd

import (
	"os"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]func() (*schema.Provider, error)

func init() {
	testAccProviders = map[string]func() (*schema.Provider, error){
		"argocd": func() (*schema.Provider, error) {
			return Provider(), nil
		},
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ = Provider()
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

func testAccPreCheckFeatureSupported(t *testing.T, feature int) {
	v := os.Getenv("ARGOCD_VERSION")
	if v == "" {
		t.Skip("ARGOCD_VERSION must be set set for feature supported acceptance tests")
	}
	serverVersion, err := semver.NewVersion(v)
	if err != nil {
		t.Fatalf("could not parse ARGOCD_VERSION as semantic version: %s", v)
	}
	versionConstraint, ok := featureVersionConstraintsMap[feature]
	if !ok {
		t.Fatal("feature constraint is not handled by the provider")
	}
	if i := versionConstraint.Compare(serverVersion); i == 1 {
		t.Skipf("version %s does not support feature", v)
	}
}

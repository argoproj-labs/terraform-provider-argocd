package argocd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"math/rand"
	"strconv"
	"testing"
)

func TestAccArgoCDProjectToken(t *testing.T) {
	//project := acctest.RandomWithPrefix("test-acc")
	//role := acctest.RandomWithPrefix("test-role")
	count := 2 + rand.Intn(8)
	expiresIn := rand.Int63n(100000)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDProjectTokenSimple(),
				Check: resource.TestCheckResourceAttrSet(
					"argocd_project_token.simple",
					"issued_at",
				),
			},
			{
				Config: testAccArgoCDProjectTokenExpiry(expiresIn),
				Check: testCheckTokenExpiresAt(
					"argocd_project_token.expires",
					expiresIn,
				),
			},
			{
				Config: testAccArgoCDProjectTokenMultiple(count),
				Check: testCheckMultipleResourceAttrSet(
					"argocd_project_token.multiple",
					"issued_at",
					count,
				),
			},
		},
	})
}

func testAccArgoCDProjectTokenSimple() string {
	return `
resource "argocd_project_token" "simple" {
  project = "myproject"
  role    = "test-role1234"
}
`
}

func testAccArgoCDProjectTokenExpiry(expiresIn int64) string {
	return fmt.Sprintf(`
resource "argocd_project_token" "expires" {
  project = "myproject"
  role    = "test-role1234"
  expires_in = %d
}
`, expiresIn)
}

func testAccArgoCDProjectTokenMultiple(count int) string {
	return fmt.Sprintf(`
resource "argocd_project_token" "multiple" {
  count   = %d
  project = "myproject"
  role    = "test-role1234"
}
`, count)
}

func testCheckTokenExpiresAt(resourceName string, expiresIn int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("token ID is not set")
		}
		_expiresAt, ok := rs.Primary.Attributes["expires_at"]
		if !ok {
			return fmt.Errorf("expires_at is not set")
		}
		_issuedAt, ok := rs.Primary.Attributes["issued_at"]
		if !ok {
			return fmt.Errorf("issued_at is not set")
		}
		expiresAt, err := strconv.ParseInt(_expiresAt, 10, 64)
		if err != nil {
			return err
		}
		issuedAt, err := strconv.ParseInt(_issuedAt, 10, 64)
		if err != nil {
			return err
		}
		if issuedAt+expiresIn != expiresAt {
			return fmt.Errorf("issuedAt + expiresIn != expiresAt : %d + %d != %d", issuedAt, expiresIn, expiresAt)
		}
		return nil
	}
}

func testCheckMultipleResourceAttrSet(name, key string, count int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for i := 0; i < count; i++ {
			ms := s.RootModule()
			_name := fmt.Sprintf("%s.%d", name, i)
			rs, ok := ms.Resources[_name]
			if !ok {
				return fmt.Errorf("not found: %s in %s", _name, ms.Path)
			}
			is := rs.Primary
			if is == nil {
				return fmt.Errorf("no primary instance: %s in %s", _name, ms.Path)
			}
			if val, ok := is.Attributes[key]; !ok || val == "" {
				return fmt.Errorf("%s: Attribute '%s' expected to be set", _name, key)
			}
		}
		return nil
	}
}

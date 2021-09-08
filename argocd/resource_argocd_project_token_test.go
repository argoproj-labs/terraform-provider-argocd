package argocd

import (
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAccArgoCDProjectToken(t *testing.T) {
	expiresInDurationFunc := func(i int) time.Duration {
		d, err := time.ParseDuration(fmt.Sprintf("%ds", i))
		assert.NoError(t, err)
		return d
	}
	count := 3 + rand.Intn(7)
	expIn1 := expiresInDurationFunc(rand.Intn(100000))
	expIn2 := expiresInDurationFunc(rand.Intn(100000))
	expIn3 := expiresInDurationFunc(rand.Intn(100000))
	expIn4 := expiresInDurationFunc(rand.Intn(100000))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDProjectTokenSimple(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_project_token.simple",
						"issued_at",
					),
					testCheckTokenIssuedAt(
						"argocd_project_token.simple",
					),
				),
			},
			{
				Config: testAccArgoCDProjectTokenExpiry(int64(expIn1.Seconds())),
				Check: testCheckTokenExpiresAt(
					"argocd_project_token.expires",
					int64(expIn1.Seconds()),
				),
			},
			{
				Config:      testAccArgoCDProjectTokenMisconfiguration(expIn2),
				ExpectError: regexp.MustCompile("token will expire within 5 minutes, check your settings"),
			},
			{
				Config: testAccArgoCDProjectTokenRenewBeforeSuccess(expIn3),
				Check: resource.ComposeTestCheckFunc(
					testCheckTokenExpiresAt(
						"argocd_project_token.renew",
						int64(expIn3.Seconds()),
					),
					resource.TestCheckResourceAttrSet(
						"argocd_project_token.renew",
						"renew_before",
					),
				),
			},
			{
				Config:      testAccArgoCDProjectTokenRenewBeforeFailure(expIn4),
				ExpectError: regexp.MustCompile("renew_before .* cannot be greater than expires_in .*"),
			},
			{
				Config: testAccArgoCDProjectTokenMultiple(count),
				Check: resource.ComposeTestCheckFunc(
					testCheckMultipleResourceAttrSet(
						"argocd_project_token.multiple1a",
						"issued_at",
						count,
					),
					testCheckMultipleResourceAttrSet(
						"argocd_project_token.multiple1b",
						"issued_at",
						count,
					),
					testCheckMultipleResourceAttrSet(
						"argocd_project_token.multiple2a",
						"issued_at",
						count,
					),
					testCheckMultipleResourceAttrSet(
						"argocd_project_token.multiple2b",
						"issued_at",
						count,
					),
				),
			},
		},
	})
}

func testAccArgoCDProjectTokenSimple() string {
	return `
resource "argocd_project_token" "simple" {
  project = "myproject1"
  role    = "test-role1234"
}
`
}

func testAccArgoCDProjectTokenExpiry(expiresIn int64) string {
	return fmt.Sprintf(`
resource "argocd_project_token" "expires" {
  project = "myproject1"
  role    = "test-role1234"
  expires_in = "%ds"
}
`, expiresIn)
}

func testAccArgoCDProjectTokenMultiple(count int) string {
	return fmt.Sprintf(`
resource "argocd_project_token" "multiple1a" {
  count   = %d
  project = "myproject1"
  role    = "test-role1234"
}

resource "argocd_project_token" "multiple1b" {
  count   = %d
  project = "myproject1"
  role    = "test-role4321"
}

resource "argocd_project_token" "multiple2a" {
  count   = %d
  project = "myproject2"
  role    = "test-role1234"
}

resource "argocd_project_token" "multiple2b" {
  count   = %d
  project = "myproject2"
  role    = "test-role4321"
}

`, count, count, count, count)
}

func testAccArgoCDProjectTokenMisconfiguration(expiresInDuration time.Duration) string {
	expiresIn := int64(expiresInDuration.Seconds())
	renewBefore := expiresIn

	return fmt.Sprintf(`
resource "argocd_project_token" "renew" {
  project = "myproject1"
  role    = "test-role1234"
  expires_in = "%ds"
  renew_before = "%ds"
}
`, expiresIn, renewBefore)
}

func testAccArgoCDProjectTokenRenewBeforeSuccess(expiresInDuration time.Duration) string {
	expiresIn := int64(expiresInDuration.Seconds())
	renewBefore := int64(math.Min(
		expiresInDuration.Seconds()-1,
		expiresInDuration.Seconds()-(rand.Float64()*expiresInDuration.Seconds()),
	)) % int64(expiresInDuration.Seconds())

	return fmt.Sprintf(`
resource "argocd_project_token" "renew" {
  project = "myproject1"
  role    = "test-role1234"
  expires_in = "%ds"
  renew_before = "%ds"
}
`, expiresIn, renewBefore)
}

func testAccArgoCDProjectTokenRenewBeforeFailure(expiresInDuration time.Duration) string {
	expiresIn := int64(expiresInDuration.Seconds())
	renewBefore := int64(math.Max(
		expiresInDuration.Seconds()+1.0,
		expiresInDuration.Seconds()+(rand.Float64()*expiresInDuration.Seconds()),
	))
	return fmt.Sprintf(`
resource "argocd_project_token" "renew" {
  project = "myproject1"
  role    = "test-role1234"
  expires_in = "%ds"
  renew_before = "%ds"
}
`, expiresIn, renewBefore)
}

func testCheckTokenIssuedAt(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("token ID is not set")
		}
		_issuedAt, ok := rs.Primary.Attributes["issued_at"]
		if !ok {
			return fmt.Errorf("testCheckTokenExpiresAt: issued_at is not set")
		}
		_, err := convertStringToInt64(_issuedAt)
		if err != nil {
			return fmt.Errorf("testCheckTokenExpiresAt: string attribute 'issued_at' stored in state cannot be converted to int64: %s", err)
		}
		return nil
	}
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
			return fmt.Errorf("testCheckTokenExpiresAt: issued_at is not set")
		}
		expiresAt, err := convertStringToInt64(_expiresAt)
		if err != nil {
			return fmt.Errorf("testCheckTokenExpiresAt: string attribute 'expires_at' stored in state cannot be converted to int64: %s", err)
		}
		issuedAt, err := convertStringToInt64(_issuedAt)
		if err != nil {
			return fmt.Errorf("testCheckTokenExpiresAt: string attribute 'issued_at' stored in state cannot be converted to int64: %s", err)
		}
		if issuedAt+expiresIn != expiresAt {
			return fmt.Errorf("testCheckTokenExpiresAt: issuedAt + expiresIn != expiresAt : %d + %d != %d", issuedAt, expiresIn, expiresAt)
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

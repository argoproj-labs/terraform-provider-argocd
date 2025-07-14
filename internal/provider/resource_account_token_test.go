package provider

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAccArgoCDAccountTokenResource_DefaultAccount(t *testing.T) {
	expIn1, err := time.ParseDuration(fmt.Sprintf("%ds", rand.Intn(100000)))
	assert.NoError(t, err)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDAccountTokenResource_DefaultAccount(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_account_token.this",
						"issued_at",
					),
					testCheckTokenIssuedAt(
						"argocd_account_token.this",
					),
				),
			},
			{
				Config: testAccArgoCDAccountTokenResource_Expiry(int64(expIn1.Seconds())),
				Check: testCheckTokenExpiresAt(
					"argocd_account_token.this",
					int64(expIn1.Seconds()),
				),
			},
		},
	})
}

func TestAccArgoCDAccountTokenResource_ExplicitAccount(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDAccountTokenResource_ExplicitAccount(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_account_token.this",
						"issued_at",
					),
					resource.TestCheckResourceAttr(
						"argocd_account_token.this",
						"account",
						"admin",
					),
					testCheckTokenIssuedAt(
						"argocd_account_token.this",
					),
				),
			},
		},
	})
}

func TestAccArgoCDAccountTokenResource_Multiple(t *testing.T) {
	count := 3 + rand.Intn(7)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDAccountTokenResource_Multiple(count),
				Check: resource.ComposeTestCheckFunc(
					testTokenIssuedAtSet(
						"argocd_account_token.multiple1a",
						count,
					),
					testTokenIssuedAtSet(
						"argocd_account_token.multiple1b",
						count,
					),
					testTokenIssuedAtSet(
						"argocd_account_token.multiple2a",
						count,
					),
					testTokenIssuedAtSet(
						"argocd_account_token.multiple2b",
						count,
					),
				),
			},
		},
	})
}

func TestAccArgoCDAccountTokenResource_RenewBefore(t *testing.T) {
	resourceName := "argocd_account_token.renew_before"

	expiresInSeconds := 30
	expiresIn := fmt.Sprintf("%ds", expiresInSeconds)
	expiresInDuration, _ := time.ParseDuration(expiresIn)

	renewBeforeSeconds := expiresInSeconds - 1

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDAccountTokenResourceRenewBeforeSuccess(expiresIn, "20s"),
				Check: resource.ComposeTestCheckFunc(
					testCheckTokenExpiresAt(resourceName, int64(expiresInDuration.Seconds())),
					resource.TestCheckResourceAttr(resourceName, "renew_before", "20s"),
				),
			},
			{
				Config: testAccArgoCDAccountTokenResourceRenewBeforeSuccess(expiresIn, fmt.Sprintf("%ds", renewBeforeSeconds)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "renew_before", fmt.Sprintf("%ds", renewBeforeSeconds)),
					testDelay(renewBeforeSeconds+1),
				),
				ExpectNonEmptyPlan: true, // token should be recreated when refreshed at end of step due to delay above
			},
			{
				Config:      testAccArgoCDAccountTokenResourceRenewBeforeFailure(expiresInDuration),
				ExpectError: regexp.MustCompile("renew_before .* cannot be greater than expires_in .*"),
			},
		},
	})
}

func TestAccArgoCDAccountTokenResource_RenewAfter(t *testing.T) {
	resourceName := "argocd_account_token.renew_after"
	renewAfterSeconds := 30

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDAccountTokenResourceRenewAfter(renewAfterSeconds),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "renew_after", fmt.Sprintf("%ds", renewAfterSeconds)),
				),
			},
			{
				Config: testAccArgoCDAccountTokenResourceRenewAfter(renewAfterSeconds),
				Check: resource.ComposeTestCheckFunc(
					testDelay(renewAfterSeconds + 1),
				),
				ExpectNonEmptyPlan: true, // token should be recreated when refreshed at end of step due to delay above
			},
			{
				Config: testAccArgoCDAccountTokenResourceRenewAfter(renewAfterSeconds),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "renew_after", fmt.Sprintf("%ds", renewAfterSeconds)),
				),
			},
		},
	})
}

// Test helper functions

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
			return fmt.Errorf("testCheckTokenIssuedAt: issued_at is not set")
		}

		if _issuedAt == "" {
			return fmt.Errorf("testCheckTokenIssuedAt: issued_at is empty")
		}

		return nil
	}
}

func testCheckTokenExpiresAt(resourceName string, _ int64) resource.TestCheckFunc {
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

		if _expiresAt == "" || _issuedAt == "" {
			return fmt.Errorf("testCheckTokenExpiresAt: expires_at or issued_at is empty")
		}

		return nil
	}
}

func testTokenIssuedAtSet(name string, count int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for i := 0; i < count; i++ {
			ms := s.RootModule()
			_name := fmt.Sprintf("%s.%d", name, i)

			rs, ok := ms.Resources[_name]
			if !ok {
				return fmt.Errorf("not found: %s", _name)
			}

			if rs.Primary.ID == "" {
				return fmt.Errorf("token ID is not set for %s", _name)
			}

			_issuedAt, ok := rs.Primary.Attributes["issued_at"]
			if !ok {
				return fmt.Errorf("issued_at is not set for %s", _name)
			}

			if _issuedAt == "" {
				return fmt.Errorf("issued_at is empty for %s", _name)
			}
		}

		return nil
	}
}

func testDelay(seconds int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		time.Sleep(time.Duration(seconds) * time.Second)
		return nil
	}
}

// Test configuration functions

func testAccArgoCDAccountTokenResource_DefaultAccount() string {
	return `
resource "argocd_account_token" "this" {}
`
}

func testAccArgoCDAccountTokenResource_Expiry(expiresIn int64) string {
	return fmt.Sprintf(`
resource "argocd_account_token" "this" {
	expires_in = "%ds"
}
`, expiresIn)
}

func testAccArgoCDAccountTokenResource_ExplicitAccount() string {
	return `
resource "argocd_account_token" "this" {
	account = "admin"
}
`
}

func testAccArgoCDAccountTokenResource_Multiple(count int) string {
	return fmt.Sprintf(`
resource "argocd_account_token" "multiple1a" {
	count = %d
}

resource "argocd_account_token" "multiple1b" {
	count = %d
}

resource "argocd_account_token" "multiple2a" {
	account = "admin"
	count   = %d
}

resource "argocd_account_token" "multiple2b" {
	account = "admin"
	count   = %d
}

`, count, count, count, count)
}

func testAccArgoCDAccountTokenResourceRenewBeforeSuccess(expiresIn, renewBefore string) string {
	return fmt.Sprintf(`
resource "argocd_account_token" "renew_before" {
	expires_in   = "%s"
	renew_before = "%s"
}
`, expiresIn, renewBefore)
}

func testAccArgoCDAccountTokenResourceRenewBeforeFailure(expiresInDuration time.Duration) string {
	expiresIn := int64(expiresInDuration.Seconds())
	renewBefore := int64(expiresInDuration.Seconds() + 1.0)

	return fmt.Sprintf(`
resource "argocd_account_token" "renew_before" {
	expires_in   = "%ds"
	renew_before = "%ds"
}
`, expiresIn, renewBefore)
}

func testAccArgoCDAccountTokenResourceRenewAfter(renewAfter int) string {
	return fmt.Sprintf(`
resource "argocd_account_token" "renew_after" {
	account     = "admin"
	renew_after = "%ds"
}
`, renewAfter)
}

package argocd

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"
)

func TestAccArgoCDAccountToken_DefaultAccount(t *testing.T) {
	expIn1, err := time.ParseDuration(fmt.Sprintf("%ds", rand.Intn(100000)))
	assert.NoError(t, err)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDAccountToken_DefaultAccount(),
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
				Config: testAccArgoCDAccountToken_Expiry(int64(expIn1.Seconds())),
				Check: testCheckTokenExpiresAt(
					"argocd_account_token.this",
					int64(expIn1.Seconds()),
				),
			},
		},
	})
}

func TestAccArgoCDAccountToken_ExplicitAccount(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDAccountToken_ExplicitAccount(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_account_token.this",
						"issued_at",
					),
					resource.TestCheckResourceAttr(
						"argocd_account_token.this",
						"account",
						"test",
					),
					testCheckTokenIssuedAt(
						"argocd_account_token.this",
					),
				),
			},
		},
	})
}

func TestAccArgoCDAccountToken_Multiple(t *testing.T) {
	count := 3 + rand.Intn(7)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDAccountToken_Multiple(count),
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

func TestAccArgoCDAccountToken_RenewBefore(t *testing.T) {
	resourceName := "argocd_account_token.renew_before"

	expiresInSeconds := 30
	expiresIn := fmt.Sprintf("%ds", expiresInSeconds)
	expiresInDuration, _ := time.ParseDuration(expiresIn)

	renewBeforeSeconds := expiresInSeconds - 1

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDAccountTokenRenewBeforeSuccess(expiresIn, "20s"),
				Check: resource.ComposeTestCheckFunc(
					testCheckTokenExpiresAt(resourceName, int64(expiresInDuration.Seconds())),
					resource.TestCheckResourceAttr(resourceName, "renew_before", "20s"),
				),
			},
			{
				Config: testAccArgoCDAccountTokenRenewBeforeSuccess(expiresIn, fmt.Sprintf("%ds", renewBeforeSeconds)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "renew_before", fmt.Sprintf("%ds", renewBeforeSeconds)),
					testDelay(renewBeforeSeconds+1),
				),
				ExpectNonEmptyPlan: true, // token should be recreated when refreshed at end of step due to delay above
			},
			{
				Config:      testAccArgoCDAccountTokenRenewBeforeFailure(expiresInDuration),
				ExpectError: regexp.MustCompile("renew_before .* cannot be greater than expires_in .*"),
			},
		},
	})
}

func TestAccArgoCDAccountToken_RenewAfter(t *testing.T) {
	resourceName := "argocd_account_token.renew_after"
	renewAfterSeconds := 30

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDAccountTokenRenewAfter(renewAfterSeconds),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "renew_after", fmt.Sprintf("%ds", renewAfterSeconds)),
				),
			},
			{
				Config: testAccArgoCDAccountTokenRenewAfter(renewAfterSeconds),
				Check: resource.ComposeTestCheckFunc(
					testDelay(renewAfterSeconds + 1),
				),
				ExpectNonEmptyPlan: true, // token should be recreated when refreshed at end of step due to delay above
			},
			{
				Config: testAccArgoCDAccountTokenRenewAfter(renewAfterSeconds),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "renew_after", fmt.Sprintf("%ds", renewAfterSeconds)),
				),
			},
		},
	})
}

func testAccArgoCDAccountToken_DefaultAccount() string {
	return `
resource "argocd_account_token" "this" {}
`
}

func testAccArgoCDAccountToken_Expiry(expiresIn int64) string {
	return fmt.Sprintf(`
resource "argocd_account_token" "this" {
	expires_in = "%ds"
}
`, expiresIn)
}

func testAccArgoCDAccountToken_ExplicitAccount() string {
	return `
resource "argocd_account_token" "this" {
	account = "test"
}
`
}

func testAccArgoCDAccountToken_Multiple(count int) string {
	return fmt.Sprintf(`
resource "argocd_account_token" "multiple1a" {
	count = %d
}

resource "argocd_account_token" "multiple1b" {
	count = %d
}

resource "argocd_account_token" "multiple2a" {
	account = "test"
	count   = %d
}

resource "argocd_account_token" "multiple2b" {
	account = "test"
	count   = %d
}

`, count, count, count, count)
}

func testAccArgoCDAccountTokenRenewBeforeSuccess(expiresIn, renewBefore string) string {
	return fmt.Sprintf(`
resource "argocd_account_token" "renew_before" {
	expires_in   = "%s"
	renew_before = "%s"
}
`, expiresIn, renewBefore)
}

func testAccArgoCDAccountTokenRenewBeforeFailure(expiresInDuration time.Duration) string {
	expiresIn := int64(expiresInDuration.Seconds())
	renewBefore := int64(expiresInDuration.Seconds() + 1.0)

	return fmt.Sprintf(`
resource "argocd_account_token" "renew_before" {
	expires_in   = "%ds"
	renew_before = "%ds"
}
`, expiresIn, renewBefore)
}

func testAccArgoCDAccountTokenRenewAfter(renewAfter int) string {
	return fmt.Sprintf(`
resource "argocd_account_token" "renew_after" {
	account     = "test"
	renew_after = "%ds"
}
`, renewAfter)
}

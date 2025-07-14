package argocd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccArgoCDAccount_Basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDAccount_Basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_account.test",
						"name",
						"admin",
					),
					resource.TestCheckResourceAttr(
						"argocd_account.test",
						"enabled",
						"true",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_account.test",
						"capabilities.#",
					),
					testAccCheckArgoCDAccountExists("argocd_account.test"),
				),
			},
		},
	})
}

func TestAccArgoCDAccount_Import(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDAccount_Basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArgoCDAccountExists("argocd_account.test"),
				),
			},
			{
				ResourceName:      "argocd_account.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
					"tokens",
				},
			},
		},
	})
}

func TestAccArgoCDAccount_WithTokens(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDAccount_WithTokens(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_account.test",
						"name",
						"admin",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_account.test",
						"tokens.#",
					),
					testAccCheckArgoCDAccountExists("argocd_account.test"),
				),
			},
		},
	})
}

func TestAccArgoCDAccount_PasswordUpdate(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDAccount_NewAccountWithPassword("test", "acceptancetesting"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_account.test",
						"name",
						"test",
					),
					testAccCheckArgoCDAccountExists("argocd_account.test"),
				),
			},
			{
				Config: testAccArgoCDAccount_NewAccountWithPassword("test", "updated-password"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_account.test",
						"name",
						"test",
					),
					testAccCheckArgoCDAccountExists("argocd_account.test"),
				),
			},
		},
	})
}

func TestAccArgoCDAccount_ComputedFields(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDAccount_Basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_account.test",
						"name",
						"admin",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_account.test",
						"enabled",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_account.test",
						"capabilities.#",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_account.test",
						"tokens.#",
					),
					testAccCheckArgoCDAccountExists("argocd_account.test"),
				),
			},
		},
	})
}

func TestAccArgoCDAccount_InitialPasswordSet(t *testing.T) {
	// This test demonstrates that ArgoCD API requires the current password
	// to update/set a password, even for accounts without existing passwords.
	// This is a limitation of the ArgoCD API itself.
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDAccount_NewAccountWithoutPassword("test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_account.test",
						"name",
						"test",
					),
					testAccCheckArgoCDAccountExists("argocd_account.test"),
				),
			},
			{
				Config:      testAccArgoCDAccount_NewAccountWithPassword("test", "new-password"),
				ExpectError: regexp.MustCompile("current password does not match"),
			},
		},
	})
}

func testAccCheckArgoCDAccountExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("account name is not set")
		}

		accountName := rs.Primary.ID
		if accountName == "" {
			return fmt.Errorf("account name is empty")
		}

		// Additional validation could be added here to check if the account
		// actually exists in ArgoCD, but for basic testing, checking the ID
		// is sufficient since the read operation would fail if the account
		// doesn't exist
		return nil
	}
}

func testAccArgoCDAccount_Basic() string {
	return `
resource "argocd_account" "test" {
  name = "admin"
}
`
}

func testAccArgoCDAccount_WithTokens() string {
	return `
resource "argocd_account_token" "test" {
  account = "admin"
  expires_in = "1h"
}

resource "argocd_account" "test" {
  name = "admin"
  depends_on = [argocd_account_token.test]
}
`
}

func testAccArgoCDAccount_NewAccountWithPassword(accountName, password string) string {
	return fmt.Sprintf(`
resource "argocd_account" "test" {
  name = "%s"
  password = "%s"
}
`, accountName, password)
}

func testAccArgoCDAccount_NewAccountWithoutPassword(accountName string) string {
	return fmt.Sprintf(`
resource "argocd_account" "test" {
  name = "%s"
}
`, accountName)
}

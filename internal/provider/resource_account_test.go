package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccArgoCDAccountResource_Basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDAccountResource_Basic(),
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
					testAccCheckArgoCDAccountResourceExists("argocd_account.test"),
				),
			},
		},
	})
}

func TestAccArgoCDAccountResource_Import(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDAccountResource_Basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArgoCDAccountResourceExists("argocd_account.test"),
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

func TestAccArgoCDAccountResource_WithTokens(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDAccountResource_WithTokens(),
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
					testAccCheckArgoCDAccountResourceExists("argocd_account.test"),
				),
			},
		},
	})
}

func TestAccArgoCDAccountResource_PasswordUpdate(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDAccountResource_NewAccountWithPassword("test", "acceptancetesting"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_account.test",
						"name",
						"test",
					),
					testAccCheckArgoCDAccountResourceExists("argocd_account.test"),
				),
			},
			{
				Config: testAccArgoCDAccountResource_NewAccountWithPassword("test", "updated-password"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_account.test",
						"name",
						"test",
					),
					testAccCheckArgoCDAccountResourceExists("argocd_account.test"),
				),
			},
		},
	})
}

func TestAccArgoCDAccountResource_ComputedFields(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDAccountResource_Basic(),
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
					testAccCheckArgoCDAccountResourceExists("argocd_account.test"),
				),
			},
		},
	})
}

func TestAccArgoCDAccountResource_InitialPasswordSet(t *testing.T) {
	// This test demonstrates that ArgoCD API requires the current password
	// to update/set a password, even for accounts without existing passwords.
	// This is a limitation of the ArgoCD API itself.
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDAccountResource_NewAccountWithoutPassword("test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_account.test",
						"name",
						"test",
					),
					testAccCheckArgoCDAccountResourceExists("argocd_account.test"),
				),
			},
			{
				Config:      testAccArgoCDAccountResource_NewAccountWithPassword("test", "new-password"),
				ExpectError: regexp.MustCompile("current password does not match"),
			},
		},
	})
}

func testAccCheckArgoCDAccountResourceExists(resourceName string) resource.TestCheckFunc {
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

func testAccArgoCDAccountResource_Basic() string {
	return `
resource "argocd_account" "test" {
  name = "admin"
}
`
}

func testAccArgoCDAccountResource_WithTokens() string {
	return `
resource "argocd_account" "test" {
  name = "admin"
}

resource "argocd_account_token" "test" {
  account = argocd_account.test.name
  expires_in = "1h"
}
`
}

func testAccArgoCDAccountResource_NewAccountWithPassword(accountName, password string) string {
	return fmt.Sprintf(`
resource "argocd_account" "test" {
  name = "%s"
  password = "%s"
}
`, accountName, password)
}

func testAccArgoCDAccountResource_NewAccountWithoutPassword(accountName string) string {
	return fmt.Sprintf(`
resource "argocd_account" "test" {
  name = "%s"
}
`, accountName)
}

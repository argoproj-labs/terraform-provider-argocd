package argocd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccArgoCDAccountDataSource_Basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDAccountDataSource_Basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.argocd_account.test",
						"name",
						"admin",
					),
					resource.TestCheckResourceAttrSet(
						"data.argocd_account.test",
						"enabled",
					),
					resource.TestCheckResourceAttrSet(
						"data.argocd_account.test",
						"capabilities.#",
					),
				),
			},
		},
	})
}

func TestAccArgoCDAccountDataSource_WithTokens(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDAccountDataSource_WithTokens(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.argocd_account.test",
						"name",
						"admin",
					),
					resource.TestCheckResourceAttrSet(
						"data.argocd_account.test",
						"tokens.#",
					),
				),
			},
		},
	})
}

func testAccArgoCDAccountDataSource_Basic() string {
	return `
data "argocd_account" "test" {
  name = "admin"
}
`
}

func testAccArgoCDAccountDataSource_WithTokens() string {
	return `
resource "argocd_account_token" "test" {
  account = "admin"
  expires_in = "1h"
}

data "argocd_account" "test" {
  name = "admin"
  depends_on = [argocd_account_token.test]
}
`
}

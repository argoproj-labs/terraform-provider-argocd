package argocd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"testing"
)

func TestAccDataArgoCDProjectToken(t *testing.T) {
	project := "myproject"
	role := "test-role1234"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		// TODO: add expiry check
		// TODO: add token regeneration check
		// TODO: add token login check
		// TODO: add multiple token creation race condition check/rate limiting/retry
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDProjectToken(project, role),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_project_token.secret",
						"issued_at",
					),
				),
			},
		},
	})
}

func testAccArgoCDProjectToken(project string, role string) string {
	return fmt.Sprintf(`
resource "argocd_project_token" "secret" {
  project = "%s"
  role    = "%s"
}
	`, project, role)
}

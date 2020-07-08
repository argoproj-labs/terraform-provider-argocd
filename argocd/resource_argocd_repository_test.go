package argocd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"testing"
)

func TestAccArgoCDRepository(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDRepositorySimple(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository.simple",
						"connection_state_status",
						"Successful",
					),
				),
			},
			{
				Config: testAccArgoCDRepositoryHelm(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository.helm",
						"connection_state_status",
						"Successful",
					),
				),
			},
		},
	})
}

func testAccArgoCDRepositorySimple() string {
	return fmt.Sprintf(`
resource "argocd_repository" "simple" {
  repo = "https://github.com/kubernetes-sigs/kustomize"
}
`)
}

func testAccArgoCDRepositoryHelm() string {
	return fmt.Sprintf(`
resource "argocd_repository" "helm" {
  repo = "https://helm.nginx.com/stable"
  name = "nginx-stable"
  type = "helm"
}
`)
}

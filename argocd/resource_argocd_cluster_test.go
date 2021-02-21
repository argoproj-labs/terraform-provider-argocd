package argocd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccArgoCDCluster(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDClusterSimple(),
				Check: resource.TestCheckResourceAttr(
					"argocd_cluster.simple",
					"info.connection_state.status",
					"Successful",
				),
			},
		},
	})
}

func testAccArgoCDClusterSimple() string {
	return fmt.Sprintf(`
resource "argocd_cluster" "simple" {
  server = "https://kubernetes.default.svc.cluster.local"
  namespaces = ["default", "foo"]
  config {
    bearer_token = "0123456789abcdef"
    tls_client_config {
      insecure = true
    }
  }
}
`)
}

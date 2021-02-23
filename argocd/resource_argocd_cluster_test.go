package argocd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccArgoCDCluster(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDClusterSimple(acctest.RandString(10)),
				Check: resource.TestCheckResourceAttr(
					"argocd_cluster.simple",
					"info.0.connection_state.0.status",
					"Successful",
				),
			},
		},
	})
}

func testAccArgoCDClusterSimple(clusterName string) string {
	return fmt.Sprintf(`
resource "argocd_cluster" "simple" {
  server = "https://kubernetes.default.svc.cluster.local"
  name   = "%s"
  namespaces = ["default", "foo"]
  config {
    # Uses Kind's' bootstrap token
    bearer_token = "abcdef.0123456789abcdef"
    tls_client_config {
      insecure = true
    }
  }
}
`, clusterName)
}

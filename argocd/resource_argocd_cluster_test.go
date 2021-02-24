package argocd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccArgoCDCluster(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDClusterBearerToken(acctest.RandString(10)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_cluster.simple",
						"info.0.connection_state.0.status",
						"Successful",
					),
					resource.TestCheckResourceAttr(
						"argocd_cluster.simple",
						"shard",
						"1",
					),
					resource.TestCheckResourceAttr(
						"argocd_cluster.simple",
						"info.0.server_version",
						"1.19",
					),
					resource.TestCheckResourceAttr(
						"argocd_cluster.simple",
						"info.0.applications_count",
						"0",
					),
					resource.TestCheckResourceAttr(
						"argocd_cluster.simple",
						"config.0.tls_client_config.0.insecure",
						"true",
					),
				),
			},
		},
	})
}

func testAccArgoCDClusterBearerToken(clusterName string) string {
	return fmt.Sprintf(`
resource "argocd_cluster" "simple" {
  server = "https://kubernetes.default.svc.cluster.local"
  name   = "%s"
  shard  = "1"
  namespaces = ["default", "foo"]
  config {
    # Uses Kind's bootstrap token whose ttl is 24 hours after cluster bootstrap.
    bearer_token = "abcdef.0123456789abcdef"
    tls_client_config {
      insecure = true
    }
  }
}
`, clusterName)
}

func testAccArgoCDClusterTLSCertificate(clusterName string) string {
	//c := getInternalKubeClientSet()

	return fmt.Sprintf(`
resource "argocd_cluster" "simple" {
  server = "https://kubernetes.default.svc.cluster.local"
  name   = "%s"
  namespaces = ["bar", "baz"]
  config {
    tls_client_config {
      insecure = true
      
    }
  }
}
`, clusterName)
}

//
//func getInternalKubeClientSet() kubernetes.Clientset {
//	return clientSet
//}

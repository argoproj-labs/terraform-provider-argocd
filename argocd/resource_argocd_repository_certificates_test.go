package argocd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccArgoCDRepositoryCertificates(t *testing.T) {
	serverName := acctest.RandomWithPrefix("mywebsite")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDRepositoryCertificateSimple(
					serverName,
					"ssh",
					"ecdsa-sha2-nistp256",
					// gitlab's
					"AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFSMqzJeV9rUzU4kWitGjeR4PWSa29SPqJ1fVkhtj3Hw9xjLVXVYrU9QlYWrOLXBpQ6KWjbjTDTdDkoohFzgbEY=",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("argocd_certificate.simple", "server_name", serverName),
					resource.TestCheckResourceAttrSet("argocd_certificate.simple", "cert_type"),
					resource.TestCheckResourceAttrSet("argocd_certificate.simple", "cert_subtype"),
					resource.TestCheckResourceAttrSet("argocd_certificate.simple", "cert_info"),
				),
			},
			// same, no diff
			{
				Config: testAccArgoCDRepositoryCertificateSimple(
					serverName,
					"ssh",
					"ecdsa-sha2-nistp256",
					// gitlab's
					"AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFSMqzJeV9rUzU4kWitGjeR4PWSa29SPqJ1fVkhtj3Hw9xjLVXVYrU9QlYWrOLXBpQ6KWjbjTDTdDkoohFzgbEY=",
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// change only the cert_data => same id => diff
			{
				Config: testAccArgoCDRepositoryCertificateSimple(
					serverName,
					"ssh",
					"ecdsa-sha2-nistp256",
					// github's
					"AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEmKSENjQEezOmxkZMy7opKgwFB9nkt5YRrYMjNuG5N87uRgg6CLrbo5wAdT/y6v0mKV0U2w0WZ2YB/++Tpockg=",
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			// change cert_subtype & cert_data => changes id => diff
			{
				Config: testAccArgoCDRepositoryCertificateSimple(
					serverName,
					"ssh",
					"ssh-rsa",
					// github's
					"AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("argocd_certificate.simple", "server_name", serverName),
					resource.TestCheckResourceAttrSet("argocd_certificate.simple", "cert_type"),
					resource.TestCheckResourceAttrSet("argocd_certificate.simple", "cert_subtype"),
				),
			},
			{
				ResourceName:            "argocd_certificate.simple",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"cert_data"},
				Destroy:                 true,
			},
		},
	})
}

func TestAccArgoCDRepositoryCertificatesInvalid(t *testing.T) {
	certType := acctest.RandomWithPrefix("cert")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDRepositoryCertificateSimple(
					"serverName",
					certType,
					"",
					"",
				),
				ExpectError: regexp.MustCompile("mismatch: can only be https or ssh"),
			},
		},
	})
}

func testAccArgoCDRepositoryCertificateSimple(serverName, cert_type, cert_subtype, cert_data string) string {
	return fmt.Sprintf(`
resource "argocd_certificate" "simple" {
  server_name  = "%s"
  cert_type    = "%s"
  cert_subtype = "%s"
  cert_data    = <<EOT
%s
EOT
}
`, serverName, cert_type, cert_subtype, cert_data)
}

package argocd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"testing"
)

func TestAccArgoCDApplication(t *testing.T) {
	name := acctest.RandomWithPrefix("test-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSimple(name),
				Check: resource.TestCheckResourceAttrSet(
					"argocd_application.simple",
					"metadata.0.uid",
				),
			},
			// Check with the same name for rapid project recreation robustness
			{
				Config: testAccArgoCDApplicationSimple(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.simple",
						"metadata.0.uid",
					),
					// TODO: check all possible attributes
				),
			},
		},
	})
}

func testAccArgoCDApplicationSimple(name string) string {
	return fmt.Sprintf(`
resource "argocd_application" "simple" {
  metadata {
    name      = "%s"
    namespace = "argocd"
    labels = {
      acceptance = "true"
    }
    annotations = {
      "this.is.a.really.long.nested.key" = "yes, really!"
    }
  }

  spec {
    source {
      repo_url        = "https://kubernetes-charts.banzaicloud.com"
      chart           = "vault-operator"
      target_revision = "1.3.3"
      helm {
        parameters {
          name  = "image.tag"
          value = "1.3.3"
        }
      }
    }

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
  }
}
	`, name)
}

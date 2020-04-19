package argocd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"math/rand"
	"testing"
	"time"
)

func TestAccArgoCDProject(t *testing.T) {
	name := "test-acc-" + acctest.RandString(50)
	// ensure generated iat is always in the past
	iat := rand.Int63() % (time.Now().Unix() - 1)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDProjectSimple(name),
				Check: resource.TestCheckResourceAttrSet(
					"argocd_project.simple",
					"metadata.uid",
				),
			},
			{
				Config: testAccArgoCDProjectCoexistenceWithTokenResource(name, iat),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_project.coexistence",
						"metadata.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_project.coexistence",
						"roles.testrole.jwt_tokens.0.iat",
						string(iat),
					),
					resource.TestCheckResourceAttrPair(
						"argocd_project.coexistence",
						"roles.testrole.jwt_tokens.1.iat",
						"argocd_project_token.coexistence_testrole",
						"issued_at"),
				),
			},
		},
	})
}

func testAccArgoCDProjectSimple(name string) string {
	return fmt.Sprintf(`
resource "argocd_project" "simple" {
  metadata {
    name      = "%s"
    namespace = "argocd"
  }

  spec {
    description = "simple"
    source_repos = ["*"]
    destinations {
      server    = "https://kubernetes.default.svc"
      namespace = "*"
    }
    
  }
}
	`, name)
}

func testAccArgoCDProjectCoexistenceWithTokenResource(name string, iat int64) string {
	return fmt.Sprintf(`
resource "argocd_project" "coexistence" {
  metadata {
    name        = "%s"
    namespace   = "argocd"
  }

  spec {
    description = "coexistence"
    destinations {
	   server    = "https://kubernetes.default.svc"
	   namespace = "*"
    }
    source_repos = ["*"]
    roles = {
      name = "testrole"
      policies = [
        "p, proj:%s:testrole, applications, override, %s/*, allow",
      ]
      jwt_tokens {
        iat = %d
      }
    }
  }
}

resource "argocd_project_token" "coexistence_testrole" {
  project = argocd_project.coexistence.name
  role    = "testrole"
}

	`, name, name, name, iat)
}

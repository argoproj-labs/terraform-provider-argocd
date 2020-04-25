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
	name := acctest.RandomWithPrefix("test-acc")
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
					"metadata.0.uid",
				),
			},
			// Check with the same name for rapid project recreation robustness
			{
				Config: testAccArgoCDProjectSimple(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_project.simple",
						"metadata.0.uid",
					),
					// TODO: check all possible attributes
				),
			},
			{
				Config: testAccArgoCDProjectCoexistenceWithTokenResource(
					"test-acc-"+acctest.RandString(10),
					iat),
				Check: resource.ComposeAggregateTestCheckFunc(
					//resource.TestCheckResourceAttrSet(
					//	"argocd_project.coexistence",
					//	"metadata.0.uid",
					//),
					//resource.TestCheckResourceAttr(
					//	"argocd_project.coexistence",
					//	"spec.0.role.0.jwt_token.0.iat",
					//	convertInt64ToString(iat),
					//),
					resource.TestCheckResourceAttrPair(
						"argocd_project.coexistence",
						"spec.0.role.0.jwt_token.1.iat",
						"argocd_project_token.coexistence_testrole",
						"issued_at",
					),
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
    labels = {
      acceptance = "true"
    }
    annotations = {
      "this.is.a.really.long.nested.key" = "yes, really!"
    }
  }

  spec {
    description  = "simple"
    source_repos = ["*"]

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "foo"
    }
    cluster_resource_whitelist {
      group = "rbac.authorization.k8s.io"
      kind  = "ClusterRoleBinding"
    }
    cluster_resource_whitelist {
      group = "rbac.authorization.k8s.io"
      kind  = "ClusterRole"
    }
    namespace_resource_blacklist {
      group = "networking.k8s.io"
      kind  = "Ingress"
    }
    orphaned_resources = {
      warn = true
    }
    sync_window {
      kind = "allow"
      applications = ["api-*"]
      clusters = ["*"]
      namespaces = ["*"]
      duration = "3600s"
      schedule = "10 1 * * *"
      manual_sync = true
    }
    sync_window {
      kind = "deny"
      applications = ["foo"]
      clusters = ["in-cluster"]
      namespaces = ["default"]
      duration = "12h"
      schedule = "22 1 5 * *"
      manual_sync = false
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
    destination {
	   server    = "https://kubernetes.default.svc"
	   namespace = "*"
    }
    source_repos = ["*"]
    role {
      name = "testrole"
      policies = [
        "p, proj:%s:testrole, applications, override, %s/*, allow",
      ]
      jwt_token {
        iat = %d
      }
    }
  }
  allow_external_jwt_tokens = true
}

resource "argocd_project_token" "coexistence_testrole" {
  project = argocd_project.coexistence.metadata.0.name
  role    = "testrole"
}
	`, name, name, name, iat)
}

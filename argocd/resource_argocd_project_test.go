package argocd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"regexp"
	"testing"
)

func TestAccArgoCDProject(t *testing.T) {
	name := acctest.RandomWithPrefix("test-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDProjectPolicyError(
					"test-acc-" + acctest.RandString(10),
				),
				ExpectError: regexp.MustCompile("invalid policy rule"),
			},
			{
				Config: testAccArgoCDProjectRoleNameError(
					"test-acc-" + acctest.RandString(10),
				),
				ExpectError: regexp.MustCompile("invalid role name"),
			},
			{
				Config: testAccArgoCDProjectSyncWindowKindError(
					"test-acc-" + acctest.RandString(10),
				),
				ExpectError: regexp.MustCompile("mismatch: can only be allow or deny"),
			},
			{
				Config: testAccArgoCDProjectSyncWindowDurationError(
					"test-acc-" + acctest.RandString(10),
				),
				ExpectError: regexp.MustCompile("cannot parse duration"),
			},
			{
				Config: testAccArgoCDProjectSyncWindowScheduleError(
					"test-acc-" + acctest.RandString(10),
				),
				ExpectError: regexp.MustCompile("cannot parse schedule"),
			},
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
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_project.simple",
						"metadata.0.uid",
					),
					// TODO: check all possible attributes
				),
			},
		},
	})
}

func TestAccArgoCDProject_tokensCoexistence(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				ExpectNonEmptyPlan: true,
				Config: testAccArgoCDProjectCoexistenceWithTokenResource(
					"test-acc-"+acctest.RandString(10),
					4,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_project.coexistence",
						"metadata.0.uid",
					),
					resource.TestCheckNoResourceAttr(
						"argocd_project.coexistence",
						"spec.0.role.0.jwt_tokens",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_project_token.coexistence_testrole_exp",
						"issued_at",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_project_token.multiple.0",
						"issued_at",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_project_token.multiple.1",
						"issued_at",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_project_token.multiple.2",
						"issued_at",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_project_token.multiple.3",
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
    orphaned_resources {
      warn = true
      ignore {
        group = "apps/v1"
        kind  = "Deployment"
        name  = "ignored1"
      }
      ignore {
        group = "apps/v1"
        kind  = "Deployment"
        name  = "ignored2"
      }
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
    signature_keys = [
      "4AEE18F83AFDEB23",
      "07E34825A909B250"
    ]
  }
}
	`, name)
}

func testAccArgoCDProjectCoexistenceWithTokenResource(name string, count int) string {
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
        "p, proj:%s:testrole, applications, override, %s/foo, allow",
      ]
    }
  }
}

resource "argocd_project_token" "multiple" {
  count   = %d
  project = argocd_project.coexistence.metadata.0.name
  role    = "testrole"
}
resource "argocd_project_token" "coexistence_testrole_exp" {
  project    = argocd_project.coexistence.metadata.0.name
  role       = "testrole"
  expires_in = "264h"
}
	`, name, name, name, count)
}

func testAccArgoCDProjectPolicyError(name string) string {
	return fmt.Sprintf(`
resource "argocd_project" "failure" {
  metadata {
    name        = "%s"
    namespace   = "argocd"
  }

  spec {
    description = "expected policy failures"
    destination {
	   server    = "https://kubernetes.default.svc"
	   namespace = "*"
    }
    source_repos = ["*"]
    role {
      name = "incorrect-policy"
      policies = [
        "p, proj:%s:bar, applicat, foo, %s/*, whatever",
      ]
    }
  }
}
	`, name, name, name)
}

func testAccArgoCDProjectRoleNameError(name string) string {
	return fmt.Sprintf(`
resource "argocd_project" "failure" {
  metadata {
    name        = "%s"
    namespace   = "argocd"
  }

  spec {
    description = "expected role name failure"
    destination {
	   server    = "https://kubernetes.default.svc"
	   namespace = "*"
    }
    source_repos = ["*"]
    role {
      name = "incorrect role name"
      policies = [
        "p, proj:%s:testrole, applications, override, %s/foo, allow",
      ]
    }
  }
}
	`, name, name, name)
}

func testAccArgoCDProjectSyncWindowScheduleError(name string) string {
	return fmt.Sprintf(`
resource "argocd_project" "failure" {
  metadata {
    name        = "%s"
    namespace   = "argocd"
  }

  spec {
    description = "expected policy failures"
    destination {
	   server    = "https://kubernetes.default.svc"
	   namespace = "*"
    }
    source_repos = ["*"]
    role {
      name = "incorrect-syncwindow"
      policies = [
        "p, proj:%s:testrole, applications, override, %s/foo, allow",
      ]
    }
	sync_window {
      kind = "allow"
      applications = ["api-*"]
      clusters = ["*"]
      namespaces = ["*"]
      duration = "3600s"
      schedule = "10 1 * * * 5"
      manual_sync = true
    }
  }
}
	`, name, name, name)
}

func testAccArgoCDProjectSyncWindowDurationError(name string) string {
	return fmt.Sprintf(`
resource "argocd_project" "failure" {
  metadata {
    name        = "%s"
    namespace   = "argocd"
  }

  spec {
    description = "expected duration failure"
    destination {
	   server    = "https://kubernetes.default.svc"
	   namespace = "*"
    }
    source_repos = ["*"]
    role {
      name = "incorrect-syncwindow"
      policies = [
        "p, proj:%s:testrole, applications, override, %s/foo, allow",
      ]
    }
	sync_window {
      kind = "allow"
      applications = ["api-*"]
      clusters = ["*"]
      namespaces = ["*"]
      duration = "123"
      schedule = "10 1 * * *"
      manual_sync = true
    }
  }
}
	`, name, name, name)
}

func testAccArgoCDProjectSyncWindowKindError(name string) string {
	return fmt.Sprintf(`
resource "argocd_project" "failure" {
  metadata {
    name        = "%s"
    namespace   = "argocd"
  }

  spec {
    description = "expected kind failure"
    destination {
	   server    = "https://kubernetes.default.svc"
	   namespace = "*"
    }
    source_repos = ["*"]
    role {
      name = "incorrect-syncwindow"
      policies = [
        "p, proj:%s:testrole, applications, override, %s/foo, allow",
      ]
    }
	sync_window {
      kind = "whatever"
      applications = ["api-*"]
      clusters = ["*"]
      namespaces = ["*"]
      duration = "600s"
      schedule = "10 1 * * *"
      manual_sync = true
    }
  }
}
	`, name, name, name)
}

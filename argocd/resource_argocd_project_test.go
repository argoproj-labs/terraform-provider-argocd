package argocd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/oboukili/terraform-provider-argocd/internal/features"
)

func TestAccArgoCDProject(t *testing.T) {
	name := acctest.RandomWithPrefix("test-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
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
				Config: testAccArgoCDProjectSyncWindowTimezoneError(
					"test-acc-" + acctest.RandString(10),
				),
				ExpectError: regexp.MustCompile("cannot parse timezone"),
			},
			{
				Config: testAccArgoCDProjectSimple(name),
				Check: resource.TestCheckResourceAttrSet(
					"argocd_project.simple",
					"metadata.0.uid",
				),
			},
			{
				ResourceName:      "argocd_project.simple",
				ImportState:       true,
				ImportStateVerify: true,
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
			{
				Config: testAccArgoCDProjectSimpleWithoutOrphaned(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_project.simple",
						"metadata.0.uid",
						// TODO: check all possible attributes
					),
				),
			},
			{
				Config: testAccArgoCDProjectSimpleWithEmptyOrphaned(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_project.simple",
						"metadata.0.uid",
						// TODO: check all possible attributes
					),
				),
			},
		},
	})
}

func TestAccArgoCDProject_tokensCoexistence(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
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

func TestAccArgoCDProjectUpdateAddRole(t *testing.T) {
	name := acctest.RandomWithPrefix("test-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDProjectSimpleWithoutRole(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_project.simple",
						"metadata.0.uid",
					),
				),
			},
			{
				ResourceName:      "argocd_project.simple",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccArgoCDProjectSimpleWithRole(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_project.simple",
						"metadata.0.uid",
					),
				),
			},
		},
	})
}

func TestAccArgoCDProjectWithClustersRepositoriesRolePolicy(t *testing.T) {
	name := acctest.RandomWithPrefix("test-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDProjectWithClustersRepositoriesRolePolicy(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_project.simple",
						"metadata.0.uid",
					),
				),
			},
			{
				ResourceName:      "argocd_project.simple",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccArgoCDProjectWithLogsExecRolePolicy(t *testing.T) {
	name := acctest.RandomWithPrefix("test-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ExecLogsPolicy) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDProjectWithExecLogsRolePolicy(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_project.simple",
						"metadata.0.uid",
					),
				),
			},
			{
				ResourceName:      "argocd_project.simple",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccArgoCDProjectWithSourceNamespaces(t *testing.T) {
	name := acctest.RandomWithPrefix("test-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ProjectSourceNamespaces) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDProjectWithSourceNamespaces(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_project.simple",
						"metadata.0.uid",
					),
				),
			},
			{
				ResourceName:      "argocd_project.simple",
				ImportState:       true,
				ImportStateVerify: true,
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
    cluster_resource_whitelist {
      group = ""
      kind  = "Namespace"
    }
    cluster_resource_blacklist {
      group = ""
      kind  = "ResourceQuota"
    }
    cluster_resource_blacklist {
      group = "*"
      kind  = "*"
    }
    namespace_resource_blacklist {
      group = "networking.k8s.io"
      kind  = "Ingress"
    }
    namespace_resource_whitelist {
      group = "*"
      kind  = "*"
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
      timezone = "Europe/London"
    }
    signature_keys = [
      "4AEE18F83AFDEB23",
      "07E34825A909B250"
    ]
  }
}
	`, name)
}

func testAccArgoCDProjectSimpleWithoutOrphaned(name string) string {
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
      description  = "simple project"
      source_repos = ["*"]
  
      destination {
        name      = "anothercluster"
        namespace = "bar"
      }
    }
  }
	`, name)
}

func testAccArgoCDProjectSimpleWithEmptyOrphaned(name string) string {
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
      description  = "simple project"
      source_repos = ["*"]
  
      destination {
        name      = "anothercluster"
        namespace = "bar"
      }
      orphaned_resources { }
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

func testAccArgoCDProjectSimpleWithoutRole(name string) string {
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
      description  = "simple project"
      source_repos = ["*"]
  
      destination {
        name      = "anothercluster"
        namespace = "bar"
      }
      orphaned_resources {
        warn = true
        ignore {
          group = "apps/v1"
          kind  = "Deployment"
          name  = "ignored1"
        }
      }
    }
  }
	`, name)
}

func testAccArgoCDProjectSimpleWithRole(name string) string {
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
      description  = "simple project"
      source_repos = ["*"]
  
      destination {
        name      = "anothercluster"
        namespace = "bar"
      }
      orphaned_resources {
        warn = true
        ignore {
          group = "apps/v1"
          kind  = "Deployment"
          name  = "ignored1"
        }
      }
      role {
        name = "anotherrole"
        policies = [
          "p, proj:%s:anotherrole, applications, get, %s/*, allow",
          "p, proj:%s:anotherrole, applications, sync, %s/*, deny",
        ]
      }
    }
  }
	`, name, name, name, name, name)
}

func testAccArgoCDProjectWithClustersRepositoriesRolePolicy(name string) string {
	return fmt.Sprintf(`
  resource "argocd_project" "simple" {
    metadata {
      name      = "%[1]s"
      namespace = "argocd"
      labels = {
        acceptance = "true"
      }
      annotations = {
        "this.is.a.really.long.nested.key" = "yes, really!"
      }
    }
  
    spec {
      description  = "simple project"
      source_repos = ["*"]
  
      destination {
        name      = "anothercluster"
        namespace = "bar"
      }
      orphaned_resources {
        warn = true
        ignore {
          group = "apps/v1"
          kind  = "Deployment"
          name  = "ignored1"
        }
      }
      role {
        name = "admin"
        policies = [
          "p, proj:%[1]s:admin, clusters, get, %[1]s/*, allow",
          "p, proj:%[1]s:admin, repositories, get, %[1]s/*, allow",
        ]
      }
    }
  }
	`, name)
}

func testAccArgoCDProjectWithExecLogsRolePolicy(name string) string {
	return fmt.Sprintf(`
  resource "argocd_project" "simple" {
    metadata {
      name      = "%[1]s"
      namespace = "argocd"
      labels = {
        acceptance = "true"
      }
      annotations = {
        "this.is.a.really.long.nested.key" = "yes, really!"
      }
    }
  
    spec {
      description  = "simple project"
      source_repos = ["*"]
  
      destination {
        name      = "anothercluster"
        namespace = "bar"
      }
      orphaned_resources {
        warn = true
        ignore {
          group = "apps/v1"
          kind  = "Deployment"
          name  = "ignored1"
        }
      }
      role {
        name = "admin"
        policies = [
          "p, proj:%[1]s:admin, exec, create, %[1]s/*, allow",
          "p, proj:%[1]s:admin, logs, get, %[1]s/*, allow",
        ]
      }
    }
  }
	`, name)
}

func testAccArgoCDProjectWithSourceNamespaces(name string) string {
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
    description  = "simple project"
    source_repos = ["*"]
    source_namespaces = ["*"]

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "foo"
    }
    orphaned_resources {
      warn = true
      ignore {
        group = "apps/v1"
        kind  = "Deployment"
        name  = "ignored1"
      }
    }
  }
}
	`, name)
}

func testAccArgoCDProjectSyncWindowTimezoneError(name string) string {
	return fmt.Sprintf(`
resource "argocd_project" "failure" {
  metadata {
    name        = "%s"
    namespace   = "argocd"
  }

  spec {
    description = "expected timezone failure"
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
      duration = "1h"
      schedule = "10 1 * * *"
      manual_sync = true
      timezone = "invalid"
    }
  }
}
  `, name, name, name)
}

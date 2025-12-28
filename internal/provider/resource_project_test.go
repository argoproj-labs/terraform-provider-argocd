package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/argoproj-labs/terraform-provider-argocd/internal/features"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccArgoCDProject(t *testing.T) {
	name := acctest.RandomWithPrefix("test-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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
		PreCheck:                 func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ExecLogsPolicy) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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
		PreCheck:                 func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ProjectSourceNamespaces) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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

func TestAccArgoCDProjectWithDestinationServiceAccounts(t *testing.T) {
	name := acctest.RandomWithPrefix("test-acc")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckFeatureSupported(t, features.ProjectDestinationServiceAccounts)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDProjectWithDestinationServiceAccounts(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_project.simple",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_project.simple",
						"spec.0.destination_service_account.0.default_service_account",
						"default",
					),
					resource.TestCheckResourceAttr(
						"argocd_project.simple",
						"spec.0.destination_service_account.1.default_service_account",
						"foo",
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

func TestAccArgoCDProjectWithFineGrainedPolicy(t *testing.T) {
	name := acctest.RandomWithPrefix("test-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ProjectFineGrainedPolicy) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDProjectWithFineGrainedPolicy(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_project.fine_grained_policy",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_project.fine_grained_policy",
						"spec.0.role.0.policies.#",
						"2",
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

func testAccArgoCDProjectWithDestinationServiceAccounts(name string) string {
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
    destination_service_account {
      default_service_account = "default"
      namespace = "default"
      server = "https://kubernetes.default.svc"
    }
    destination_service_account {
      default_service_account = "foo"
      namespace = "foo"
      server = "https://kubernetes.default.svc"
    }
  }
}
  `, name)
}

// TestAccArgoCDProject_MetadataFieldsConsistency tests consistency of metadata fields
func TestAccArgoCDProject_MetadataFieldsConsistency(t *testing.T) {
	name := acctest.RandString(10)
	config := fmt.Sprintf(`
resource "argocd_project" "metadata_consistency" {
  metadata {
    name      = "%[1]s"
    namespace = "argocd"
    labels = {
      acceptance = "true"
      environment = "test"
    }
    annotations = {
      "this.is.a.really.long.nested.key" = "yes, really!"
      "description" = "test project"
    }
  }

  spec {
    description  = "test project for metadata consistency"
    source_repos = ["*"]

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
  }
}
`, name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_project.metadata_consistency",
						"metadata.0.name",
						name,
					),
					resource.TestCheckResourceAttr(
						"argocd_project.metadata_consistency",
						"metadata.0.labels.acceptance",
						"true",
					),
					resource.TestCheckResourceAttr(
						"argocd_project.metadata_consistency",
						"metadata.0.labels.environment",
						"test",
					),
					resource.TestCheckResourceAttr(
						"argocd_project.metadata_consistency",
						"metadata.0.annotations.this.is.a.really.long.nested.key",
						"yes, really!",
					),
				),
			},
			{
				// Apply the same configuration again to test for consistency
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_project.metadata_consistency",
						"metadata.0.name",
						name,
					),
					resource.TestCheckResourceAttr(
						"argocd_project.metadata_consistency",
						"metadata.0.labels.acceptance",
						"true",
					),
					resource.TestCheckResourceAttr(
						"argocd_project.metadata_consistency",
						"metadata.0.labels.environment",
						"test",
					),
					resource.TestCheckResourceAttr(
						"argocd_project.metadata_consistency",
						"metadata.0.annotations.this.is.a.really.long.nested.key",
						"yes, really!",
					),
				),
			},
		},
	})
}

// TestAccArgoCDProject_RolesConsistency tests consistency of role fields
func TestAccArgoCDProject_RolesConsistency(t *testing.T) {
	name := acctest.RandString(10)
	config := fmt.Sprintf(`
resource "argocd_project" "roles_consistency" {
  metadata {
    name      = "%[1]s"
    namespace = "argocd"
  }

  spec {
    description  = "test project with roles"
    source_repos = ["*"]

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }

    role {
      name = "admin"
      policies = [
        "p, proj:%[1]s:admin, applications, get, %[1]s/*, allow",
        "p, proj:%[1]s:admin, applications, sync, %[1]s/*, allow",
      ]
      groups = ["admin-group", "ops-group"]
    }
    role {
      name = "read-only"
      policies = [
        "p, proj:%[1]s:read-only, applications, get, %[1]s/*, allow",
      ]
      groups = ["dev-group"]
    }
  }
}
`, name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_project.roles_consistency",
						"spec.0.role.0.name",
						"admin",
					),
					resource.TestCheckResourceAttr(
						"argocd_project.roles_consistency",
						"spec.0.role.0.policies.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"argocd_project.roles_consistency",
						"spec.0.role.0.groups.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"argocd_project.roles_consistency",
						"spec.0.role.1.name",
						"read-only",
					),
				),
			},
			{
				// Apply the same configuration again to test for consistency
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_project.roles_consistency",
						"spec.0.role.0.name",
						"admin",
					),
					resource.TestCheckResourceAttr(
						"argocd_project.roles_consistency",
						"spec.0.role.0.policies.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"argocd_project.roles_consistency",
						"spec.0.role.0.groups.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"argocd_project.roles_consistency",
						"spec.0.role.1.name",
						"read-only",
					),
				),
			},
		},
	})
}

// TestAccArgoCDProject_SyncWindowsConsistency tests consistency of sync window fields
func TestAccArgoCDProject_SyncWindowsConsistency(t *testing.T) {
	name := acctest.RandString(10)
	config := fmt.Sprintf(`
resource "argocd_project" "sync_windows_consistency" {
  metadata {
    name      = "%[1]s"
    namespace = "argocd"
  }

  spec {
    description  = "test project with sync windows"
    source_repos = ["*"]

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
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
  }
}
`, name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_project.sync_windows_consistency",
						"spec.0.sync_window.0.kind",
						"allow",
					),
					resource.TestCheckResourceAttr(
						"argocd_project.sync_windows_consistency",
						"spec.0.sync_window.0.duration",
						"3600s",
					),
					resource.TestCheckResourceAttr(
						"argocd_project.sync_windows_consistency",
						"spec.0.sync_window.0.manual_sync",
						"true",
					),
					resource.TestCheckResourceAttr(
						"argocd_project.sync_windows_consistency",
						"spec.0.sync_window.1.timezone",
						"Europe/London",
					),
				),
			},
			{
				// Apply the same configuration again to test for consistency
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_project.sync_windows_consistency",
						"spec.0.sync_window.0.kind",
						"allow",
					),
					resource.TestCheckResourceAttr(
						"argocd_project.sync_windows_consistency",
						"spec.0.sync_window.0.duration",
						"3600s",
					),
					resource.TestCheckResourceAttr(
						"argocd_project.sync_windows_consistency",
						"spec.0.sync_window.0.manual_sync",
						"true",
					),
					resource.TestCheckResourceAttr(
						"argocd_project.sync_windows_consistency",
						"spec.0.sync_window.1.timezone",
						"Europe/London",
					),
				),
			},
		},
	})
}

// TestAccArgoCDProject_OrphanedResourcesConsistency tests consistency of orphaned resources
func TestAccArgoCDProject_OrphanedResourcesConsistency(t *testing.T) {
	name := acctest.RandString(10)
	config := fmt.Sprintf(`
resource "argocd_project" "orphaned_resources_consistency" {
  metadata {
    name      = "%[1]s"
    namespace = "argocd"
  }

  spec {
    description  = "test project with orphaned resources"
    source_repos = ["*"]

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
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
  }
}
`, name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_project.orphaned_resources_consistency",
						"spec.0.orphaned_resources.0.warn",
						"true",
					),
					resource.TestCheckResourceAttr(
						"argocd_project.orphaned_resources_consistency",
						"spec.0.orphaned_resources.0.ignore.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"argocd_project.orphaned_resources_consistency",
						"spec.0.orphaned_resources.0.ignore.0.name",
						"ignored1",
					),
				),
			},
			{
				// Apply the same configuration again to test for consistency
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_project.orphaned_resources_consistency",
						"spec.0.orphaned_resources.0.warn",
						"true",
					),
					resource.TestCheckResourceAttr(
						"argocd_project.orphaned_resources_consistency",
						"spec.0.orphaned_resources.0.ignore.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"argocd_project.orphaned_resources_consistency",
						"spec.0.orphaned_resources.0.ignore.0.name",
						"ignored1",
					),
				),
			},
		},
	})
}

func testAccArgoCDProjectWithFineGrainedPolicy(name string) string {
	return fmt.Sprintf(`
  resource "argocd_project" "fine_grained_policy" {
    metadata {
      name      = "%[1]s"
      namespace = "argocd"
      labels = {
        acceptance = "true"
      }
    }

    spec {
      description  = "simple project with fine-grained policies"
      source_repos = ["*"]

      destination {
        server    = "https://kubernetes.default.svc"
        namespace = "default"
      }

      role {
        name = "fine-grained"
        policies = [
          "p, proj:%[1]s:fine-grained, applications, update/*, %[1]s/*, allow",
          "p, proj:%[1]s:fine-grained, applications, delete/*/Pod/*/*, %[1]s/*, allow",
        ]
      }
    }
  }
	`, name)
}

// TestAccArgoCDProject_ProviderUpgradeStateMigration tests that resources created with the
// old SDK-based provider (v7.12.0) can be successfully read and managed by the new
// framework-based provider. This ensures backward compatibility when upgrading the provider.
func TestAccArgoCDProject_ProviderUpgradeStateMigration(t *testing.T) {
	name := acctest.RandomWithPrefix("test-acc-migrate")
	config := testAccArgoCDProjectForStateMigration(name)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				// Step 1: Create project using old SDK-based provider (v7.12.0)
				ExternalProviders: map[string]resource.ExternalProvider{
					"argocd": {
						VersionConstraint: "7.12.0",
						Source:            "argoproj-labs/argocd",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("argocd_project.migration", "metadata.0.name", name),
					resource.TestCheckResourceAttrSet("argocd_project.migration", "metadata.0.uid"),
				),
			},
			{
				// Step 2: Upgrade to new framework-based provider - verify it can read existing state
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("argocd_project.migration", "metadata.0.name", name),
					resource.TestCheckResourceAttr("argocd_project.migration", "spec.0.description", "project for state migration testing"),
					resource.TestCheckResourceAttr("argocd_project.migration", "spec.0.source_repos.#", "2"),
					resource.TestCheckResourceAttr("argocd_project.migration", "spec.0.destination.#", "2"),
					resource.TestCheckResourceAttr("argocd_project.migration", "spec.0.role.#", "2"),
				),
			},
			{
				// Step 3: Verify no unexpected plan changes after migration
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccArgoCDProjectForStateMigration(name string) string {
	return fmt.Sprintf(`
resource "argocd_project" "migration" {
  metadata {
    name      = "%[1]s"
    namespace = "argocd"
    labels = {
      test = "migration"
      env  = "acceptance"
    }
    annotations = {
      "description" = "testing provider upgrade"
    }
  }

  spec {
    description  = "project for state migration testing"
    source_repos = ["https://github.com/example/repo1", "https://github.com/example/repo2"]

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "production"
    }

    cluster_resource_whitelist {
      group = "rbac.authorization.k8s.io"
      kind  = "ClusterRole"
    }

    namespace_resource_blacklist {
      group = "v1"
      kind  = "ConfigMap"
    }

    orphaned_resources {
      warn = true
      ignore {
        group = "apps/v1"
        kind  = "Deployment"
        name  = "legacy-app"
      }
    }

    role {
      name = "admin"
      description = "Admin role"
      policies = [
        "p, proj:%[1]s:admin, applications, *, %[1]s/*, allow",
      ]
      groups = ["platform-team"]
    }

    role {
      name = "readonly"
      description = "Read-only role"
      policies = [
        "p, proj:%[1]s:readonly, applications, get, %[1]s/*, allow",
      ]
      groups = ["developers"]
    }

    sync_window {
      kind         = "allow"
      applications = ["*"]
      clusters     = ["*"]
      namespaces   = ["*"]
      duration     = "1h"
      schedule     = "0 22 * * *"
      manual_sync  = true
    }
  }
}
	`, name)
}

// TestAccArgoCDProject_ProviderUpgradeStateMigration_WithoutNamespace tests the specific
// case reported in issue #783 where projects created without an explicit namespace field
// in v7.12.1 cause forced replacement when upgrading to v7.12.3+.
// The namespace should be computed from the API response without causing drift.
func TestAccArgoCDProject_ProviderUpgradeStateMigration_WithoutNamespace(t *testing.T) {
	name := acctest.RandomWithPrefix("test-acc-migrate-no-ns")
	config := testAccArgoCDProjectForStateMigrationWithoutNamespace(name)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				// Step 1: Create project using old SDK-based provider (v7.12.1)
				// without specifying namespace in metadata (this is the key scenario)
				ExternalProviders: map[string]resource.ExternalProvider{
					"argocd": {
						VersionConstraint: "7.12.1",
						Source:            "argoproj-labs/argocd",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("argocd_project.tech", "metadata.0.name", name),
					resource.TestCheckResourceAttrSet("argocd_project.tech", "metadata.0.uid"),
				),
			},
			{
				// Step 2: Upgrade to new framework-based provider
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("argocd_project.tech", "metadata.0.name", name),
					resource.TestCheckResourceAttr("argocd_project.tech", "spec.0.source_repos.#", "1"),
					resource.TestCheckResourceAttr("argocd_project.tech", "spec.0.destination.#", "1"),
					resource.TestCheckResourceAttr("argocd_project.tech", "spec.0.cluster_resource_whitelist.#", "1"),
					// Namespace should be computed from API, not forcing replacement
					resource.TestCheckResourceAttr("argocd_project.tech", "metadata.0.namespace", "argocd"),
				),
			},
			{
				// Step 3: Verify no unexpected plan changes after migration (issue #783)
				// This should NOT show a forced replacement due to namespace changing
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccArgoCDProjectForStateMigrationWithoutNamespace(name string) string {
	return fmt.Sprintf(`
resource "argocd_project" "tech" {
  metadata {
    name = "%s"
    # NOTE: namespace is intentionally NOT specified here to test issue #783
  }

  spec {
    source_repos = ["*"]

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "*"
    }

    cluster_resource_whitelist {
      group = "*"
      kind  = "*"
    }
  }
}
	`, name)
}

// TestAccArgoCDProject_EmptySourceRepos tests the issue #788 where an empty source_repos list
// causes "Provider produced inconsistent result after apply" error.
// The provider should maintain an empty list as empty list, not convert it to null.
func TestAccArgoCDProject_EmptySourceRepos(t *testing.T) {
	name := acctest.RandomWithPrefix("test-acc-empty-repos")
	config := testAccArgoCDProjectWithEmptySourceRepos(name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("argocd_project.empty_repos", "metadata.0.name", name),
					resource.TestCheckResourceAttr("argocd_project.empty_repos", "spec.0.source_repos.#", "0"),
				),
			},
			{
				// Apply the same configuration again to verify no drift
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccArgoCDProjectWithEmptySourceRepos(name string) string {
	return fmt.Sprintf(`
resource "argocd_project" "empty_repos" {
  metadata {
    name      = "%s"
    namespace = "argocd"
  }

  spec {
    description  = "project with empty source_repos"
    source_repos = []

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
  }
}
	`, name)
}

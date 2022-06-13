package argocd

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccArgoCDApplication(t *testing.T) {
	commonName := acctest.RandomWithPrefix("test-acc")
	revisionHistoryLimit := acctest.RandIntRange(0, 9)
	helmValues := `
ingress:
  enabled: true
  path: /
  hosts:
    - mydomain.example.com
  annotations:
    kubernetes.io/ingress.class: nginx
    kubernetes.io/tls-acme: "true"
  labels: {}
  tls:
    - secretName: mydomain-tls
      hosts:
        - mydomain.example.com
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSimple(commonName),
				Check: resource.TestCheckResourceAttrSet(
					"argocd_application.simple",
					"metadata.0.uid",
				),
			},
			// Check with the same name for rapid application recreation robustness
			{
				Config: testAccArgoCDApplicationSimple(commonName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.simple",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.simple",
						"spec.0.source.0.target_revision",
						"16.9.11",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.simple",
						"spec.0.revision_history_limit",
						"10",
					),
				),
			},
			{
				Config:             testAccArgoCDApplicationSimpleWait(commonName),
				ExpectNonEmptyPlan: true,
				Check: resource.TestCheckResourceAttr(
					"argocd_application.simple",
					"wait",
					"true",
				),
			},
			{
				Config: testAccArgoCDApplicationHelm(
					acctest.RandomWithPrefix("test-acc"),
					helmValues),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.helm",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.helm",
						"spec.0.source.0.helm.0.values",
						helmValues+"\n",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.helm",
						"spec.0.source.0.helm.0.value_files.0",
						"values.yaml",
					),
				),
			},
			{
				Config: testAccArgoCDApplicationKustomize(
					acctest.RandomWithPrefix("test-acc")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.kustomize",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.kustomize",
						"spec.0.source.0.target_revision",
						"release-kustomize-v3.7",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.kustomize",
						"spec.0.source.0.kustomize.0.name_suffix",
						"-bar",
					),
				),
			},
			{
				Config: testAccArgoCDApplicationDirectory(
					acctest.RandomWithPrefix("test-acc")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.directory",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.directory.0.recurse",
						"false",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.directory.0.jsonnet.0.ext_var.0.name",
						"somename",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.directory.0.jsonnet.0.ext_var.0.value",
						"somevalue",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.directory.0.jsonnet.0.ext_var.0.code",
						"false",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.directory.0.jsonnet.0.ext_var.1.name",
						"anothername",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.directory.0.jsonnet.0.ext_var.1.value",
						"anothervalue",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.directory.0.jsonnet.0.ext_var.1.code",
						"true",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.directory.0.jsonnet.0.tla.0.name",
						"yetanothername",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.directory.0.jsonnet.0.tla.0.value",
						"yetanothervalue",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.directory.0.jsonnet.0.tla.0.code",
						"true",
					),
				),
			},
			{
				Config: testAccArgoCDApplicationSyncPolicy(
					acctest.RandomWithPrefix("test-acc")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.sync_policy",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.sync_policy",
						"spec.0.sync_policy.0.automated.prune",
						"true",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.sync_policy",
						"spec.0.sync_policy.0.automated.self_heal",
						"true",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.sync_policy",
						"spec.0.sync_policy.0.automated.allow_empty",
						"true",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.sync_policy",
						"spec.0.sync_policy.0.retry.0.backoff.duration",
						"30s",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.sync_policy",
						"spec.0.sync_policy.0.retry.0.backoff.max_duration",
						"2m",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.sync_policy",
						"spec.0.sync_policy.0.retry.0.backoff.factor",
						"2",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.sync_policy",
						"spec.0.sync_policy.0.retry.0.limit",
						"5",
					),
				),
			},
			{
				Config: testAccArgoCDApplicationIgnoreDifferences(
					acctest.RandomWithPrefix("test-acc")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.ignore_differences",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.ignore_differences",
						"spec.0.ignore_difference.0.kind",
						"Deployment",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.ignore_differences",
						"spec.0.ignore_difference.1.group",
						"apps",
					),
				),
			},
			{
				SkipFunc: testAccSkipFeatureIgnoreDiffJQPathExpressions,
				Config: testAccArgoCDApplicationIgnoreDiffJQPathExpressions(
					acctest.RandomWithPrefix("test-acc")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.ignore_differences_jqpe",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.ignore_differences_jqpe",
						"spec.0.ignore_difference.0.jq_path_expressions.0",
						".spec.replicas",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.ignore_differences_jqpe",
						"spec.0.ignore_difference.1.jq_path_expressions.1",
						".spec.template.spec.metadata.labels.somelabel",
					),
				),
			},
			{
				Config: testAccArgoCDApplicationSimpleRevisionHistory(commonName, revisionHistoryLimit),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.simple",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.simple",
						"spec.0.revision_history_limit",
						fmt.Sprint(revisionHistoryLimit),
					),
				),
			},
		},
	})
}

func TestAccArgoCDApplication_NoSyncPolicyBlock(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationNoSyncPolicy(acctest.RandomWithPrefix("test-acc")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.simple",
						"metadata.0.uid",
					),
					resource.TestCheckNoResourceAttr(
						"argocd_application.simple",
						"spec.0.sync_policy.0.retry.0.backoff.duration",
					),
					resource.TestCheckNoResourceAttr(
						"argocd_application.simple",
						"spec.0.sync_policy.0.automated.prune",
					),
				),
			},
		}})
}

func TestAccArgoCDApplication_Recurse(t *testing.T) {
	name := acctest.RandomWithPrefix("test-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationRecurseDirectory(name, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.directory.0.recurse",
						"true",
					),
					resource.TestCheckNoResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.directory.0.jsonnet",
					),
				),
			},
			{
				Config: testAccArgoCDApplicationDirectoryImplicitNonRecurse(acctest.RandomWithPrefix("test-acc")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.directory.0.recurse",
						"false",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.directory.0.jsonnet.0.ext_var.0.name",
						"somename",
					),
				),
			},
			{
				Config: testAccArgoCDApplicationRecurseDirectory(name, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.directory.0.recurse",
					),
					resource.TestCheckNoResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.directory.0.jsonnet",
					),
				),
			},
			{
				Config: testAccArgoCDApplicationRecurseDirectory(name, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.directory.0.recurse",
						"true",
					),
					resource.TestCheckNoResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.directory.0.jsonnet",
					),
				),
			},
			{
				Config: testAccArgoCDApplicationRecurseDirectory(name, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.directory.0.recurse",
					),
					resource.TestCheckNoResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.directory.0.jsonnet",
					),
				),
			},
		}})
}

func TestAccArgoCDApplication_EmptySyncPolicyBlock(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationEmptySyncPolicy(acctest.RandomWithPrefix("test-acc")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.simple",
						"metadata.0.uid",
					),
					resource.TestCheckNoResourceAttr(
						"argocd_application.simple",
						"spec.0.sync_policy.0.retry.0.backoff.duration",
					),
					resource.TestCheckNoResourceAttr(
						"argocd_application.simple",
						"spec.0.sync_policy.0.automated.prune",
					),
				),
			},
		}})
}

func TestAccArgoCDApplication_NoAutomatedBlock(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationNoAutomated(acctest.RandomWithPrefix("test-acc")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.simple",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_application.simple",
						"spec.0.sync_policy.0.retry.0.backoff.duration",
					),
					resource.TestCheckNoResourceAttr(
						"argocd_application.simple",
						"spec.0.sync_policy.0.automated.prune",
					),
				),
			},
		}})
}

func TestAccArgoCDApplication_EmptyAutomatedBlock(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationEmptyAutomated(acctest.RandomWithPrefix("test-acc")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.simple",
						"metadata.0.uid",
					),
					resource.TestCheckNoResourceAttr(
						"argocd_application.simple",
						"spec.0.sync_policy.0.automated.prune",
					),
				),
			},
		}})
}

func TestAccArgoCDApplication_OptionalPath(t *testing.T) {
	app := acctest.RandomWithPrefix("test-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationDirectoryNoPath(app),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.directory",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.path",
						".",
					),
				),
			},
			{
				Config: testAccArgoCDApplicationDirectoryPath(app, "."),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.directory",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.path",
						".",
					),
				),
			},
			{
				Config: testAccArgoCDApplicationDirectoryNoPath(app),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.directory",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.path",
						".",
					),
				),
			},
		}})
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
      repo_url        = "https://charts.bitnami.com/bitnami"
      chart           = "redis"
      target_revision = "16.9.11"
      helm {
        parameter {
          name  = "image.tag"
          value = "6.2.5"
        }
        parameter {
          name  = "architecture"
          value = "standalone"
        }
        release_name = "testing"
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

func testAccArgoCDApplicationSimpleWait(name string) string {
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
      repo_url        = "https://charts.bitnami.com/bitnami"
      chart           = "redis"
      target_revision = "16.9.11"
      helm {
        parameter {
          name  = "image.tag"
          value = "6.2.5"
        }
        parameter {
          name  = "architecture"
          value = "standalone"
        }
        release_name = "testing"
      }
    }
    sync_policy {
      automated = {
        prune     = true
        self_heal = true
      }
    }
    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
  }
  wait = true
}
	`, name)
}

func testAccArgoCDApplicationHelm(name, helmValues string) string {
	return fmt.Sprintf(`
resource "argocd_application" "helm" {
  metadata {
    name      = "%s"
    namespace = "argocd"
    labels = {
      acceptance = "true"
    }
  }

  spec {
    source {
      repo_url        = "https://charts.bitnami.com/bitnami"
      chart           = "redis"
      target_revision = "16.9.11"
      helm {
        release_name = "testing"
        
        parameter {
          name  = "image.tag"
          value = "6.2.5"
        }
        parameter {
          name  = "architecture"
          value = "standalone"
        }

        value_files = ["values.yaml"]

        values = <<EOT
%s
EOT
      }
    }

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
  }
}
	`, name, helmValues)
}

func testAccArgoCDApplicationKustomize(name string) string {
	return fmt.Sprintf(`
resource "argocd_application" "kustomize" {
  metadata {
    name      = "%s"
    namespace = "argocd"
    labels = {
      acceptance = "true"
    }
  }

  spec {
    source {
      repo_url        = "https://github.com/kubernetes-sigs/kustomize"
      path            = "examples/helloWorld"
      target_revision = "release-kustomize-v3.7"
      kustomize {
  	    name_prefix  = "foo-"
	  	name_suffix = "-bar"
	  	images = [
          "hashicorp/terraform:light",
	    ]
	  	common_labels = {
		  "this.is.a.common" = "la-bel"
		  "another.io/one"   = "true" 
	    }
        common_annotations = {
		  "this.is.a.common" = "anno-tation"
		  "another.io/one"   = "false"
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

func testAccArgoCDApplicationDirectoryNoPath(name string) string {
	return fmt.Sprintf(`
resource "argocd_application" "directory" {
  metadata {
    name      = "%s"
    namespace = "argocd"
    labels = {
      acceptance = "true"
    }
  }

  spec {
    source {
      repo_url        = "https://github.com/MrLuje/argocd-example"
      target_revision = "yaml-at-root"
    }

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
  }
}
	`, name)
}

func testAccArgoCDApplicationDirectoryPath(name string, path string) string {
	return fmt.Sprintf(`
resource "argocd_application" "directory" {
  metadata {
    name      = "%s"
    namespace = "argocd"
    labels = {
      acceptance = "true"
    }
  }

  spec {
    source {
      repo_url        = "https://github.com/MrLuje/argocd-example"
      path            = "%s"
      target_revision = "yaml-at-root"
    }

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
  }
}
	`, name, path)
}

func testAccArgoCDApplicationDirectory(name string) string {
	return fmt.Sprintf(`
resource "argocd_application" "directory" {
  metadata {
    name      = "%s"
    namespace = "argocd"
    labels = {
      acceptance = "true"
    }
  }

  spec {
    source {
      repo_url        = "https://github.com/solo-io/gloo"
      path            = "install/helm/gloo"
      target_revision = "v1.4.2"
      directory {
        recurse = false
        jsonnet {
          ext_var {
            name  = "somename"
            value = "somevalue"
            code  = false
          }
          ext_var {
            name  = "anothername"
            value = "anothervalue"
            code  = true
          }
          tla {
            name  = "yetanothername"
            value = "yetanothervalue"
            code  = true
          }
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

func testAccArgoCDApplicationDirectoryImplicitNonRecurse(name string) string {
	return fmt.Sprintf(`
resource "argocd_application" "directory" {
  metadata {
    name      = "%s"
    namespace = "argocd"
    labels = {
      acceptance = "true"
    }
  }

  spec {
    source {
      repo_url        = "https://github.com/solo-io/gloo"
      path            = "install/helm/gloo"
      target_revision = "v1.4.2"
      directory {
        jsonnet {
          ext_var {
            name  = "somename"
            value = "somevalue"
            code  = false
          }
          ext_var {
            name  = "anothername"
            value = "anothervalue"
            code  = true
          }
          tla {
            name  = "yetanothername"
            value = "yetanothervalue"
            code  = true
          }
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

func testAccArgoCDApplicationRecurseDirectory(name string, recurse bool) string {
	return fmt.Sprintf(`
resource "argocd_application" "directory" {
  metadata {
    name      = "%s"
    namespace = "argocd"
    labels = {
      acceptance = "true"
    }
  }

  spec {
    source {
      repo_url        = "https://github.com/argoproj/argocd-example-apps"
      path            = "guestbook"
      target_revision = "HEAD"
      directory {
        recurse = %s
      }
    }

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
  }
}
	`, name, strconv.FormatBool(recurse))
}

func testAccArgoCDApplicationSyncPolicy(name string) string {
	return fmt.Sprintf(`
resource "argocd_application" "sync_policy" {
  metadata {
    name      = "%s"
    namespace = "argocd"
    labels = {
      acceptance = "true"
    }
  }

  spec {
    source {
      repo_url        = "https://charts.bitnami.com/bitnami"
      chart           = "redis"
      target_revision = "16.9.11"
    }

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
    
    sync_policy {
      automated = {
        prune       = true
        self_heal   = true
        allow_empty = true
      }
      retry {
        limit   = "5"
        backoff = {
          duration     = "30s"
          max_duration = "2m"
          factor       = "2"
        }
      }
    }
  }
}
	`, name)
}

func testAccArgoCDApplicationIgnoreDifferences(name string) string {
	return fmt.Sprintf(`
resource "argocd_application" "ignore_differences" {
  metadata {
    name      = "%s"
    namespace = "argocd"
    labels = {
      acceptance = "true"
    }
  }

  spec {
    source {
      repo_url        = "https://charts.bitnami.com/bitnami"
      chart           = "redis"
      target_revision = "16.9.11"
    }

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
    
    ignore_difference {
      group               = "apps"
      kind                = "Deployment"
      json_pointers       = ["/spec/replicas"]
    }

    ignore_difference {
      group         = "apps"
      kind          = "StatefulSet"
      name          = "someStatefulSet"
      json_pointers = [
        "/spec/replicas",
        "/spec/template/spec/metadata/labels/somelabel",
      ]
    }
  }
}
	`, name)
}

func testAccArgoCDApplicationIgnoreDiffJQPathExpressions(name string) string {
	return fmt.Sprintf(`
resource "argocd_application" "ignore_differences_jqpe" {
  metadata {
    name      = "%s"
    namespace = "argocd"
    labels = {
      acceptance = "true"
    }
  }

  spec {
    source {
      repo_url        = "https://charts.bitnami.com/bitnami"
      chart           = "redis"
      target_revision = "16.9.11"
    }

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
    
    ignore_difference {
      group               = "apps"
      kind                = "Deployment"
      jq_path_expressions = [".spec.replicas"]
    }

    ignore_difference {
      group         = "apps"
      kind          = "StatefulSet"
      name          = "someStatefulSet"
      jq_path_expressions = [
        ".spec.replicas",
        ".spec.template.spec.metadata.labels.somelabel",
      ]
    }
  }
}
	`, name)
}

func testAccArgoCDApplicationNoSyncPolicy(name string) string {
	return fmt.Sprintf(`
resource "argocd_application" "simple" {
  metadata {
    name      = "%s"
    namespace = "argocd"
  }
  spec {
    source {
      repo_url        = "https://charts.bitnami.com/bitnami"
      chart           = "redis"
      target_revision = "16.9.11"
      helm {
        release_name = "testing"
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

func testAccArgoCDApplicationEmptySyncPolicy(name string) string {
	return fmt.Sprintf(`
resource "argocd_application" "simple" {
  metadata {
    name      = "%s"
    namespace = "argocd"
  }
  spec {
    source {
      repo_url        = "https://charts.bitnami.com/bitnami"
      chart           = "redis"
      target_revision = "16.9.11"
      helm {
        release_name = "testing"
      }
    }
    sync_policy {
    }
    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
  }
}
	`, name)
}

func testAccArgoCDApplicationNoAutomated(name string) string {
	return fmt.Sprintf(`
resource "argocd_application" "simple" {
  metadata {
    name      = "%s"
    namespace = "argocd"
  }
  spec {
    source {
      repo_url        = "https://charts.bitnami.com/bitnami"
      chart           = "redis"
      target_revision = "16.9.11"
      helm {
        release_name = "testing"
      }
    }
    sync_policy {
      retry {
        limit   = "5"
        backoff = {
          duration     = "30s"
          max_duration = "2m"
          factor       = "2"
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

func testAccArgoCDApplicationEmptyAutomated(name string) string {
	return fmt.Sprintf(`
resource "argocd_application" "simple" {
  metadata {
    name      = "%s"
    namespace = "argocd"
  }
  spec {
    source {
      repo_url        = "https://charts.bitnami.com/bitnami"
      chart           = "redis"
      target_revision = "16.9.11"
      helm {
        release_name = "testing"
      }
    }
    sync_policy {
      automated = {}
    }
    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
  }
}
	`, name)
}

func testAccArgoCDApplicationSimpleRevisionHistory(name string, revision_history_limit int) string {
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
    revision_history_limit = %d
    source {
      repo_url        = "https://charts.bitnami.com/bitnami"
      chart           = "redis"
      target_revision = "16.9.11"
      helm {
        parameter {
          name  = "image.tag"
          value = "6.2.5"
        }
        parameter {
          name  = "architecture"
          value = "standalone"
        }
        release_name = "testing"
      }
    }

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
  }
}
	`, name, revision_history_limit)
}

func testAccSkipFeatureIgnoreDiffJQPathExpressions() (bool, error) {
	p, _ := testAccProviders["argocd"]()
	_ = p.Configure(context.Background(), &terraform.ResourceConfig{})
	server := p.Meta().(*ServerInterface)
	err := server.initClients()
	if err != nil {
		return false, err
	}
	featureSupported, err := server.isFeatureSupported(featureIgnoreDiffJQPathExpressions)
	if err != nil {
		return false, err
	}
	if !featureSupported {
		return true, nil
	}
	return false, nil
}

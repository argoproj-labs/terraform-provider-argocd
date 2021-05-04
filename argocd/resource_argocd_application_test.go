package argocd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccArgoCDApplication(t *testing.T) {
	commonName := acctest.RandomWithPrefix("test-acc")
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
						"1.3.3",
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
				// TODO: ArgoCD API ApplicationQuery does not return Directory attributes, investigate?
				// TODO: this provokes perpetual TF state drift as spec.0.source.0.directory cannot be read
				// TODO: the Directory attributes are to be used with care until a fix is made upstream
				Config: testAccArgoCDApplicationDirectory(
					acctest.RandomWithPrefix("test-acc")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.directory",
						"metadata.0.uid",
					),
					//resource.TestCheckResourceAttr(
					//	"argocd_application.directory",
					//	"spec.0.source.0.directory.0.recurse",
					//	"false",
					//),
					//resource.TestCheckResourceAttr(
					//	"argocd_application.directory",
					//	"spec.0.source.0.directory.0.jsonnet",
					//	"false",
					//),
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
        parameter {
          name  = "image.tag"
          value = "1.3.3"
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
      repo_url        = "https://kubernetes-charts.banzaicloud.com"
      chart           = "vault-operator"
      target_revision = "1.3.3"
      helm {
        parameter {
          name  = "image.tag"
          value = "1.3.3"
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
      repo_url        = "https://kubernetes-charts.banzaicloud.com"
      chart           = "vault-operator"
      target_revision = "1.3.3"
      helm {
        release_name = "testing"
        
        parameter {
          name  = "image.tag"
          value = "1.3.3"
        }
        parameter {
          name  = "banks-vaults.version"
          value = "1.3.3"
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
      repo_url        = "https://kubernetes-charts.banzaicloud.com"
      chart           = "vault-operator"
      target_revision = "1.3.3"
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
      repo_url        = "https://kubernetes-charts.banzaicloud.com"
      chart           = "vault-operator"
      target_revision = "1.3.3"
    }

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
    
    ignore_difference {
      group         = "apps"
      kind          = "Deployment"
      json_pointers = ["/spec/replicas"]
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

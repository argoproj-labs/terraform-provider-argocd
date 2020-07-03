package argocd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"testing"
)

func TestAccArgoCDApplication(t *testing.T) {
	commonName := acctest.RandomWithPrefix("test-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSimple(commonName),
				Check: resource.TestCheckResourceAttrSet(
					"argocd_application.simple",
					"metadata.0.uid",
				),
			},
			// Check with the same name for rapid project recreation robustness
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
						"master",
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
				// TODO: this provokes perpetual TF state drift
				// TODO: the Directory attributes are hence disabled until a fix is made upstream
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
						"spec.0.source.0.directory.0.jsonnet",
						"false",
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
      target_revision = "master"
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

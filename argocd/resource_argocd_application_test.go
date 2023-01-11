package argocd

import (
	"context"
	"fmt"
	"regexp"
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
			{
				ResourceName:            "argocd_application.simple",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait", "cascade"},
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
				Config: testAccArgoCDApplicationSimpleWait(commonName),
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
				ResourceName:            "argocd_application.helm",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait", "cascade", "metadata.0.generation", "metadata.0.resource_version"},
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
				ResourceName:            "argocd_application.kustomize",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait", "cascade", "metadata.0.generation", "metadata.0.resource_version"},
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
				ResourceName:            "argocd_application.directory",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait", "cascade", "metadata.0.generation", "metadata.0.resource_version"},
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
						"spec.0.sync_policy.0.automated.0.prune",
						"true",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.sync_policy",
						"spec.0.sync_policy.0.automated.0.self_heal",
						"true",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.sync_policy",
						"spec.0.sync_policy.0.automated.0.allow_empty",
						"true",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.sync_policy",
						"spec.0.sync_policy.0.retry.0.backoff.0.duration",
						"30s",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.sync_policy",
						"spec.0.sync_policy.0.retry.0.backoff.0.max_duration",
						"2m",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.sync_policy",
						"spec.0.sync_policy.0.retry.0.backoff.0.factor",
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
				ResourceName:            "argocd_application.sync_policy",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait", "cascade", "metadata.0.generation", "metadata.0.resource_version"},
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
				ResourceName:            "argocd_application.ignore_differences",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait", "cascade"},
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
				ResourceName:            "argocd_application.ignore_differences_jqpe",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait", "cascade"},
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
						"spec.0.sync_policy.0.retry.0.backoff.0.duration",
					),
					resource.TestCheckNoResourceAttr(
						"argocd_application.simple",
						"spec.0.sync_policy.0.automated.0.prune",
					),
				),
			},
		},
	})
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
						"spec.0.source.0.directory.0.jsonnet.0",
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
						"spec.0.source.0.directory.0.jsonnet.0",
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
						"spec.0.source.0.directory.0.jsonnet.0",
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
						"spec.0.source.0.directory.0.jsonnet.0",
					),
				),
			},
		},
	})
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
						"spec.0.sync_policy.0.retry.0.backoff.0.duration",
					),
					resource.TestCheckNoResourceAttr(
						"argocd_application.simple",
						"spec.0.sync_policy.0.automated.0.prune",
					),
				),
			},
		},
	})
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
						"spec.0.sync_policy.0.retry.0.backoff.0.duration",
					),
					resource.TestCheckNoResourceAttr(
						"argocd_application.simple",
						"spec.0.sync_policy.0.automated.0.prune",
					),
				),
			},
		},
	})
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
					resource.TestCheckResourceAttrSet(
						"argocd_application.simple",
						"spec.0.sync_policy.0.automated.#",
					),
					resource.TestCheckNoResourceAttr(
						"argocd_application.simple",
						"spec.0.sync_policy.0.automated.0.prune",
					),
				),
			},
		},
	})
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
		},
	})
}

func TestAccArgoCDApplication_Info(t *testing.T) {
	name := acctest.RandomWithPrefix("test-acc-info")
	info := acctest.RandString(8)
	value := acctest.RandString(8)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationInfo(name, info, value),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.info",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.info",
						"spec.0.info.0.name",
						info,
					),
					resource.TestCheckResourceAttr(
						"argocd_application.info",
						"spec.0.info.0.value",
						value,
					),
				),
			},
			{
				Config: testAccArgoCDApplicationInfoNoName(name, value),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.info",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.info",
						"spec.0.info.0.name",
						"",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.info",
						"spec.0.info.0.value",
						value,
					),
				),
			},
			{
				Config: testAccArgoCDApplicationInfoNoValue(name, info),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.info",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.info",
						"spec.0.info.0.name",
						info,
					),
					resource.TestCheckResourceAttr(
						"argocd_application.info",
						"spec.0.info.0.value",
						"",
					),
				),
			},
			{
				Config: testAccArgoCDApplicationInfo(name, info, value),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.info",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.info",
						"spec.0.info.0.name",
						info,
					),
					resource.TestCheckResourceAttr(
						"argocd_application.info",
						"spec.0.info.0.value",
						value,
					),
				),
			},
			{
				Config: testAccArgoCDApplicationNoInfo(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.info",
						"metadata.0.uid",
					),
					resource.TestCheckNoResourceAttr(
						"argocd_application.info",
						"spec.0.info.0",
					),
				),
			},
			{
				Config:      testAccArgoCDApplicationInfoEmpty(name),
				ExpectError: regexp.MustCompile("info: cannot be empty."),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.info",
						"metadata.0.uid",
					),
					resource.TestCheckNoResourceAttr(
						"argocd_application.info",
						"spec.0.info.0",
					),
				),
			},
		},
	})
}

func TestProvider_headers(t *testing.T) {
	name := acctest.RandomWithPrefix("test-acc")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf("%s %s", `
				provider "argocd" {
					headers = [
						"Hello: HiThere",
					]
				}`, testAccArgoCDApplicationSimple(name),
				),
			},
		},
	})
}

func TestAccArgoCDApplication_SkipCrds_NotSupported_On_OlderVersions(t *testing.T) {
	name := acctest.RandomWithPrefix("test-acc-crds")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureNotSupported(t, featureApplicationHelmSkipCrds) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			// Create tests
			{
				Config:      testAccArgoCDApplicationSkipCrds(acctest.RandomWithPrefix("test-acc-crds"), true),
				ExpectError: regexp.MustCompile("application helm skip_crds is only supported from ArgoCD"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.crds",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.crds",
						"spec.0.source.0.helm.0.skip_crds",
						"true",
					),
				),
			},
			// Update tests
			{
				Config: testAccArgoCDApplicationSkipCrds_NoSkip(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.crds",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.crds",
						"spec.0.source.0.helm.0.skip_crds",
						"false",
					),
				),
			},
			{
				Config: testAccArgoCDApplicationSkipCrds(name, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.crds",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.crds",
						"spec.0.source.0.helm.0.skip_crds",
						"false",
					),
				),
			},
			{
				Config:      testAccArgoCDApplicationSkipCrds(name, true),
				ExpectError: regexp.MustCompile("application helm skip_crds is only supported from ArgoCD"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.crds",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.crds",
						"spec.0.source.0.helm.0.skip_crds",
						"true",
					),
				),
			},
		},
	})
}

func TestAccArgoCDApplication_SkipCrds(t *testing.T) {
	name := acctest.RandomWithPrefix("test-acc-crds")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, featureApplicationHelmSkipCrds) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSkipCrds_NoSkip(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.crds",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.crds",
						"spec.0.source.0.helm.0.skip_crds",
						"false",
					),
				),
			},
			{
				Config: testAccArgoCDApplicationSkipCrds(name, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.crds",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.crds",
						"spec.0.source.0.helm.0.skip_crds",
						"true",
					),
				),
			},
			{
				Config: testAccArgoCDApplicationSkipCrds(name, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.crds",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.crds",
						"spec.0.source.0.helm.0.skip_crds",
						"false",
					),
				),
			},
		},
	})
}

func TestAccArgoCDApplication_CustomNamespace(t *testing.T) {
	name := acctest.RandomWithPrefix("test-acc-custom-namespace")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, featureProjectSourceNamespaces) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationCustomNamespace(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.simple",
						"metadata.0.uid",
					),
				),
			},
			{
				ResourceName:            "argocd_application.simple",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait", "cascade"},
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
      repo_url        = "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami"
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
      repo_url        = "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami"
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
      automated {
        prune       = true
        self_heal   = true
        allow_empty   = false
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
      repo_url        = "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami"
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
      repo_url        = "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami"
      chart           = "redis"
      target_revision = "16.9.11"
    }

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
    
    sync_policy {
      automated {
        prune       = true
        self_heal   = true
        allow_empty = true
      }
      retry {
        limit   = "5"
        backoff {
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
      repo_url        = "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami"
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
      repo_url        = "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami"
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
      repo_url        = "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami"
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
      repo_url        = "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami"
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
      repo_url        = "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami"
      chart           = "redis"
      target_revision = "16.9.11"
      helm {
        release_name = "testing"
      }
    }
    sync_policy {
      retry {
        limit   = "5"
        backoff {
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
      repo_url        = "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami"
      chart           = "redis"
      target_revision = "16.9.11"
      helm {
        release_name = "testing"
      }
    }
    sync_policy {
      automated {}
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
      repo_url        = "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami"
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

func testAccArgoCDApplicationInfo(name, info, value string) string {
	return fmt.Sprintf(`
resource "argocd_application" "info" {
  metadata {
    name      = "%s"
    namespace = "argocd"
    labels = {
      acceptance = "true"
    }
  }

  spec {
    info {
      name = "%s"
      value = "%s"
    }
    source {
      repo_url        = "https://github.com/argoproj/argocd-example-apps"
      path            = "guestbook"
      target_revision = "HEAD"
    }

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
  }
}
    `, name, info, value)
}

func testAccArgoCDApplicationInfoNoName(name, value string) string {
	return fmt.Sprintf(`
resource "argocd_application" "info" {
  metadata {
    name      = "%s"
    namespace = "argocd"
    labels = {
      acceptance = "true"
    }
  }

  spec {
    info {
      value = "%s"
    }
    source {
      repo_url        = "https://github.com/argoproj/argocd-example-apps"
      path            = "guestbook"
      target_revision = "HEAD"
    }

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
  }
}
    `, name, value)
}

func testAccArgoCDApplicationInfoNoValue(name, info string) string {
	return fmt.Sprintf(`
resource "argocd_application" "info" {
  metadata {
    name      = "%s"
    namespace = "argocd"
    labels = {
      acceptance = "true"
    }
  }

  spec {
    info {
      name = "%s"
    }
    source {
      repo_url        = "https://github.com/argoproj/argocd-example-apps"
      path            = "guestbook"
      target_revision = "HEAD"
    }

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
  }
}
    `, name, info)
}

func testAccArgoCDApplicationInfoEmpty(name string) string {
	return fmt.Sprintf(`
resource "argocd_application" "info" {
  metadata {
    name      = "%s"
    namespace = "argocd"
    labels = {
      acceptance = "true"
    }
  }

  spec {
    info {
    }
    source {
      repo_url        = "https://github.com/argoproj/argocd-example-apps"
      path            = "guestbook"
      target_revision = "HEAD"
    }

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
  }
}
    `, name)
}

func testAccArgoCDApplicationNoInfo(name string) string {
	return fmt.Sprintf(`
resource "argocd_application" "info" {
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
    }

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
  }
}
    `, name)
}

func testAccArgoCDApplicationSkipCrds(name string, SkipCrds bool) string {
	return fmt.Sprintf(`
resource "argocd_application" "crds" {
  metadata {
    name      = "%s"
    namespace = "argocd"
    labels = {
      acceptance = "true"
    }
  }

  spec {
    source {
		repo_url        = "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami"
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
		  skip_crds = %t
		}
	}
    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
  }
}
	`, name, SkipCrds)
}

func testAccArgoCDApplicationSkipCrds_NoSkip(name string) string {
	return fmt.Sprintf(`
resource "argocd_application" "crds" {
  metadata {
    name      = "%s"
    namespace = "argocd"
    labels = {
      acceptance = "true"
    }
  }

  spec {
    source {
		repo_url        = "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami"
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

func testAccArgoCDApplicationCustomNamespace(name string) string {
	return fmt.Sprintf(`
resource "argocd_project" "simple" {
  metadata {
    name      = "%[1]s"
    namespace = "argocd"
  }

  spec {
    description  = "project with source namespace"
    source_repos = ["*"]
    source_namespaces = ["mynamespace-1"]

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
  }
}

resource "argocd_application" "simple" {
  metadata {
    name      = "%[1]s"
    namespace = "mynamespace-1"
  }

  spec {
    project = argocd_project.simple.metadata[0].name
    source {
      repo_url        = "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami"
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

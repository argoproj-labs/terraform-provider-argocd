package argocd

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/oboukili/terraform-provider-argocd/internal/features"
)

func TestAccArgoCDApplication(t *testing.T) {
	name := acctest.RandomWithPrefix("test-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSimple(name, "8.0.0", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application."+name,
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application."+name,
						"spec.0.source.0.target_revision",
						"8.0.0",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_application."+name,
						"status.0.%",
					),
				),
			},
			{
				ResourceName:            "argocd_application." + name,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait", "cascade", "metadata.0.generation", "metadata.0.resource_version", "status"},
			},
			{
				// Update
				Config: testAccArgoCDApplicationSimple(name, "9.0.0", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application."+name,
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application."+name,
						"spec.0.source.0.target_revision",
						"9.0.0",
					),
				),
			},
			{
				// Update with wait = true
				Config: testAccArgoCDApplicationSimple(name, "9.4.1", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_application."+name,
						"wait",
						"true",
					),
					resource.TestCheckResourceAttr(
						"argocd_application."+name,
						"spec.0.source.0.target_revision",
						"9.4.1",
					),
					resource.TestCheckResourceAttr(
						"argocd_application."+name,
						"status.0.health.0.status",
						"Healthy",
					),
					resource.TestCheckResourceAttr(
						"argocd_application."+name,
						"status.0.sync.0.status",
						"Synced",
					),
				),
			},
			{
				ResourceName:            "argocd_application." + name,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait", "cascade", "metadata.0.generation", "metadata.0.resource_version"},
			},
		},
	})
}

func TestAccArgoCDApplication_Helm(t *testing.T) {
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
					resource.TestCheckResourceAttr(
						"argocd_application.helm",
						"spec.0.source.0.helm.0.pass_credentials",
						"true",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.helm",
						"spec.0.source.0.helm.0.ignore_missing_value_files",
						"true",
					),
				),
			},
			{
				ResourceName:            "argocd_application.helm",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait", "cascade", "metadata.0.generation", "metadata.0.resource_version", "status"},
			},
		},
	})
}

func TestAccArgoCDApplication_Helm_FileParameters(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationHelm_FileParameters(acctest.RandomWithPrefix("test-acc")),
				// Setting up tests for this is non-trivial so it is easier to test for an expected failure (since file does not exist) than to test for success
				ExpectError: regexp.MustCompile(`(?s)Error: failed parsing.*--set-file data`),
			},
		},
	})
}

func TestAccArgoCDApplication_Kustomize(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
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
				ImportStateVerifyIgnore: []string{"wait", "cascade", "metadata.0.generation", "metadata.0.resource_version", "status"},
			},
		},
	})
}

func TestAccArgoCDApplication_IgnoreDifferences(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
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
				ImportStateVerifyIgnore: []string{"wait", "cascade", "status"},
			},
			{
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
				ImportStateVerifyIgnore: []string{"wait", "cascade", "status"},
			},
		},
	})
}

func TestAccArgoCDApplication_RevisionHistoryLimit(t *testing.T) {
	revisionHistoryLimit := acctest.RandIntRange(0, 9)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationRevisionHistory(
					acctest.RandomWithPrefix("test-acc"),
					revisionHistoryLimit,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.revision_history_limit",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.revision_history_limit",
						"spec.0.revision_history_limit",
						fmt.Sprint(revisionHistoryLimit),
					),
				),
			},
			{
				ResourceName:            "argocd_application.revision_history_limit",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait", "cascade", "status"},
			},
		},
	})
}

func TestAccArgoCDApplication_OptionalDestinationNamespace(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplication_OptionalDestinationNamespace(acctest.RandomWithPrefix("test-acc")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.no_namespace",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.no_namespace",
						"spec.0.destination.0.namespace",
						"", // optional strings are maintained in state as blank strings
					),
				),
			},
			{
				ResourceName:            "argocd_application.no_namespace",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait", "cascade", "status"},
			},
		},
	})
}

func TestAccArgoCDApplication_DirectoryJsonnet(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplication_DirectoryJsonnet(
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
					resource.TestCheckResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.directory.0.jsonnet.0.libs.0",
						"vendor",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.directory.0.jsonnet.0.libs.1",
						"foo",
					),
				),
			},
			{
				ResourceName:            "argocd_application.directory",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait", "cascade", "metadata.0.generation", "metadata.0.resource_version", "status"},
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

func TestAccArgoCDApplication_EmptyDirectory(t *testing.T) {
	name := acctest.RandomWithPrefix("test-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplication_EmptyDirectory(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.directory.0.recurse",
					),
				),
			},
			{
				ResourceName:            "argocd_application.directory",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait", "cascade", "metadata.0.generation", "metadata.0.resource_version", "status"},
			},
		},
	})
}

func TestAccArgoCDApplication_DirectoryIncludeExclude(t *testing.T) {
	name := acctest.RandomWithPrefix("test-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplication_DirectoryIncludeExclude(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.directory.0.include",
						"*.yaml",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.directory",
						"spec.0.source.0.directory.0.exclude",
						"config.yaml",
					),
				),
			},
			{
				ResourceName:            "argocd_application.directory",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait", "cascade", "metadata.0.generation", "metadata.0.resource_version", "status"},
			},
		},
	})
}

func TestAccArgoCDApplication_SyncPolicy(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
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
				ImportStateVerifyIgnore: []string{"wait", "cascade", "metadata.0.generation", "metadata.0.resource_version", "status"},
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
						"argocd_application.no_sync_policy",
						"metadata.0.uid",
					),
					resource.TestCheckNoResourceAttr(
						"argocd_application.no_sync_policy",
						"spec.0.sync_policy.0.retry.0.backoff.0.duration",
					),
					resource.TestCheckNoResourceAttr(
						"argocd_application.no_sync_policy",
						"spec.0.sync_policy.0.automated.0.prune",
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
						"argocd_application.empty_sync_policy",
						"metadata.0.uid",
					),
					resource.TestCheckNoResourceAttr(
						"argocd_application.empty_sync_policy",
						"spec.0.sync_policy.0.retry.0.backoff.0.duration",
					),
					resource.TestCheckNoResourceAttr(
						"argocd_application.empty_sync_policy",
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
						"argocd_application.no_automated",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_application.no_automated",
						"spec.0.sync_policy.0.retry.0.backoff.0.duration",
					),
					resource.TestCheckNoResourceAttr(
						"argocd_application.no_automated",
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
						"argocd_application.empty_automated",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_application.empty_automated",
						"spec.0.sync_policy.0.automated.#",
					),
					resource.TestCheckNoResourceAttr(
						"argocd_application.empty_automated",
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

func TestAccArgoCDApplication_SkipCrds(t *testing.T) {
	name := acctest.RandomWithPrefix("test-acc-crds")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
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
	name := acctest.RandomWithPrefix("test-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ProjectSourceNamespaces) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationCustomNamespace(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.custom_namespace",
						"metadata.0.uid",
					),
				),
			},
			{
				ResourceName:            "argocd_application.custom_namespace",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait", "cascade", "status"},
			},
		},
	})
}

func TestAccArgoCDApplication_MultipleSources(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.MultipleApplicationSources) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationMultipleSources(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.multiple_sources",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.multiple_sources",
						"spec.0.source.0.chart",
						"elasticsearch",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.multiple_sources",
						"spec.0.source.1.path",
						"guestbook",
					),
				),
			},
			{
				ResourceName:            "argocd_application.multiple_sources",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait", "cascade", "metadata.0.generation", "metadata.0.resource_version", "status"},
			},
		},
	})
}

func TestAccArgoCDApplication_HelmValuesFromExternalGitRepo(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.MultipleApplicationSources) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationHelmValuesFromExternalGitRepo(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.helm_values_external",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.helm_values_external",
						"spec.0.source.0.chart",
						"wordpress",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_application.helm_values_external",
						"spec.0.source.0.helm.0.value_files.#",
					),
					resource.TestCheckResourceAttr(
						"argocd_application.helm_values_external",
						"spec.0.source.1.ref",
						"values",
					),
				),
			},
			{
				ResourceName:            "argocd_application.helm_values_external",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait", "cascade", "metadata.0.generation", "metadata.0.resource_version", "status"},
			},
		},
	})
}

func TestAccArgoCDApplication_ManagedNamespaceMetadata(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ManagedNamespaceMetadata) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplication_ManagedNamespaceMetadata(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("argocd_application.namespace_metadata", "metadata.0.uid"),
					resource.TestCheckResourceAttrSet("argocd_application.namespace_metadata", "spec.0.sync_policy.0.managed_namespace_metadata.0.annotations.%"),
					resource.TestCheckResourceAttrSet("argocd_application.namespace_metadata", "spec.0.sync_policy.0.managed_namespace_metadata.0.labels.%"),
				),
			},
			{
				ResourceName:            "argocd_application.namespace_metadata",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait", "cascade", "metadata.0.generation", "metadata.0.resource_version"},
			},
		},
	})
}

func TestAccArgoCDApplication_Wait(t *testing.T) {
	chartRevision := "9.4.1"
	name := acctest.RandomWithPrefix("test-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSimple(name, chartRevision, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_application."+name,
						"wait",
						"true",
					),
					resource.TestCheckResourceAttr(
						"argocd_application."+name,
						"spec.0.source.0.target_revision",
						chartRevision,
					),
				),
			},
		},
	})
}

func testAccArgoCDApplicationSimple(name, targetRevision string, wait bool) string {
	return fmt.Sprintf(`
resource "argocd_application" "%[1]s" {
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
    source {
      repo_url        = "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami"
      chart           = "apache"
      target_revision = "%[2]s"
      helm {
        parameter {
          name  = "service.type"
          value = "NodePort"
        }
        release_name = "testing"
      }
    }

    sync_policy {
      automated {
        prune       = true
        self_heal   = true
        allow_empty = false
      }
      sync_options = ["CreateNamespace=true"]
    }

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "%[1]s"
    }
  }
  wait = %[3]t
}
    `, name, targetRevision, wait)
}

func testAccArgoCDApplicationHelm(name, helmValues string) string {
	return fmt.Sprintf(`
resource "argocd_application" "helm" {
  metadata {
    name      = "%[1]s"
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

        pass_credentials = true
        ignore_missing_value_files = true

        value_files = ["values.yaml"]

        values = <<EOT
%[2]s
EOT
      }
    }

    sync_policy {
      sync_options = ["CreateNamespace=true"]
    }

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "%[1]s"
    }
  }
}
	`, name, helmValues)
}

func testAccArgoCDApplicationHelm_FileParameters(name string) string {
	return fmt.Sprintf(`
resource "argocd_application" "helm_file_parameters" {
	metadata {
		name      = "%[1]s"
		namespace = "argocd"
	}

	spec {
		source {
			repo_url        = "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami"
			chart           = "redis"
			target_revision = "16.9.11"

			helm {
				release_name = "testing"
				file_parameter {
					name = "foo"
					path = "does-not-exist.txt"
				}
			}
		}

		sync_policy {
			sync_options = ["CreateNamespace=true"]
		}

		destination {
			server    = "https://kubernetes.default.svc"
			namespace = "%[1]s"
		}
	}
}`, name)
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

func testAccArgoCDApplication_DirectoryJsonnet(name string) string {
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
          libs = [
            "vendor",
            "foo"
          ]
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

func testAccArgoCDApplication_EmptyDirectory(name string) string {
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
      directory {}
    }

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
  }
}
    `, name)
}

func testAccArgoCDApplication_DirectoryIncludeExclude(name string) string {
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
        recurse = true
		exclude = "config.yaml"
		include = "*.yaml"
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

func testAccArgoCDApplication_OptionalDestinationNamespace(name string) string {
	return fmt.Sprintf(`
resource "argocd_application" "no_namespace" {
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
    }
  }
}
`, name)
}

func testAccArgoCDApplicationNoSyncPolicy(name string) string {
	return fmt.Sprintf(`
resource "argocd_application" "no_sync_policy" {
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
resource "argocd_application" "empty_sync_policy" {
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
resource "argocd_application" "no_automated" {
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
resource "argocd_application" "empty_automated" {
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

func testAccArgoCDApplicationRevisionHistory(name string, revision_history_limit int) string {
	return fmt.Sprintf(`
resource "argocd_application" "revision_history_limit" {
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
resource "argocd_project" "custom_namespace" {
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

resource "argocd_application" "custom_namespace" {
  metadata {
    name      = "%[1]s"
    namespace = "mynamespace-1"
  }

  spec {
    project = argocd_project.custom_namespace.metadata[0].name
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

func testAccArgoCDApplicationMultipleSources() string {
	return `
resource "argocd_application" "multiple_sources" {
  metadata {
    name      = "multiple-sources"
    namespace = "argocd"
  }

  spec {
    project = "default" 
	
	source {
		repo_url        = "https://helm.elastic.co"
		chart           = "elasticsearch"
		target_revision = "8.5.1"
	}

	source {
		repo_url        = "https://github.com/argoproj/argocd-example-apps.git"
		path            = "guestbook"
		target_revision = "HEAD"
	}

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
  }
}`
}

func testAccArgoCDApplication_ManagedNamespaceMetadata() string {
	return `
resource "argocd_application" "namespace_metadata" {
	metadata {
		name      = "namespace-metadata"
		namespace = "argocd"
	}

	spec {
		project = "default" 

		source {
			repo_url        = "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami"
			chart           = "apache"
			target_revision = "9.4.1"
		}

		destination {
			server    = "https://kubernetes.default.svc"
			namespace = "managed-namespace"
		}

		sync_policy {
			managed_namespace_metadata {
				annotations = {
					"this.is.a.really.long.nested.key" = "yes, really!"
				}
				labels = {
					foo = "bar"
				}
			}
			sync_options = ["CreateNamespace=true"]
		}
	}
}`
}

func testAccArgoCDApplicationHelmValuesFromExternalGitRepo() string {
	return `
resource "argocd_application" "helm_values_external" {
  metadata {
    name      = "helm-values-external"
    namespace = "argocd"
  }

  spec {
    project = "default" 
  
    source {
      repo_url        = "https://charts.helm.sh/stable"
      chart           = "wordpress"
      target_revision = "9.0.3"
      helm {
        value_files = ["$values/helm-dependency/values.yaml"]
      }
    }

    source {
      repo_url        = "https://github.com/argoproj/argocd-example-apps.git"
      target_revision = "HEAD"
      ref             = "values"
    }

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
  }
}`
}

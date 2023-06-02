package argocd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/oboukili/terraform-provider-argocd/internal/features"
)

func TestAccArgoCDApplicationSet_clusters(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSet) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSet_clusters(),
				Check: resource.TestCheckResourceAttrSet(
					"argocd_application_set.clusters",
					"metadata.0.uid",
				),
			},
			{
				ResourceName:            "argocd_application_set.clusters",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccArgoCDApplicationSet_clustersSelector(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSet) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSet_clustersSelector(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.clusters_selector",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.clusters_selector",
						"spec.0.generator.0.clusters.0.selector.0.match_labels.%",
					),
				),
			},
			{
				ResourceName:            "argocd_application_set.clusters_selector",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccArgoCDApplicationSet_clusterDecisionResource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSet) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSet_clusterDecisionResource(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.cluster_decision_resource",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.cluster_decision_resource",
						"spec.0.generator.0.cluster_decision_resource.0.config_map_ref",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.cluster_decision_resource",
						"spec.0.generator.0.cluster_decision_resource.0.name",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.cluster_decision_resource",
						"spec.0.generator.0.cluster_decision_resource.0.label_selector.0.match_labels.%",
					),
				),
			},
			{
				ResourceName:            "argocd_application_set.cluster_decision_resource",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccArgoCDApplicationSet_gitDirectories(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSet) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSet_scmProviderGitDirectories(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.git_directories",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.git_directories",
						"spec.0.generator.0.git.0.directory.0.path",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.git_directories",
						"spec.0.generator.0.git.0.directory.1.path",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.git_directories",
						"spec.0.generator.0.git.0.directory.1.exclude",
					),
				),
			},
			{
				ResourceName:            "argocd_application_set.git_directories",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccArgoCDApplicationSet_gitFiles(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSet) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSet_scmProviderGitFiles(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.git_files",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.git_files",
						"spec.0.generator.0.git.0.file.0.path",
					),
				),
			},
			{
				ResourceName:            "argocd_application_set.git_files",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccArgoCDApplicationSet_list(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSet) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSet_list(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.list",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.list",
						"spec.0.generator.0.list.0.elements.0.cluster",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.list",
						"spec.0.generator.0.list.0.elements.0.url",
					),
				),
			},
			{
				ResourceName:            "argocd_application_set.list",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccArgoCDApplicationSet_matrix(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSet) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSet_matrix(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.matrix",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.matrix",
						"spec.0.generator.0.matrix.0.generator.0.git.0.directory.0.path",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.matrix",
						"spec.0.generator.0.matrix.0.generator.1.clusters.0.selector.0.match_labels.%",
					),
				),
			},
			{
				ResourceName:            "argocd_application_set.matrix",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccArgoCDApplicationSet_matrixNested(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSet) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSet_matrixNested(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.matrix_nested",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.matrix_nested",
						"spec.0.generator.0.matrix.0.generator.0.clusters.0.selector.0.match_labels.%",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.matrix_nested",
						"spec.0.generator.0.matrix.0.generator.1.matrix.0.generator.0.git.0.repo_url",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.matrix_nested",
						"spec.0.generator.0.matrix.0.generator.1.matrix.0.generator.1.list.0.elements.0.cluster",
					),
				),
			},
			{
				ResourceName:            "argocd_application_set.matrix_nested",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccArgoCDApplicationSet_matrixInvalid(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSet) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccArgoCDApplicationSet_matrixInsufficientGenerators(),
				ExpectError: regexp.MustCompile("Error: Insufficient generator blocks"),
			},
			{
				Config:      testAccArgoCDApplicationSet_matrixTooManyGenerators(),
				ExpectError: regexp.MustCompile("Error: Too many generator blocks"),
			},
			{
				Config:      testAccArgoCDApplicationSet_matrixNestedInsufficientGenerators(),
				ExpectError: regexp.MustCompile("Error: Insufficient generator blocks"),
			},
			{
				Config:      testAccArgoCDApplicationSet_matrixOnly1LevelOfNesting(),
				ExpectError: regexp.MustCompile("Blocks of type \"matrix\" are not expected here."),
			},
		},
	})
}

func TestAccArgoCDApplicationSet_merge(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSet) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSet_merge(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.merge",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application_set.merge",
						"spec.0.generator.0.merge.0.merge_keys.0",
						"server",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.merge",
						"spec.0.generator.0.merge.0.generator.0.clusters.0.values.%",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.merge",
						"spec.0.generator.0.merge.0.generator.1.clusters.0.selector.0.match_labels.%",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.merge",
						"spec.0.generator.0.merge.0.generator.2.list.0.elements.0.server",
					),
				),
			},
			{
				ResourceName:            "argocd_application_set.merge",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccArgoCDApplicationSet_mergeNested(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSet) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSet_mergeNested(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.merge_nested",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application_set.merge_nested",
						"spec.0.generator.0.merge.0.merge_keys.0",
						"server",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.merge_nested",
						"spec.0.generator.0.merge.0.generator.0.list.0.elements.0.server",
					),
					resource.TestCheckResourceAttr(
						"argocd_application_set.merge_nested",
						"spec.0.generator.0.merge.0.generator.1.merge.0.merge_keys.0",
						"server",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.merge_nested",
						"spec.0.generator.0.merge.0.generator.1.merge.0.generator.1.clusters.0.values.%",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.merge_nested",
						"spec.0.generator.0.merge.0.generator.1.merge.0.generator.1.clusters.0.selector.0.match_labels.%",
					),
				),
			},
			{
				ResourceName:            "argocd_application_set.merge_nested",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccArgoCDApplicationSet_scmProviderAzureDevOps(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSet) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSet_scmProviderAzureDevOps(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.scm_ado",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application_set.scm_ado",
						"spec.0.generator.0.scm_provider.0.azure_devops.0.organization",
						"myorg",
					),
				),
			},
			{
				ResourceName:            "argocd_application_set.scm_ado",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccArgoCDApplicationSet_scmProviderBitbucketCloud(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSet) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSet_scmProviderBitbucketCloud(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.scm_bitbucket_cloud",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application_set.scm_bitbucket_cloud",
						"spec.0.generator.0.scm_provider.0.bitbucket_cloud.0.owner",
						"example-owner",
					),
				),
			},
			{
				ResourceName:            "argocd_application_set.scm_bitbucket_cloud",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccArgoCDApplicationSet_scmProviderBitbucketServer(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSet) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSet_scmProviderBitbucketServer(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.scm_bitbucket_server",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application_set.scm_bitbucket_server",
						"spec.0.generator.0.scm_provider.0.bitbucket_server.0.project",
						"myproject",
					),
				),
			},
			{
				ResourceName:            "argocd_application_set.scm_bitbucket_server",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccArgoCDApplicationSet_scmProviderGitea(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSet) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSet_scmProviderGitea(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.scm_gitea",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application_set.scm_gitea",
						"spec.0.generator.0.scm_provider.0.gitea.0.owner",
						"myorg",
					),
				),
			},
			{
				ResourceName:            "argocd_application_set.scm_gitea",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccArgoCDApplicationSet_scmProviderGithub(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSet) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSet_scmProviderGithub(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.scm_github",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application_set.scm_github",
						"spec.0.generator.0.scm_provider.0.github.0.organization",
						"myorg",
					),
				),
			},
			{
				ResourceName:            "argocd_application_set.scm_github",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccArgoCDApplicationSet_scmProviderGitlab(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSet) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSet_scmProviderGitlab(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.scm_gitlab",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application_set.scm_gitlab",
						"spec.0.generator.0.scm_provider.0.gitlab.0.group",
						"8675309",
					),
				),
			},
			{
				ResourceName:            "argocd_application_set.scm_gitlab",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccArgoCDApplicationSet_scmProviderWithFilters(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSet) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSet_scmProviderWithFilters(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.scm_filters",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application_set.scm_filters",
						"spec.0.generator.0.scm_provider.0.filter.0.repository_match",
						"^myapp",
					),
					resource.TestCheckResourceAttr(
						"argocd_application_set.scm_filters",
						"spec.0.generator.0.scm_provider.0.filter.0.paths_exist.0",
						"kubernetes/kustomization.yaml",
					),
					resource.TestCheckResourceAttr(
						"argocd_application_set.scm_filters",
						"spec.0.generator.0.scm_provider.0.filter.1.repository_match",
						"^otherapp",
					),
					resource.TestCheckResourceAttr(
						"argocd_application_set.scm_filters",
						"spec.0.generator.0.scm_provider.0.filter.1.paths_do_not_exist.0",
						"disabledrepo.txt",
					),
				),
			},
			{
				ResourceName:            "argocd_application_set.scm_filters",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccArgoCDApplicationSet_pullRequestBitbucketServer(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSet) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSet_pullRequestBitbucketServer(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.pr_bitbucket_server",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application_set.pr_bitbucket_server",
						"spec.0.generator.0.pull_request.0.bitbucket_server.0.project",
						"myproject",
					),
				),
			},
			{
				ResourceName:            "argocd_application_set.pr_bitbucket_server",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccArgoCDApplicationSet_pullRequestGitea(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSet) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSet_pullRequestGitea(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.pr_gitea",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application_set.pr_gitea",
						"spec.0.generator.0.pull_request.0.gitea.0.owner",
						"myorg",
					),
				),
			},
			{
				ResourceName:            "argocd_application_set.pr_gitea",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccArgoCDApplicationSet_pullRequestGithub(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSet) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSet_pullRequestGithub(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.pr_github",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application_set.pr_github",
						"spec.0.generator.0.pull_request.0.github.0.owner",
						"myorg",
					),
					resource.TestCheckResourceAttr(
						"argocd_application_set.pr_github",
						"spec.0.generator.0.pull_request.0.github.0.labels.0",
						"preview",
					),
				),
			},
			{
				ResourceName:            "argocd_application_set.pr_github",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccArgoCDApplicationSet_pullRequestGitlab(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSet) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSet_pullRequestGitlab(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.pr_gitlab",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application_set.pr_gitlab",
						"spec.0.generator.0.pull_request.0.gitlab.0.project",
						"myproject",
					),
					resource.TestCheckResourceAttr(
						"argocd_application_set.pr_gitlab",
						"spec.0.generator.0.pull_request.0.gitlab.0.labels.0",
						"preview",
					),
				),
			},
			{
				ResourceName:            "argocd_application_set.pr_gitlab",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccArgoCDApplicationSet_mergeInvalid(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSet) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccArgoCDApplicationSet_mergeInsufficientGenerators(),
				ExpectError: regexp.MustCompile("Error: Insufficient generator blocks"),
			},
			{
				Config:      testAccArgoCDApplicationSet_mergeNestedInsufficientGenerators(),
				ExpectError: regexp.MustCompile("Error: Insufficient generator blocks"),
			},
			{
				Config:      testAccArgoCDApplicationSet_mergeOnly1LevelOfNesting(),
				ExpectError: regexp.MustCompile("Blocks of type \"merge\" are not expected here."),
			},
		},
	})
}

func TestAccArgoCDApplicationSet_generatorTemplate(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSet) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSet_generatorTemplate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.generator_template",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application_set.generator_template",
						"spec.0.generator.0.list.0.elements.0.cluster",
						"engineering-dev",
					),
					resource.TestCheckResourceAttr(
						"argocd_application_set.generator_template",
						"spec.0.generator.0.list.0.template.0.spec.0.project",
						"default",
					),
				),
			},
			{
				ResourceName:            "argocd_application_set.generator_template",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccArgoCDApplicationSet_goTemplate(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSet) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSet_goTemplate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.go_template",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application_set.go_template",
						"spec.0.go_template",
						"true",
					),
				),
			},
			{
				ResourceName:            "argocd_application_set.go_template",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccArgoCDApplicationSet_syncPolicy(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSet) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSet_syncPolicy(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.sync_policy",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application_set.sync_policy",
						"spec.0.sync_policy.0.preserve_resources_on_deletion",
						"true",
					),
				),
			},
			{
				ResourceName:            "argocd_application_set.sync_policy",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccArgoCDApplicationSet_progressiveSync(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.ApplicationSetProgressiveSync) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDApplicationSet_progressiveSync(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application_set.progressive_sync",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_application_set.progressive_sync",
						"spec.0.strategy.0.type",
						"RollingSync",
					),
				),
			},
			{
				ResourceName:            "argocd_application_set.progressive_sync",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func testAccArgoCDApplicationSet_clusters() string {
	return `
resource "argocd_application_set" "clusters" {
	metadata {
		name = "clusters"
	}
	
	spec {
		generator {
			clusters {} # Automatically use all clusters defined within Argo CD
		}
	
		template {
			metadata {
				name = "{{name}}-clusters"
			}
		
			spec {
				source {
					repo_url        = "https://github.com/argoproj/argocd-example-apps/"
					target_revision = "HEAD"
					path            = "guestbook"
				}
		
				destination {
					server    = "{{server}}"
					namespace = "default"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_clustersSelector() string {
	return `
resource "argocd_application_set" "clusters_selector" {
	metadata {
		name = "clusters-selector"
	}
	
	spec {
		generator {
			clusters {
				selector {
					match_labels = {
						"argocd.argoproj.io/secret-type" = "cluster"
					}
				}
			} 
		}
	
		template {
			metadata {
				name = "{{name}}-clusters-selector"
			}
		
			spec {
				source {
					repo_url        = "https://github.com/argoproj/argocd-example-apps/"
					target_revision = "HEAD"
					path = "guestbook"
				}
		
				destination {
					server    = "{{server}}"
					namespace = "default"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_clusterDecisionResource() string {
	return `
resource "argocd_application_set" "cluster_decision_resource" {
	metadata {
		name = "cluster-decision-resource"
	}
	
	spec {
		generator {
			cluster_decision_resource {
				config_map_ref        = "my-configmap"
				name                  = "quak"
				requeue_after_seconds = "180"

				label_selector {
					match_labels = {
						duck = "spotted"
					}

					match_expressions {
						key = "duck"
						operator = "In"
						values = [
							"spotted",
							"canvasback"
						]
					} 
				}
			} 
		}
	
		template {
			metadata {
				name = "{{name}}-cluster-decision-resource"
			}
		
			spec {
				source {
					repo_url        = "https://github.com/argoproj/argocd-example-apps/"
					target_revision = "HEAD"
					path            = "guestbook"
				}
		
				destination {
					server    = "{{server}}"
					namespace = "default"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_scmProviderGitDirectories() string {
	return `
resource "argocd_application_set" "git_directories" {
	metadata {
		name = "git-directories"
	}
	
	spec {
		generator {
			git {
				repo_url = "https://github.com/argoproj/argo-cd.git"
      			revision = "HEAD"
				
				directory {
					path = "applicationset/examples/git-generator-directory/cluster-addons/*"
				}
				
				directory {
					path = "applicationset/examples/git-generator-directory/excludes/cluster-addons/exclude-helm-guestbook"
					exclude = true
				}
			} 
		}
	
		template {
			metadata {
				name = "{{path.basename}}-git-directories"
			}
		
			spec {
				source {
					repo_url        = "https://github.com/argoproj/argo-cd.git"
					target_revision = "HEAD"
					path            = "{{path}}"
				}
		
				destination {
					server    = "https://kubernetes.default.svc"
					namespace = "{{path.basename}}"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_scmProviderGitFiles() string {
	return `
resource "argocd_application_set" "git_files" {
	metadata {
		name = "git-files"
	}
	
	spec {
		generator {
			git {
				repo_url = "https://github.com/argoproj/argo-cd.git"
      			revision = "HEAD"
				
				file {
					path = "applicationset/examples/git-generator-files-discovery/cluster-config/**/config.json"
				}
			} 
		}
	
		template {
			metadata {
				name = "{{cluster.name}}-git-files"
			}
		
			spec {
				source {
					repo_url        = "https://github.com/argoproj/argo-cd.git"
					target_revision = "HEAD"
					path            = "applicationset/examples/git-generator-files-discovery/apps/guestbook"
				}
		
				destination {
					server    = "{{cluster.address}}"
					namespace = "guestbook"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_list() string {
	return `
resource "argocd_application_set" "list" {
	metadata {
		name = "list"
	}
	
	spec {
		generator {
			list {
				elements = [
					{
						cluster = "engineering-dev"
						url     = "https://kubernetes.default.svc"
					}
				]
			}
		}
	
		template {
			metadata {
				name = "{{cluster}}-guestbook"
			}
		
			spec {
				project = "default"
		
				source {
					repo_url        = "https://github.com/argoproj/argo-cd.git"
					target_revision = "HEAD"
					path            = "applicationset/examples/list-generator/guestbook/{{cluster}}"
				}
		
				destination {
					server    = "{{url}}"
					namespace = "guestbook"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_matrix() string {
	return `
resource "argocd_application_set" "matrix" {
	metadata {
		name = "matrix"
	}
	
	spec {
		generator {
			matrix {
				generator {
					git {
						repo_url = "https://github.com/argoproj/argo-cd.git"
						revision = "HEAD"

						directory {
							path = "applicationset/examples/matrix/cluster-addons/*"
						}
					}
				}

				generator {
					clusters{
						selector{
							match_labels = {
								"argocd.argoproj.io/secret-type" = "cluster"
							}
						}
					}
				}
			}
		}
	
		template {
			metadata {
				name = "{{path.basename}}-{{name}}"
			}
		
			spec {
				project = "default"
		
				source {
					repo_url        = "https://github.com/argoproj/argo-cd.git"
					target_revision = "HEAD"
					path            = "{{path}}"
				}
		
				destination {
					server    = "{{server}}"
					namespace = "{{path.basename}}"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_matrixNested() string {
	return `
resource "argocd_application_set" "matrix_nested" {
	metadata {
		name = "matrix-nested"
	}
	
	spec {
		generator {
			matrix {
				generator {
					clusters{
						selector{
							match_labels = {
								"argocd.argoproj.io/secret-type" = "cluster"
							}
						}
					}
				}

				generator {
					matrix {
						generator {
							git {
								repo_url = "https://github.com/argoproj/argo-cd.git"
								revision = "HEAD"
		
								directory {
									path = "applicationset/examples/matrix/cluster-addons/*"
								}
							}
						}

						generator {
							list {
								elements = [
									{
										cluster = "engineering-dev"
										url     = "https://kubernetes.default.svc"
									}
								]
							}
						}
					}
				}
			}
		}
	
		template {
			metadata {
				name = "nested-{{path.basename}}-{{name}}"
			}
		
			spec {
				project = "default"
		
				source {
					repo_url        = "https://github.com/argoproj/argo-cd.git"
					target_revision = "HEAD"
					path            = "{{path}}"
				}
		
				destination {
					server    = "{{server}}"
					namespace = "{{path.basename}}"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_matrixInsufficientGenerators() string {
	return `
resource "argocd_application_set" "matrix_insufficient_generators" {
	metadata {
		name = "matrix-insufficient-generators"
	}
	
	spec {
		generator {
			matrix {
				generator {
					git {
						repo_url = "https://github.com/argoproj/argo-cd.git"
						revision = "HEAD"

						directory {
							path = "applicationset/examples/matrix/cluster-addons/*"
						}
					}
				}
			}
		}
	
		template {
			metadata {
				name = "{{path.basename}}-{{name}}"
			}
		
			spec {
				project = "default"
		
				source {
					repo_url        = "https://github.com/argoproj/argo-cd.git"
					target_revision = "HEAD"
					path            = "{{path}}"
				}
		
				destination {
					server    = "{{server}}"
					namespace = "{{path.basename}}"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_matrixTooManyGenerators() string {
	return `
resource "argocd_application_set" "matrix_too_many_generators" {
	metadata {
		name = "matrix-too-many-generators"
	}
	
	spec {
		generator {
			matrix {
				generator {
					git {
						repo_url = "https://github.com/argoproj/argo-cd.git"
						revision = "HEAD"

						directory {
							path = "applicationset/examples/matrix/cluster-addons/*"
						}
					}
				}

				generator {
					git {
						repo_url = "https://github.com/argoproj/argo-cd.git"
						revision = "HEAD"

						directory {
							path = "applicationset/examples/matrix/cluster-addons/*"
						}
					}
				}

				generator {
					git {
						repo_url = "https://github.com/argoproj/argo-cd.git"
						revision = "HEAD"

						directory {
							path = "applicationset/examples/matrix/cluster-addons/*"
						}
					}
				}
			}
		}
	
		template {
			metadata {
				name = "{{path.basename}}-{{name}}"
			}
		
			spec {
				project = "default"
		
				source {
					repo_url        = "https://github.com/argoproj/argo-cd.git"
					target_revision = "HEAD"
					path            = "{{path}}"
				}
		
				destination {
					server    = "{{server}}"
					namespace = "{{path.basename}}"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_matrixNestedInsufficientGenerators() string {
	return `
resource "argocd_application_set" "matrix_nested_insufficient_generators" {
	metadata {
		name = "matrix-nested-insufficient-generators"
	}
	
	spec {
		generator {
			matrix {
				generator {
					clusters{
						selector{
							match_labels = {
								"argocd.argoproj.io/secret-type" = "cluster"
							}
						}
					}
				}

				generator {
					matrix {
						generator {
							git {
								repo_url = "https://github.com/argoproj/argo-cd.git"
								revision = "HEAD"
		
								directory {
									path = "applicationset/examples/matrix/cluster-addons/*"
								}
							}
						}
					}
				}
			}
		}
	
		template {
			metadata {
				name = "nested-{{path.basename}}-{{name}}"
			}
		
			spec {
				project = "default"
		
				source {
					repo_url        = "https://github.com/argoproj/argo-cd.git"
					target_revision = "HEAD"
					path            = "{{path}}"
				}
		
				destination {
					server    = "{{server}}"
					namespace = "{{path.basename}}"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_matrixOnly1LevelOfNesting() string {
	return `
resource "argocd_application_set" "matrix_nested_invalid" {
	metadata {
		name = "matrix-nested-invalid"
	}
	
	spec {
		generator {
			matrix {
				generator {
					clusters{
						selector{
							match_labels = {
								"argocd.argoproj.io/secret-type" = "cluster"
							}
						}
					}
				}

				generator {
					matrix {
						generator {
							git {
								repo_url = "https://github.com/argoproj/argo-cd.git"
								revision = "HEAD"
		
								directory {
									path = "applicationset/examples/matrix/cluster-addons/*"
								}
							}
						}

						generator {
							matrix {
								generator {
									git {
										repo_url = "https://github.com/argoproj/argo-cd.git"
										revision = "HEAD"
				
										directory {
											path = "applicationset/examples/matrix/cluster-addons/*"
										}
									}
								}
		
								generator {
									list {
										elements = [
											{
												cluster = "engineering-dev"
												url     = "https://kubernetes.default.svc"
											}
										]
									}
								}
							}
						}
					}
				}
			}
		}
	
		template {
			metadata {
				name = "nested-{{path.basename}}-{{name}}"
			}
		
			spec {
				project = "default"
		
				source {
					repo_url        = "https://github.com/argoproj/argo-cd.git"
					target_revision = "HEAD"
					path            = "{{path}}"
				}
		
				destination {
					server    = "{{server}}"
					namespace = "{{path.basename}}"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_merge() string {
	return `
resource "argocd_application_set" "merge" {
	metadata {
		name = "merge"
	}
	  
	spec {
		generator {
			merge {
				merge_keys = [
					"server"
				]
	  
				generator {
					clusters {
						values = {
							kafka = true
							redis = false
						}
					}
				}
	  
				generator {
					clusters {
						selector {
							match_labels = {
								use-kafka = "false"
							}
						}
		
						values = {
							kafka = "false"
						}
					}
				}
	  
				generator {
					list {
						elements = [
							{
								server         = "https://2.4.6.8"
								"values.redis" = "true"
							},
						]
					}
				}
			}
		}
	  
		template {
			metadata {
				name = "{{name}}"
			}
	  
			spec {
			  	project = "default"
	  
			  	source {
					repo_url        = "https://github.com/argoproj/argo-cd.git"
					path            = "app"
					target_revision = "HEAD"
	  
					helm {
						parameter {
							name  = "kafka"
							value = "{{values.kafka}}"
						}
	  
						parameter {
							name  = "redis"
							value = "{{values.redis}}"
						}
					}
			  	}
	  
				destination {
					server    = "{{server}}"
					namespace = "default"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_mergeNested() string {
	return `
resource "argocd_application_set" "merge_nested" {
	metadata {
		name = "merge-nested"
	}
	  
	spec {
		generator {
			merge {
				merge_keys = [
					"server"
				]
	  
				generator {
					list {
						elements = [
							{
								server         = "https://2.4.6.8"
								"values.redis" = "true"
							},
						]
					}
				}
	  
				generator {
					merge {
						merge_keys = [
							"server"
						]

						generator {
							clusters {
								values = {
									kafka = true
									redis = false
								}
							}
						}
	  
						generator {
							clusters {
								selector {
									match_labels = {
										use-kafka = "false"
									}
								}
				
								values = {
									kafka = "false"
								}
							}
						}
					}
				}
			}
		}
	  
		template {
			metadata {
				name = "{{name}}"
			}
	  
			spec {
			  	project = "default"
	  
			  	source {
					repo_url        = "https://github.com/argoproj/argo-cd.git"
					path            = "app"
					target_revision = "HEAD"
	  
					helm {
						parameter {
							name  = "kafka"
							value = "{{values.kafka}}"
						}
	  
						parameter {
							name  = "redis"
							value = "{{values.redis}}"
						}
					}
			  	}
	  
				destination {
					server    = "{{server}}"
					namespace = "default"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_mergeInsufficientGenerators() string {
	return `
resource "argocd_application_set" "merge_insufficient_generators" {
	metadata {
		name = "merge-insufficient-generators"
	}
		
	spec {
		generator {
			merge {
				merge_keys = [
					"server"
				]
		
				generator {
					clusters {
						values = {
							kafka = true
							redis = false
						}
					}
				}
			}
		}
		
		template {
			metadata {
				name = "{{name}}"
			}
		
			spec {
					project = "default"
		
					source {
					repo_url        = "https://github.com/argoproj/argo-cd.git"
					path            = "app"
					target_revision = "HEAD"
		
					helm {
						parameter {
							name  = "kafka"
							value = "{{values.kafka}}"
						}
		
						parameter {
							name  = "redis"
							value = "{{values.redis}}"
						}
					}
					}
		
				destination {
					server    = "{{server}}"
					namespace = "default"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_mergeNestedInsufficientGenerators() string {
	return `
resource "argocd_application_set" "merge_nested_insufficient_generators" {
	metadata {
		name = "merge-nested-insufficient-generators"
	}
		
	spec {
		generator {
			merge {
				merge_keys = [
					"server"
				]
		
				generator {
					list {
						elements = [
							{
								server         = "https://2.4.6.8"
								"values.redis" = "true"
							},
						]
					}
				}
		
				generator {
					merge {
						merge_keys = [
							"server"
						]

						generator {
							clusters {
								values = {
									kafka = true
									redis = false
								}
							}
						}
					}
				}
			}
		}
		
		template {
			metadata {
				name = "{{name}}"
			}
		
			spec {
					project = "default"
		
					source {
					repo_url        = "https://github.com/argoproj/argo-cd.git"
					path            = "app"
					target_revision = "HEAD"
		
					helm {
						parameter {
							name  = "kafka"
							value = "{{values.kafka}}"
						}
		
						parameter {
							name  = "redis"
							value = "{{values.redis}}"
						}
					}
					}
		
				destination {
					server    = "{{server}}"
					namespace = "default"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_mergeOnly1LevelOfNesting() string {
	return `
resource "argocd_application_set" "merge_nested_invalid" {
	metadata {
		name = "merge-nested-invalid"
	}
	
	spec {
		generator {
			merge {
				merge_keys = [
					"server"
				]
	
				generator {
					list {
						elements = [
							{
								server         = "https://2.4.6.8"
								"values.redis" = "true"
							},
						]
					}
				}
	
				generator {
					merge {
						merge_keys = [
							"server"
						]

						generator {
							clusters {
								values = {
									kafka = true
									redis = false
								}
							}
						}
	
						generator {
							merge {
								merge_keys = [
									"server"
								]
		
								generator {
									clusters {
										values = {
											kafka = true
											redis = false
										}
									}
								}
			
								generator {
									clusters {
										selector {
											match_labels = {
												use-kafka = "false"
											}
										}
						
										values = {
											kafka = "false"
										}
									}
								}
							}
						}
					}
				}
			}
		}
	
		template {
			metadata {
				name = "{{name}}"
			}
	
			spec {
				project = "default"
	
				source {
					repo_url        = "https://github.com/argoproj/argo-cd.git"
					path            = "app"
					target_revision = "HEAD"
	
					helm {
						parameter {
							name  = "kafka"
							value = "{{values.kafka}}"
						}
	
						parameter {
							name  = "redis"
							value = "{{values.redis}}"
						}
					}
				}
	
				destination {
					server    = "{{server}}"
					namespace = "default"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_scmProviderAzureDevOps() string {
	return `
resource "argocd_application_set" "scm_ado" {
	metadata {
		name = "scm-ado"
	}
	  
	spec {
		generator {
			scm_provider {
				azure_devops {
					all_branches = true
					api          = "https://dev.azure.com"
					organization = "myorg"
					team_project = "myProject"

					access_token_ref {
						secret_name = "azure-devops-scm"
						key         =  "accesstoken"
					}
				}
			}
		}
	  
		template {
			metadata {
				name = "{{repository}}"
			}
	  
			spec {
			  	project = "default"
	  
			  	source {
					repo_url        = "{{url}}"
					path            = "kubernetes/"
					target_revision = "{{branch}}"
			  	}
	  
				destination {
					server    = "https://kubernetes.default.svc"
					namespace = "default"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_scmProviderBitbucketCloud() string {
	return `
resource "argocd_application_set" "scm_bitbucket_cloud" {
	metadata {
		name = "scm-bitbucket-cloud"
	}
	
	spec {
		generator {
			scm_provider {
				bitbucket_cloud {
					all_branches = true
					owner        = "example-owner"
					user         = "example-user"
			
					app_password_ref {
						secret_name = "appPassword"
						key         = "password"
					}
				}
			}
		}
	
		template {
			metadata {
				name = "{{repository}}"
			}
		
			spec {
				project = "default"
		
				source {
					repo_url        = "{{url}}"
					path            = "kubernetes/"
					target_revision = "{{branch}}"
				}
		
				destination {
					server    = "https://kubernetes.default.svc"
					namespace = "default"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_scmProviderBitbucketServer() string {
	return `
resource "argocd_application_set" "scm_bitbucket_server" {
	metadata {
		name = "scm-bitbucket-server"
	}
	
	spec {
		generator {
			scm_provider {
				bitbucket_server {
					all_branches = true
					api          = "https://bitbucket.org/rest"
					project      = "myproject"
			
					basic_auth {
						username = "myuser"
						password_ref {
							secret_name = "mypassword"
							key         = "password"
						}
					}
				}
			}
		}
	
		template {
			metadata {
				name = "{{repository}}"
			}
		
			spec {
				project = "default"
		
				source {
					repo_url        = "{{url}}"
					path            = "kubernetes/"
					target_revision = "{{branch}}"
				}
		
				destination {
					server    = "https://kubernetes.default.svc"
					namespace = "default"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_scmProviderGitea() string {
	return `
resource "argocd_application_set" "scm_gitea" {
	metadata {
		name = "scm-gitea"
	}
	  
	spec {
		generator {
			scm_provider {
				gitea {
					all_branches = true
					owner        = "myorg"
					api          = "https://gitea.mydomain.com/"
		
					token_ref {
					secret_name = "gitea-token"
					key         = "token"
					}
				}
			}
		}
	  
		template {
			metadata {
				name = "{{repository}}"
			}
		
			spec {
				project = "default"
		
				source {
					repo_url        = "{{url}}"
					path            = "kubernetes/"
					target_revision = "{{branch}}"
				}
		
				destination {
					server    = "https://kubernetes.default.svc"
					namespace = "default"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_scmProviderGithub() string {
	return `
resource "argocd_application_set" "scm_github" {
	metadata {
		name = "scm-github"
	}
	  
	spec {
		generator {
			scm_provider {
				github {
					all_branches    = true
					api             = "https://git.example.com/"
					app_secret_name = "gh-app-repo-creds"
					organization    = "myorg"
		
					token_ref {
						secret_name = "github-token"
						key         = "token"
					}
				}
			}
		}
	  
		template {
			metadata {
				name = "{{repository}}"
			}
		
			spec {
				project = "default"
		
				source {
					repo_url        = "{{url}}"
					path            = "kubernetes/"
					target_revision = "{{branch}}"
				}
		
				destination {
					server    = "https://kubernetes.default.svc"
					namespace = "default"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_scmProviderGitlab() string {
	return `
resource "argocd_application_set" "scm_gitlab" {
	metadata {
		name = "scm-gitlab"
	}
	  
	spec {
		generator {
			scm_provider {
				gitlab {
					all_branches      = true
					api               = "https://gitlab.example.com/"
					group             = "8675309"
					include_subgroups = false
			
					token_ref {
						secret_name = "gitlab-token"
						key         = "token"
					}
				}
			}
		}
	  
		template {
			metadata {
			  	name = "{{repository}}"
			}
	  
			spec {
			  	project = "default"
	  
				source {
					repo_url        = "{{url}}"
					path            = "kubernetes/"
					target_revision = "{{branch}}"
				}
	  
				destination {
					server    = "https://kubernetes.default.svc"
					namespace = "default"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_scmProviderWithFilters() string {
	return `
resource "argocd_application_set" "scm_filters" {
	metadata {
		name = "scm-filters"
	}
	  
	spec {
		generator {
			scm_provider {
				github {
					all_branches    = true
					api             = "https://git.example.com/"
					app_secret_name = "gh-app-repo-creds"
					organization    = "myorg"
		
					token_ref {
						secret_name = "github-token"
						key         = "token"
					}
				}

				filter {
					repository_match = "^myapp"
					paths_exist = [
						"kubernetes/kustomization.yaml"
					]
					label_match = "deploy-ok"
				}

				filter {
					repository_match = "^otherapp"
					paths_exist = [
						"helm"
					]
					paths_do_not_exist = [
						"disabledrepo.txt"
					]
				}
			}
		}
	  
		template {
			metadata {
				name = "{{repository}}"
			}
		
			spec {
				project = "default"
		
				source {
					repo_url        = "{{url}}"
					path            = "kubernetes/"
					target_revision = "{{branch}}"
				}
		
				destination {
					server    = "https://kubernetes.default.svc"
					namespace = "default"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_pullRequestBitbucketServer() string {
	return `
resource "argocd_application_set" "pr_bitbucket_server" {
	metadata {
		name = "pr-bitbucket-server"
	}
	
	spec {
		generator {
			pull_request {
				bitbucket_server {
					api     = "https://bitbucket.org/rest"
					project = "myproject"
					repo    = "myrepository"
			
					basic_auth {
						username = "myuser"
						password_ref {
							secret_name = "mypassword"
							key         = "password"
						}
					}
				}
			}
		}
	
		template {
			metadata {
				name = "myapp-{{branch}}-{{number}}"
			}
		
			spec {
				project = "default"
		
				source {
					repo_url        = "https://github.com/myorg/myrepo.git"
					path            = "kubernetes/"
					target_revision = "{{head_sha}}"

					helm {
						parameter {
							name  = "image.tag"
							value = "pull-{{head_sha}}"
						}
					}
				}
		
				destination {
					server    = "https://kubernetes.default.svc"
					namespace = "default"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_pullRequestGitea() string {
	return `
resource "argocd_application_set" "pr_gitea" {
	metadata {
		name = "pr-gitea"
	}
	
	spec {
		generator {
			pull_request {
				gitea {
					api      = "https://gitea.mydomain.com/"
					insecure = true
					owner    = "myorg"
					repo     = "myrepository"
			
					token_ref {
						secret_name = "gitea-token"
						key         = "token"
					}
				}
			}
		}
	
		template {
			metadata {
				name = "myapp-{{branch}}-{{number}}"
			}
		
			spec {
				project = "default"
		
				source {
					repo_url        = "https://github.com/myorg/myrepo.git"
					path            = "kubernetes/"
					target_revision = "{{head_sha}}"

					helm {
						parameter {
							name  = "image.tag"
							value = "pull-{{head_sha}}"
						}
					}
				}
		
				destination {
					server    = "https://kubernetes.default.svc"
					namespace = "default"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_pullRequestGithub() string {
	return `
resource "argocd_application_set" "pr_github" {
	metadata {
		name = "pr-github"
	}
	
	spec {
		generator {
			pull_request {
				github {
					api             = "https://git.example.com/"
					owner           = "myorg"
					repo            = "myrepository"
					app_secret_name = "github-app-repo-creds"
			
					token_ref {
						secret_name = "github-token"
						key         = "token"
					}

					labels = [
						"preview"
					]
				}
			}
		}
	
		template {
			metadata {
				name = "myapp-{{branch}}-{{number}}"
			}
		
			spec {
				project = "default"
		
				source {
					repo_url        = "https://github.com/myorg/myrepo.git"
					path            = "kubernetes/"
					target_revision = "{{head_sha}}"

					helm {
						parameter {
							name  = "image.tag"
							value = "pull-{{head_sha}}"
						}
					}
				}
		
				destination {
					server    = "https://kubernetes.default.svc"
					namespace = "default"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_pullRequestGitlab() string {
	return `
resource "argocd_application_set" "pr_gitlab" {
	metadata {
		name = "pr-gitlab"
	}
	
	spec {
		generator {
			pull_request {
				gitlab {
					api                = "https://git.example.com/"
					project            = "myproject"
					pull_request_state = "opened"
			
					token_ref {
						secret_name = "gitlab-token"
						key         = "token"
					}

					labels = [
						"preview"
					]
				}
			}
		}
	
		template {
			metadata {
				name = "myapp-{{branch}}-{{number}}"
			}
		
			spec {
				project = "default"
		
				source {
					repo_url        = "https://github.com/myorg/myrepo.git"
					path            = "kubernetes/"
					target_revision = "{{head_sha}}"

					helm {
						parameter {
							name  = "image.tag"
							value = "pull-{{head_sha}}"
						}
					}
				}
		
				destination {
					server    = "https://kubernetes.default.svc"
					namespace = "default"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_generatorTemplate() string {
	return `
resource "argocd_application_set" "generator_template" {
	metadata {
		name = "generator-template"
	}
	
	spec {
		generator {
			list {
				elements = [
					{
						cluster = "engineering-dev"
						url     = "https://kubernetes.default.svc"
					}
				]

				template {
					metadata {}
					spec {
						project = "default"
						source {
							repo_url        = "https://github.com/argoproj/argo-cd.git"
							target_revision = "HEAD"
							path            = "applicationset/examples/template-override/{{.cluster}}-override"
						} 
						destination {}
					}
				}
			} 
		}

		go_template = true

		template {
			metadata {
				name = "appset-generator-template-{{.cluster}}"
			}
		
			spec {
				project = "default"

				source {
					repo_url        = "https://github.com/argoproj/argo-cd.git"
					target_revision = "HEAD"
					path            = "applicationset/examples/template-override/default"
				}
		
				destination {
					server    = "{{.url}}"
					namespace = "guestbook"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_goTemplate() string {
	return `
resource "argocd_application_set" "go_template" {
	metadata {
		name = "go-template"
	}
	
	spec {
		generator {
			clusters {} # Automatically use all clusters defined within Argo CD
		}

		go_template = true
	
		template {
			metadata {
				name = "appset-go-template-{{.name}}"
			}
		
			spec {
				source {
					repo_url        = "https://github.com/argoproj/argocd-example-apps/"
					target_revision = "HEAD"
					path            = "guestbook"
				}
		
				destination {
					server    = "{{.server}}"
					namespace = "default"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_syncPolicy() string {
	return `
resource "argocd_application_set" "sync_policy" {
	metadata {
		name = "sync-policy"
	}
	
	spec {
		generator {
			clusters {} # Automatically use all clusters defined within Argo CD
		}

		sync_policy {
			preserve_resources_on_deletion = true
		}
	
		template {
			metadata {
				name = "appset-sync-policy-{{name}}"
			}
		
			spec {
				source {
					repo_url        = "https://github.com/argoproj/argocd-example-apps/"
					target_revision = "HEAD"
					path            = "guestbook"
				}
		
				destination {
					server    = "{{server}}"
					namespace = "default"
				}
			}
		}
	}
}`
}

func testAccArgoCDApplicationSet_progressiveSync() string {
	return `
resource "argocd_application_set" "progressive_sync" {
	metadata {
		name = "progressive-sync"
	}
	
	spec {
		generator {
			list {
				elements = [
					{
						cluster = "engineering-dev"
						url     = "https://1.2.3.4"
						env     = "env-dev"
					},
					{
						cluster = "engineering-qa"
						url     = "https://2.4.6.8"
						env     = "env-qa"
					},
					{
						cluster = "engineering-prod"
						url     = "https://9.8.7.6/"
						env     = "env-prod"
					}
				]
			}
		}
	  
		strategy {
			type = "RollingSync"
			rolling_sync {
				step {
					match_expressions {
						key      = "envLabel"
						operator = "In"
						values = [
							"env-dev"
						]
					}
			
					# max_update = "100%"  # if undefined, all applications matched are updated together (default is 100%)
				}
		
				step {
					match_expressions {
						key      = "envLabel"
						operator = "In"
						values = [
							"env-qa"
						]
					}
			
					max_update = "0"
				}
		
				step {
					match_expressions {
						key      = "envLabel"
						operator = "In"
						values = [
							"env-prod"
						]
					}
			
					max_update = "10%"
				}
			}
		}
	  
		go_template = true
	  
		template {
			metadata {
				name = "appset-progressive-sync-{{.cluster}}"
				labels = {
					envLabel = "{{.env}}"
				}
			}
	  
			spec {
			  	project = "default"
	  
				source {
					repo_url        = "https://github.com/infra-team/cluster-deployments.git"
					path            = "guestbook/{{.cluster}}"
					target_revision = "HEAD"
				}
	  
				destination {
					server    = "{{.url}}"
					namespace = "guestbook"
				}
			}
		}
	}
}`
}

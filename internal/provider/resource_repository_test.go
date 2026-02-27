package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/assert"
)

func TestAccArgoCDRepository_Simple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDRepositorySimple(),
				Check: resource.TestCheckResourceAttr(
					"argocd_repository.simple",
					"connection_state_status",
					"Successful",
				),
			},
			{
				ResourceName:      "argocd_repository.simple",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccArgoCDRepositoryPublicUsageInApplication(acctest.RandString(10)),
				Check: resource.TestCheckResourceAttrSet(
					"argocd_application.public",
					"metadata.0.uid",
				),
			},
		},
	})
}

func TestAccArgoCDRepository_UseAzureWorkloadIdentity(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccArgoCDRepositoryUseAzureWorkloadIdentity(),
				ExpectError: regexp.MustCompile("failed to acquire a token"),
			},
		},
	})
}

func testAccArgoCDRepositoryUseAzureWorkloadIdentity() string {
	return `
resource "argocd_repository" "azurewi" {
  repo                        = "https://github.com/argoproj-labs/terraform-provider-argocd"
  use_azure_workload_identity = true
}
`
}

func TestAccArgoCDRepository_Helm(t *testing.T) {
	projectName := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDRepositoryHelm(),
				Check: resource.TestCheckResourceAttr(
					"argocd_repository.helm",
					"connection_state_status",
					"Successful",
				),
			},
			{
				ResourceName:      "argocd_repository.helm",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccArgoCDRepositoryHelmProjectScoped(projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository.helm",
						"connection_state_status",
						"Successful",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.helm",
						"project",
						projectName,
					),
				),
			},
		},
	})
}

func TestAccArgoCDRepository_PrivateSSH(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDRepositoryPrivateGitSSH("git@private-git-repository.argocd.svc.cluster.local:~/project-1.git"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository.private_ssh",
						"connection_state_status",
						"Successful",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.private_ssh",
						"inherited_creds",
						"false",
					),
				),
			},
			{
				ResourceName:            "argocd_repository.private_ssh",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ssh_private_key"},
			},
			{
				Config: testAccArgoCDRepositoryMultiplePrivateGitSSH(10),
				Check: testCheckMultipleResourceAttr(
					"argocd_repository.private_ssh",
					"connection_state_status",
					"Successful",
					10,
				),
			},
		},
	})
}

func TestAccArgoCDRepository_GitHubApp(t *testing.T) {
	sshPrivateKey, err := generateSSHPrivateKey()
	assert.NoError(t, err)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDRepositoryGitHubApp(
					"https://private-git-repository.argocd.svc.cluster.local/project-1.git",
					"123456",
					"987654321",
					"https://ghe.example.com/api/v3",
					sshPrivateKey,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository.githubapp",
						"githubapp_id",
						"123456",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.githubapp",
						"githubapp_installation_id",
						"987654321",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.githubapp",
						"githubapp_enterprise_base_url",
						"https://ghe.example.com/api/v3",
					),
				),
			},
		},
	})
}

// TestAccArgoCDRepository_GitHubAppConsistency tests the fix for issue #697
// This test verifies that GitHub App authentication fields remain consistent
// across multiple applies without configuration changes
func TestAccArgoCDRepository_GitHubAppConsistency(t *testing.T) {
	sshPrivateKey, err := generateSSHPrivateKey()
	assert.NoError(t, err)

	config := testAccArgoCDRepositoryGitHubApp(
		"https://github.com/MyCompany/MyRepo.git",
		"12345",
		"987654",
		"",
		sshPrivateKey,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository.githubapp",
						"githubapp_id",
						"12345",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.githubapp",
						"githubapp_installation_id",
						"987654",
					),
				),
			},
			{
				// Apply the same configuration again to test for consistency
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository.githubapp",
						"githubapp_id",
						"12345",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.githubapp",
						"githubapp_installation_id",
						"987654",
					),
				),
			},
			{
				// Third apply to ensure fields don't become null over time
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository.githubapp",
						"githubapp_id",
						"12345",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.githubapp",
						"githubapp_installation_id",
						"987654",
					),
				),
			},
		},
	})
}

// TestAccArgoCDRepository_BearerTokenConsistency tests consistency of bearer token field
// Note: This test uses a Helm repository which doesn't require authentication but allows token auth
func TestAccArgoCDRepository_BearerTokenConsistency(t *testing.T) {
	config := `
resource "argocd_repository" "bearer_token" {
  repo     = "https://helm.nginx.com/stable"
  type     = "helm"
	bearer_token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.KMUFsIDTnFmyG3nMiGM6H9FNFUROf3wh7SmqJp-QV30"
}
`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository.bearer_token",
						"bearer_token",
						"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.KMUFsIDTnFmyG3nMiGM6H9FNFUROf3wh7SmqJp-QV30",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.bearer_token",
						"connection_state_status",
						"Successful",
					),
				),
			},
			{
				// Apply the same configuration again to test for consistency
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository.bearer_token",
						"bearer_token",
						"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.KMUFsIDTnFmyG3nMiGM6H9FNFUROf3wh7SmqJp-QV30",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.bearer_token",
						"connection_state_status",
						"Successful",
					),
				),
			},
		},
	})
}

// TestAccArgoCDRepository_UsernamePasswordConsistency tests consistency of username/password fields
// Note: This test uses a Helm repository which doesn't require authentication but allows username/password fields
func TestAccArgoCDRepository_UsernamePasswordConsistency(t *testing.T) {
	config := `
resource "argocd_repository" "username_password" {
  repo     = "https://helm.nginx.com/stable"
  type     = "helm"
  username = "testuser"
  password = "testpass"
}
`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository.username_password",
						"username",
						"testuser",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.username_password",
						"password",
						"testpass",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.username_password",
						"connection_state_status",
						"Successful",
					),
				),
			},
			{
				// Apply the same configuration again to test for consistency
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository.username_password",
						"username",
						"testuser",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.username_password",
						"password",
						"testpass",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.username_password",
						"connection_state_status",
						"Successful",
					),
				),
			},
		},
	})
}

// TestAccArgoCDRepository_TLSCertificateConsistency tests consistency of TLS certificate fields
func TestAccArgoCDRepository_TLSCertificateConsistency(t *testing.T) {
	certData := `-----BEGIN CERTIFICATE-----
MIICljCCAX4CCQCKXiP0ZxJxHDANBgkqhkiG9w0BAQsFADANMQswCQYDVQQGEwJV
UzAeFw0yNDAxMDEwMDAwMDBaFw0yNTAxMDEwMDAwMDBaMA0xCzAJBgNVBAYTAlVT
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA3qZ8x/example...
-----END CERTIFICATE-----`

	certKey := `-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDepnzH/example...
-----END PRIVATE KEY-----`

	config := fmt.Sprintf(`
resource "argocd_repository" "tls_cert" {
  repo                 = "https://github.com/kubernetes-sigs/kustomize"
  tls_client_cert_data = %q
  tls_client_cert_key  = %q
}
`, certData, certKey)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository.tls_cert",
						"tls_client_cert_data",
						certData,
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.tls_cert",
						"tls_client_cert_key",
						certKey,
					),
				),
			},
			{
				// Apply the same configuration again to test for consistency
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository.tls_cert",
						"tls_client_cert_data",
						certData,
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.tls_cert",
						"tls_client_cert_key",
						certKey,
					),
				),
			},
		},
	})
}

// TestAccArgoCDRepository_OptionalFieldsConsistency tests consistency of optional fields
func TestAccArgoCDRepository_OptionalFieldsConsistency(t *testing.T) {
	projectName := acctest.RandString(10)

	config := fmt.Sprintf(`
resource "argocd_project" "test" {
  metadata {
    name      = "%[1]s"
    namespace = "argocd"
  }
  spec {
    description  = "test project"
    source_repos = ["*"]
    destination {
      name      = "in-cluster"
      namespace = "default"
    }
  }
}

resource "argocd_repository" "optional_fields" {
  repo    = "https://helm.nginx.com/stable/"
  name    = "nginx-stable-test"
  type    = "helm"
  project = argocd_project.test.metadata[0].name
}
`, projectName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository.optional_fields",
						"name",
						"nginx-stable-test",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.optional_fields",
						"type",
						"helm",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.optional_fields",
						"project",
						projectName,
					),
				),
			},
			{
				// Apply the same configuration again to test for consistency
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository.optional_fields",
						"name",
						"nginx-stable-test",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.optional_fields",
						"type",
						"helm",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.optional_fields",
						"project",
						projectName,
					),
				),
			},
		},
	})
}

// TestAccArgoCDRepository_BooleanFieldsConsistency tests consistency of boolean fields
func TestAccArgoCDRepository_BooleanFieldsConsistency(t *testing.T) {
	config := `
resource "argocd_repository" "boolean_fields" {
  repo       = "https://github.com/kubernetes-sigs/kustomize"
  enable_lfs = true
  enable_oci = true
  insecure   = true
}
`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository.boolean_fields",
						"enable_lfs",
						"true",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.boolean_fields",
						"enable_oci",
						"true",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.boolean_fields",
						"insecure",
						"true",
					),
				),
			},
			{
				// Apply the same configuration again to test for consistency
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository.boolean_fields",
						"enable_lfs",
						"true",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.boolean_fields",
						"enable_oci",
						"true",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.boolean_fields",
						"insecure",
						"true",
					),
				),
			},
		},
	})
}

func TestAccArgoCDRepository_ProviderUpgradeStateMigration(t *testing.T) {
	config := `
resource "argocd_repository" "private" {
  count = 1
  repo  = "https://github.com/kubernetes-sigs/kustomize"
  name  = "gitlab-private"
  type  = "git"
}
`
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"argocd": {
						VersionConstraint: "7.8.0",
						Source:            "argoproj-labs/argocd",
					},
				},
				Config: config,
			},
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   config,
			},
			{
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

// TestAccArgoCDRepository_EmptyStringFieldsConsistency tests handling of empty string fields
func TestAccArgoCDRepository_EmptyStringFieldsConsistency(t *testing.T) {
	config := `
resource "argocd_repository" "empty_strings" {
  repo                          = "https://github.com/kubernetes-sigs/kustomize"
  name                          = ""
  githubapp_enterprise_base_url = ""
}
`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository.empty_strings",
						"name",
						"",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.empty_strings",
						"githubapp_enterprise_base_url",
						"",
					),
				),
			},
			{
				// Apply the same configuration again to test for consistency
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository.empty_strings",
						"name",
						"",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.empty_strings",
						"githubapp_enterprise_base_url",
						"",
					),
				),
			},
		},
	})
}

func testAccArgoCDRepositorySimple() string {
	return `
resource "argocd_repository" "simple" {
  repo = "https://github.com/kubernetes-sigs/kustomize"
}
`
}

func testAccArgoCDRepositoryHelm() string {
	return `
resource "argocd_repository" "helm" {
  repo = "https://helm.nginx.com/stable"
  name = "nginx-stable"
  type = "helm"
}
`
}

func testAccArgoCDRepositoryHelmProjectScoped(project string) string {
	return fmt.Sprintf(`
resource "argocd_project" "simple" {
  metadata {
    name      = "%[1]s"
    namespace = "argocd"
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

resource "argocd_repository" "helm" {
  repo = "https://helm.nginx.com/stable/"
  name = "nginx-stable-scoped"
  type = "helm"
  project = "%[1]s"
}
`, project)
}

func testAccArgoCDRepositoryPublicUsageInApplication(name string) string {
	return testAccArgoCDRepositorySimple() + fmt.Sprintf(`
resource "argocd_application" "public" {
  metadata {
    name      = "%s"
    namespace = "argocd"
  }
  spec {
    source {
      repo_url        = argocd_repository.simple.repo
      path            = "examples/helloWorld"
      target_revision = "release-kustomize-v3.7"
    }
    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
  }
}
`, name)
}

func testAccArgoCDRepositoryPrivateGitSSH(repoUrl string) string {
	return fmt.Sprintf(`
resource "argocd_repository" "private_ssh" {
  repo            = "%s"
  type            = "git"
  insecure        = true
  ssh_private_key = "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW\nQyNTUxOQAAACCGe6Vx0gbKqKCI0wIplfgK5JBjCDO3bhtU3sZfLoeUZgAAAJB9cNEifXDR\nIgAAAAtzc2gtZWQyNTUxOQAAACCGe6Vx0gbKqKCI0wIplfgK5JBjCDO3bhtU3sZfLoeUZg\nAAAEAJeUrObjoTbGO1Sq4TXHl/j4RJ5aKMC1OemWuHmLK7XYZ7pXHSBsqooIjTAimV+Ark\nkGMIM7duG1Texl8uh5RmAAAAC3Rlc3RAYXJnb2NkAQI=\n-----END OPENSSH PRIVATE KEY-----"
}
`, repoUrl)
}

func testAccArgoCDRepositoryMultiplePrivateGitSSH(repoCount int) string {
	return fmt.Sprintf(`
resource "argocd_repository" "private_ssh" {
  count           = %d
  repo            = format("git@private-git-repository.argocd.svc.cluster.local:~/project-%%d.git", count.index+1)
  type            = "git"
  insecure        = true
  ssh_private_key = "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW\nQyNTUxOQAAACCGe6Vx0gbKqKCI0wIplfgK5JBjCDO3bhtU3sZfLoeUZgAAAJB9cNEifXDR\nIgAAAAtzc2gtZWQyNTUxOQAAACCGe6Vx0gbKqKCI0wIplfgK5JBjCDO3bhtU3sZfLoeUZg\nAAAEAJeUrObjoTbGO1Sq4TXHl/j4RJ5aKMC1OemWuHmLK7XYZ7pXHSBsqooIjTAimV+Ark\nkGMIM7duG1Texl8uh5RmAAAAC3Rlc3RAYXJnb2NkAQI=\n-----END OPENSSH PRIVATE KEY-----"
}
`, repoCount)
}

func testAccArgoCDRepositoryGitHubApp(repoUrl, id, installID, baseURL, appKey string) string {
	return fmt.Sprintf(`
resource "argocd_repository" "githubapp" {
  project                       = "default"
  repo             				= "%s"
  githubapp_id    				= "%s"
  githubapp_installation_id 	= "%s"
  githubapp_enterprise_base_url = "%s"
  githubapp_private_key 		= <<EOT
%s
EOT
}
`, repoUrl, id, installID, baseURL, appKey)
}

// TestAccArgoCDRepository_MultiProject tests the fix for issue #719
// This test verifies that the same repository can be added to multiple projects
// and that Terraform correctly identifies them as separate resources
func TestAccArgoCDRepository_MultiProject(t *testing.T) {
	projectA := acctest.RandString(10)
	projectB := acctest.RandString(10)
	repoURL := "https://helm.nginx.com/stable"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create the repository in project A
				Config: testAccArgoCDRepositoryMultiProjectStepOne(projectA, repoURL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository.helm_project_a",
						"project",
						projectA,
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.helm_project_a",
						"repo",
						repoURL,
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.helm_project_a",
						"connection_state_status",
						"Successful",
					),
				),
			},
			{
				// Create the same repository in project B
				// This should create a separate resource, not try to update project A's repository
				Config: testAccArgoCDRepositoryMultiProjectStepTwo(projectA, projectB, repoURL),
				Check: resource.ComposeTestCheckFunc(
					// Verify project A repository still exists with correct project
					resource.TestCheckResourceAttr(
						"argocd_repository.helm_project_a",
						"project",
						projectA,
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.helm_project_a",
						"repo",
						repoURL,
					),
					// Verify project B repository exists with correct project
					resource.TestCheckResourceAttr(
						"argocd_repository.helm_project_b",
						"project",
						projectB,
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.helm_project_b",
						"repo",
						repoURL,
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.helm_project_b",
						"connection_state_status",
						"Successful",
					),
				),
			},
		},
	})
}

func testAccArgoCDRepositoryMultiProjectStepOne(projectA, repoURL string) string {
	return fmt.Sprintf(`
resource "argocd_project" "project_a" {
  metadata {
    name      = "%[1]s"
    namespace = "argocd"
  }
  spec {
    description  = "Project A"
    source_repos = ["*"]
    destination {
      name      = "in-cluster"
      namespace = "default"
    }
  }
}

resource "argocd_repository" "helm_project_a" {
  repo    = "%[2]s"
  name    = "nginx-stable-project-a"
  type    = "helm"
  project = argocd_project.project_a.metadata[0].name
}
`, projectA, repoURL)
}

func testAccArgoCDRepositoryMultiProjectStepTwo(projectA, projectB, repoURL string) string {
	return fmt.Sprintf(`
resource "argocd_project" "project_a" {
  metadata {
    name      = "%[1]s"
    namespace = "argocd"
  }
  spec {
    description  = "Project A"
    source_repos = ["*"]
    destination {
      name      = "in-cluster"
      namespace = "default"
    }
  }
}

resource "argocd_repository" "helm_project_a" {
  repo    = "%[3]s"
  name    = "nginx-stable-project-a"
  type    = "helm"
  project = argocd_project.project_a.metadata[0].name
}

resource "argocd_project" "project_b" {
  metadata {
    name      = "%[2]s"
    namespace = "argocd"
  }
  spec {
    description  = "Project B"
    source_repos = ["*"]
    destination {
      name      = "in-cluster"
      namespace = "default"
    }
  }
}

resource "argocd_repository" "helm_project_b" {
  repo    = "%[3]s"
  name    = "nginx-stable-project-b"
  type    = "helm"
  project = argocd_project.project_b.metadata[0].name
}
`, projectA, projectB, repoURL)
}

// TestAccArgoCDRepository_ProjectChange tests that changing the project field requires replacement
func TestAccArgoCDRepository_ProjectChange(t *testing.T) {
	projectA := acctest.RandString(10)
	projectB := acctest.RandString(10)
	repoURL := "https://helm.nginx.com/stable"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create repository in project A
				Config: testAccArgoCDRepositoryProjectChange(projectA, projectB, repoURL, projectA),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository.changing_project",
						"project",
						projectA,
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.changing_project",
						"repo",
						repoURL,
					),
				),
			},
			{
				// Change to project B - should require replacement
				Config: testAccArgoCDRepositoryProjectChange(projectA, projectB, repoURL, projectB),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository.changing_project",
						"project",
						projectB,
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.changing_project",
						"repo",
						repoURL,
					),
				),
			},
		},
	})
}

// TestAccArgoCDRepository_ProjectToGlobal tests changing from project-scoped to global
func TestAccArgoCDRepository_ProjectToGlobal(t *testing.T) {
	projectName := acctest.RandString(10)
	repoURL := "https://helm.nginx.com/stable"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create project-scoped repository
				Config: testAccArgoCDRepositoryProjectToGlobalStep1(projectName, repoURL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository.project_to_global",
						"project",
						projectName,
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.project_to_global",
						"repo",
						repoURL,
					),
				),
			},
			{
				// Change to global (remove project) - should require replacement
				Config: testAccArgoCDRepositoryProjectToGlobalStep2(repoURL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(
						"argocd_repository.project_to_global",
						"project",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.project_to_global",
						"repo",
						repoURL,
					),
				),
			},
		},
	})
}

// TestAccArgoCDRepository_GlobalToProject tests changing from global to project-scoped
func TestAccArgoCDRepository_GlobalToProject(t *testing.T) {
	projectName := acctest.RandString(10)
	repoURL := "https://helm.nginx.com/stable"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create global repository
				Config: testAccArgoCDRepositoryGlobalToProjectStep1(repoURL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(
						"argocd_repository.global_to_project",
						"project",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.global_to_project",
						"repo",
						repoURL,
					),
				),
			},
			{
				// Change to project-scoped - should require replacement
				Config: testAccArgoCDRepositoryGlobalToProjectStep2(projectName, repoURL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository.global_to_project",
						"project",
						projectName,
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.global_to_project",
						"repo",
						repoURL,
					),
				),
			},
		},
	})
}

func testAccArgoCDRepositoryProjectChange(projectA, projectB, repoURL, currentProject string) string {
	return fmt.Sprintf(`
resource "argocd_project" "project_a" {
  metadata {
    name      = "%[1]s"
    namespace = "argocd"
  }
  spec {
    description  = "Project A"
    source_repos = ["*"]
    destination {
      name      = "in-cluster"
      namespace = "default"
    }
  }
}

resource "argocd_project" "project_b" {
  metadata {
    name      = "%[2]s"
    namespace = "argocd"
  }
  spec {
    description  = "Project B"
    source_repos = ["*"]
    destination {
      name      = "in-cluster"
      namespace = "default"
    }
  }
}

resource "argocd_repository" "changing_project" {
  repo    = "%[3]s"
  name    = "nginx-stable-changing"
  type    = "helm"
  project = "%[4]s"
}
`, projectA, projectB, repoURL, currentProject)
}

func testAccArgoCDRepositoryProjectToGlobalStep1(projectName, repoURL string) string {
	return fmt.Sprintf(`
resource "argocd_project" "test" {
  metadata {
    name      = "%[1]s"
    namespace = "argocd"
  }
  spec {
    description  = "Test Project"
    source_repos = ["*"]
    destination {
      name      = "in-cluster"
      namespace = "default"
    }
  }
}

resource "argocd_repository" "project_to_global" {
  repo    = "%[2]s"
  name    = "nginx-stable-p2g"
  type    = "helm"
  project = argocd_project.test.metadata[0].name
}
`, projectName, repoURL)
}

func testAccArgoCDRepositoryProjectToGlobalStep2(repoURL string) string {
	return fmt.Sprintf(`
resource "argocd_repository" "project_to_global" {
  repo = "%[1]s"
  name = "nginx-stable-p2g"
  type = "helm"
}
`, repoURL)
}

func testAccArgoCDRepositoryGlobalToProjectStep1(repoURL string) string {
	return fmt.Sprintf(`
resource "argocd_repository" "global_to_project" {
  repo = "%[1]s"
  name = "nginx-stable-g2p"
  type = "helm"
}
`, repoURL)
}

func testAccArgoCDRepositoryGlobalToProjectStep2(projectName, repoURL string) string {
	return fmt.Sprintf(`
resource "argocd_project" "test" {
  metadata {
    name      = "%[1]s"
    namespace = "argocd"
  }
  spec {
    description  = "Test Project"
    source_repos = ["*"]
    destination {
      name      = "in-cluster"
      namespace = "default"
    }
  }
}

resource "argocd_repository" "global_to_project" {
  repo    = "%[2]s"
  name    = "nginx-stable-g2p"
  type    = "helm"
  project = argocd_project.test.metadata[0].name
}
`, projectName, repoURL)
}

// TestAccArgoCDRepository_ProxyConnectivityError verifies that proxy configuration
// is correctly passed to ArgoCD by expecting a connection failure when using an invalid proxy.
func TestAccArgoCDRepository_ProxyConnectivityError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "argocd_repository" "proxy_fail" {
  repo     = "https://helm.nginx.com/stable"
  name     = "nginx-stable-proxy-fail"
  type     = "helm"
  proxy    = "http://proxy.example.com:8080"
}
`,
				ExpectError: regexp.MustCompile("proxyconnect tcp|no such host|context deadline exceeded|Unable to connect to repository"),
			},
		},
	})
}

// TestAccArgoCDRepository_OCI verifies that OCI repository type is correctly handled.
// We use the public ArgoCD OCI registry which allows anonymous access.
func TestAccArgoCDRepository_OCI(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "argocd_repository" "oci_test" {
  repo = "oci://ghcr.io/argoproj/argo-helm/argo-cd"
  name = "argocd-oci"
  type = "oci"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("argocd_repository.oci_test", "repo", "oci://ghcr.io/argoproj/argo-helm/argo-cd"),
					resource.TestCheckResourceAttr("argocd_repository.oci_test", "type", "oci"),
				),
			},
		},
	})
}

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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

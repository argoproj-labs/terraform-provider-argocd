package argocd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAccArgoCDRepository_Simple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
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
  repo = "https://helm.nginx.com/stable"
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

func testCheckMultipleResourceAttr(name, key, value string, count int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for i := 0; i < count; i++ {
			ms := s.RootModule()
			_name := fmt.Sprintf("%s.%d", name, i)

			rs, ok := ms.Resources[_name]
			if !ok {
				return fmt.Errorf("not found: %s in %s", _name, ms.Path)
			}

			is := rs.Primary
			if is == nil {
				return fmt.Errorf("no primary instance: %s in %s", _name, ms.Path)
			}

			if val, ok := is.Attributes[key]; !ok || val != value {
				return fmt.Errorf("%s: Attribute '%s' expected to be set and have value '%s': %s", _name, key, value, val)
			}
		}

		return nil
	}
}

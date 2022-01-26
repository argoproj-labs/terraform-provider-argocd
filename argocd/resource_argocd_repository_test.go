package argocd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccArgoCDRepository(t *testing.T) {
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
				Config: testAccArgoCDRepositoryHelm(),
				Check: resource.TestCheckResourceAttr(
					"argocd_repository.helm",
					"connection_state_status",
					"Successful",
				),
			},
			{
				Config: testAccArgoCDRepositoryPublicUsageInApplication(acctest.RandString(10)),
				Check: resource.TestCheckResourceAttrSet(
					"argocd_application.public",
					"metadata.0.uid",
				),
			},
			{
				Config: testAccArgoCDRepositoryPrivateGitSSH("git@private-git-repository.argocd.svc.cluster.local:~/project-1.git"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository.private",
						"connection_state_status",
						"Successful",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.private",
						"inherited_creds",
						"false",
					),
				),
			},
			{
				Config: testAccArgoCDRepositoryMultiplePrivateGitSSH(10),
				Check: testCheckMultipleResourceAttr(
					"argocd_repository.private",
					"connection_state_status",
					"Successful",
					10,
				),
			},
		},
	})
}

func testAccArgoCDRepositorySimple() string {
	return fmt.Sprintf(`
resource "argocd_repository" "simple" {
  repo = "https://github.com/kubernetes-sigs/kustomize"
}
`)
}

func testAccArgoCDRepositoryHelm() string {
	return fmt.Sprintf(`
resource "argocd_repository" "helm" {
  repo = "https://helm.nginx.com/stable"
  name = "nginx-stable"
  type = "helm"
}
`)
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
resource "argocd_repository" "private" {
  repo            = "%s"
  type            = "git"
  insecure        = true
  ssh_private_key = "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW\nQyNTUxOQAAACCGe6Vx0gbKqKCI0wIplfgK5JBjCDO3bhtU3sZfLoeUZgAAAJB9cNEifXDR\nIgAAAAtzc2gtZWQyNTUxOQAAACCGe6Vx0gbKqKCI0wIplfgK5JBjCDO3bhtU3sZfLoeUZg\nAAAEAJeUrObjoTbGO1Sq4TXHl/j4RJ5aKMC1OemWuHmLK7XYZ7pXHSBsqooIjTAimV+Ark\nkGMIM7duG1Texl8uh5RmAAAAC3Rlc3RAYXJnb2NkAQI=\n-----END OPENSSH PRIVATE KEY-----"
}
`, repoUrl)
}

func testAccArgoCDRepositoryMultiplePrivateGitSSH(repoCount int) string {
	return fmt.Sprintf(`
resource "argocd_repository" "private" {
  count           = %d
  repo            = format("git@private-git-repository.argocd.svc.cluster.local:~/project-%%d.git", count.index+1)
  type            = "git"
  insecure        = true
  ssh_private_key = "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW\nQyNTUxOQAAACCGe6Vx0gbKqKCI0wIplfgK5JBjCDO3bhtU3sZfLoeUZgAAAJB9cNEifXDR\nIgAAAAtzc2gtZWQyNTUxOQAAACCGe6Vx0gbKqKCI0wIplfgK5JBjCDO3bhtU3sZfLoeUZg\nAAAEAJeUrObjoTbGO1Sq4TXHl/j4RJ5aKMC1OemWuHmLK7XYZ7pXHSBsqooIjTAimV+Ark\nkGMIM7duG1Texl8uh5RmAAAAC3Rlc3RAYXJnb2NkAQI=\n-----END OPENSSH PRIVATE KEY-----"
}
`, repoCount)
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

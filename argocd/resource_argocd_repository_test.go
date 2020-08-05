package argocd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"testing"
)

func TestAccArgoCDRepository(t *testing.T) {
	repoUrl := "git@private-git-repository.argocd.svc.cluster.local:project.git"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			//{
			//	Config: testAccArgoCDRepositorySimple(),
			//	Check: resource.ComposeTestCheckFunc(
			//		resource.TestCheckResourceAttr(
			//			"argocd_repository.simple",
			//			"connection_state_status",
			//			"Successful",
			//		),
			//	),
			//},
			//{
			//	Config: testAccArgoCDRepositoryHelm(),
			//	Check: resource.ComposeTestCheckFunc(
			//		resource.TestCheckResourceAttr(
			//			"argocd_repository.helm",
			//			"connection_state_status",
			//			"Successful",
			//		),
			//	),
			//},
			//{
			//	Config: testAccArgoCDRepositoryPublicUsageInApplication(acctest.RandString(10)),
			//	Check: resource.ComposeTestCheckFunc(
			//		resource.TestCheckResourceAttrSet(
			//			"argocd_application.public",
			//			"metadata.0.uid",
			//		),
			//	),
			//},
			{
				Config: testAccArgoCDRepositoryPrivateGitSSH(repoUrl),
				//ExpectNonEmptyPlan: true,
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
      repo_url = argocd_repository.simple.repo
      path     = "examples/helloWorld"
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

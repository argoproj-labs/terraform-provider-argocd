package argocd

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"
)

func TestAccArgoCDRepositoryCredentials(t *testing.T) {
	sshPrivateKey, err := generateSSHPrivateKey()
	assert.NoError(t, err)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDRepositoryCredentialsSimple(
					"https://private-git-repository.argocd.svc.cluster.local/project-1.git",
					"git",
					sshPrivateKey,
				),
				Check: resource.TestCheckResourceAttr(
					"argocd_repository_credentials.simple",
					"username",
					"git",
				),
			},
			{
				Config: testAccArgoCDRepositoryCredentialsRepositoryCoexistence(),
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

func testAccArgoCDRepositoryCredentialsSimple(repoUrl, username, sshPrivateKey string) string {
	return fmt.Sprintf(`
resource "argocd_repository_credentials" "simple" {
  url             = "%s"
  username        = "%s"
  ssh_private_key = <<EOT
%s
EOT
}
`, repoUrl, username, sshPrivateKey)
}

func testAccArgoCDRepositoryCredentialsRepositoryCoexistence() string {
	return fmt.Sprintf(`
resource "argocd_repository" "private" {
  count      = 10
  repo       = format("git@private-git-repository.argocd.svc.cluster.local:~/project-%%d.git", count.index+1)
  insecure   = true
  depends_on = [argocd_repository_credentials.private]
}

resource "argocd_repository_credentials" "private" {
  url             = "git@private-git-repository.argocd.svc.cluster.local"
  username        = "git"
  ssh_private_key = "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW\nQyNTUxOQAAACCGe6Vx0gbKqKCI0wIplfgK5JBjCDO3bhtU3sZfLoeUZgAAAJB9cNEifXDR\nIgAAAAtzc2gtZWQyNTUxOQAAACCGe6Vx0gbKqKCI0wIplfgK5JBjCDO3bhtU3sZfLoeUZg\nAAAEAJeUrObjoTbGO1Sq4TXHl/j4RJ5aKMC1OemWuHmLK7XYZ7pXHSBsqooIjTAimV+Ark\nkGMIM7duG1Texl8uh5RmAAAAC3Rlc3RAYXJnb2NkAQI=\n-----END OPENSSH PRIVATE KEY-----"
}
`)
}

func generateSSHPrivateKey() (privateKey string, err error) {
	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return
	}
	err = pk.Validate()
	if err != nil {
		return
	}
	privDER := x509.MarshalPKCS1PrivateKey(pk)
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}
	return string(pem.EncodeToMemory(&privBlock)), nil
}

func mustGenerateSSHPrivateKey(t *testing.T) string {
	pk, err := generateSSHPrivateKey()
	assert.NoError(t, err)
	return pk
}

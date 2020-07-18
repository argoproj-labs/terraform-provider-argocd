package argocd

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAccArgoCDRepositoryCredentials(t *testing.T) {
	repoUrl := "https://private-git-repository.argocd.svc.clusterlocal/project.git"
	username := "git"
	sshPrivateKey, err := generateSSHPrivateKey()
	assert.NoError(t, err)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDRepositoryCredentialsSimple(repoUrl, username, sshPrivateKey),
				Check: resource.TestCheckResourceAttr(
					"argocd_repository_credentials.simple",
					"username",
					username,
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

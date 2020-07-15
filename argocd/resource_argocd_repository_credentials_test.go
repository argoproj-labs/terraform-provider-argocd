package argocd

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"testing"
)

func TestAccArgoCDRepositoryCredentials(t *testing.T) {
	repoUrl := fmt.Sprintf("https://git.local/%s/%s",
		acctest.RandString(10),
		acctest.RandString(10))
	username := fmt.Sprintf(acctest.RandString(10))
	sshPrivateKey, err := generateSSHPrivateKey()
	if err != nil {
		panic(err)
	}
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

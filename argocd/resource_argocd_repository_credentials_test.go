package argocd

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"testing"
)

func TestAccArgoCDRepositoryCredentials(t *testing.T) {
	sshPrivateKey, err := generateSSHPrivateKey()
	if err != nil {
		panic(err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDRepositoryCredentialsSimple(sshPrivateKey),
				Check: resource.TestCheckResourceAttr(
					"argocd_repository_credentials.simple",
					"ssh_private_key",
					sshPrivateKey,
				),
			},
		},
	})
}

func testAccArgoCDRepositoryCredentialsSimple(sshPrivateKey string) string {
	return fmt.Sprintf(`
resource "argocd_repository_credentials" "simple" {
  url             = "https://github.com/kubernetes-sigs/kustomize"
  ssh_private_key = "%s"
}
`, sshPrivateKey)
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

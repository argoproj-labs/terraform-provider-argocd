package argocd

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func TestAccArgoCDCluster(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDClusterTLSCertificate(t, acctest.RandString(10)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_cluster.tls",
						"info.0.connection_state.0.status",
						"Successful",
					),
					resource.TestCheckResourceAttr(
						"argocd_cluster.tls",
						"info.0.server_version",
						"1.20",
					),
					resource.TestCheckResourceAttr(
						"argocd_cluster.tls",
						"config.0.tls_client_config.0.insecure",
						"false",
					),
				),
			},
		},
	})
}

func testAccArgoCDClusterTLSCertificate(t *testing.T, clusterName string) string {
	rc, err := getInternalRestConfig()
	if err != nil {
		t.Error(err)
	}
	return fmt.Sprintf(`
resource "argocd_cluster" "tls" {
  server = "https://kubernetes.default.svc.cluster.local"
  name   = "%s"
  namespaces = ["bar", "baz"]
  config {
    tls_client_config {
      key_data    = <<EOT
%s
EOT
      cert_data   = <<EOT
%s
EOT
      ca_data     = <<EOT
%s
EOT
      server_name = "%s"
      insecure    = false
    }
  }
}
`, clusterName, rc.KeyData, rc.CertData, rc.CAData, rc.ServerName)
}

// getInternalRestConfig returns the internal Kubernetes cluster REST config.
func getInternalRestConfig() (*rest.Config, error) {
	rc := &rest.Config{}
	var kubeConfigFilePath string

	switch runtime.GOOS {
	case "windows":
		kubeConfigFilePath = fmt.Sprintf("%s\\.kube\\config", homedir.HomeDir())
	default:
		kubeConfigFilePath = fmt.Sprintf("%s/.kube/config", homedir.HomeDir())
	}
	cfg, err := clientcmd.LoadFromFile(kubeConfigFilePath)
	if err != nil {
		return nil, err
	}
	for key, cluster := range cfg.Clusters {
		if key == "k3d-argocd" {
			authInfo := cfg.AuthInfos["admin@k3d-argocd"]
			rc.Host = cluster.Server
			rc.ServerName = cluster.TLSServerName
			rc.TLSClientConfig.CAData = cluster.CertificateAuthorityData
			rc.TLSClientConfig.CertData = authInfo.ClientCertificateData
			rc.TLSClientConfig.KeyData = authInfo.ClientKeyData
			return rc, nil
		}
	}
	return nil, fmt.Errorf("could not find a k3d-argocd cluster from the current ~/.kube/config file")
}

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
				Config: testAccArgoCDClusterBearerToken(acctest.RandString(10)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_cluster.simple",
						"info.0.connection_state.0.status",
						"Successful",
					),
					resource.TestCheckResourceAttr(
						"argocd_cluster.simple",
						"shard",
						"1",
					),
					resource.TestCheckResourceAttr(
						"argocd_cluster.simple",
						"info.0.server_version",
						"1.19",
					),
					resource.TestCheckResourceAttr(
						"argocd_cluster.simple",
						"info.0.applications_count",
						"0",
					),
					resource.TestCheckResourceAttr(
						"argocd_cluster.simple",
						"config.0.tls_client_config.0.insecure",
						"true",
					),
				),
			},
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
						"1.19",
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

func testAccArgoCDClusterBearerToken(clusterName string) string {
	return fmt.Sprintf(`
resource "argocd_cluster" "simple" {
  server = "https://kubernetes.default.svc.cluster.local"
  name   = "%s"
  shard  = "1"
  namespaces = ["default", "foo"]
  config {
    # Uses Kind's bootstrap token whose ttl is 24 hours after cluster bootstrap.
    bearer_token = "abcdef.0123456789abcdef"
    tls_client_config {
      insecure = true
    }
  }
}
`, clusterName)
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
		if key == "kind-argocd" {
			authInfo := cfg.AuthInfos[key]
			rc.Host = cluster.Server
			rc.ServerName = cluster.TLSServerName
			rc.TLSClientConfig.CAData = cluster.CertificateAuthorityData
			rc.TLSClientConfig.CertData = authInfo.ClientCertificateData
			rc.TLSClientConfig.KeyData = authInfo.ClientKeyData
			return rc, nil
		}
	}
	return nil, fmt.Errorf("could not find a kind-argocd cluster from the current ~/.kube/config file")
}

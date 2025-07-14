package argocd

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/argoproj-labs/terraform-provider-argocd/internal/provider"
	"github.com/argoproj-labs/terraform-provider-argocd/internal/testhelpers"
	"github.com/argoproj/argo-cd/v3/pkg/apiclient/cluster"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
					resource.TestCheckResourceAttrSet(
						"argocd_cluster.simple",
						"info.0.server_version",
					),
					resource.TestCheckResourceAttr(
						"argocd_cluster.simple",
						"info.0.applications_count",
						"0",
					),
					resource.TestCheckResourceAttr(
						"argocd_cluster.simple",
						"config.0.tls_client_config.0.insecure",
						strconv.FormatBool(isInsecure()),
					),
				),
			},
			{
				ResourceName:            "argocd_cluster.simple",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"config.0.bearer_token", "info", "config.0.tls_client_config.0.key_data"},
			},
			{
				Config: testAccArgoCDClusterTLSCertificate(t, acctest.RandString(10)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_cluster.tls",
						"info.0.connection_state.0.status",
						"Successful",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_cluster.tls",
						"info.0.server_version",
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

func TestAccArgoCDCluster_projectScope(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDClusterProjectScope(acctest.RandString(10), "myproject1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_cluster.project_scope",
						"info.0.connection_state.0.status",
						"Successful",
					),
					resource.TestCheckResourceAttr(
						"argocd_cluster.project_scope",
						"config.0.tls_client_config.0.insecure",
						strconv.FormatBool(isInsecure()),
					),
					resource.TestCheckResourceAttr(
						"argocd_cluster.project_scope",
						"project",
						"myproject1",
					),
				),
			},
			{
				ResourceName:            "argocd_cluster.project_scope",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"config.0.bearer_token", "info", "config.0.tls_client_config.0.key_data"},
			},
		},
	})
}

func TestAccArgoCDCluster_optionalName(t *testing.T) {
	name := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDClusterMetadataNoName(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_cluster.cluster_metadata",
						"info.0.connection_state.0.status",
						"Successful",
					),
					resource.TestCheckResourceAttr(
						"argocd_cluster.cluster_metadata",
						"config.0.tls_client_config.0.insecure",
						strconv.FormatBool(isInsecure()),
					),
					resource.TestCheckResourceAttr(
						"argocd_cluster.cluster_metadata",
						"name",
						"https://kubernetes.default.svc.cluster.local",
					),
				),
			},
			{
				Config: testAccArgoCDClusterMetadata(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_cluster.cluster_metadata",
						"info.0.connection_state.0.status",
						"Successful",
					),
					resource.TestCheckResourceAttr(
						"argocd_cluster.cluster_metadata",
						"config.0.tls_client_config.0.insecure",
						strconv.FormatBool(isInsecure()),
					),
					resource.TestCheckResourceAttr(
						"argocd_cluster.cluster_metadata",
						"name",
						name,
					),
				),
			},
			{
				Config: testAccArgoCDClusterMetadataNoName(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_cluster.cluster_metadata",
						"info.0.connection_state.0.status",
						"Successful",
					),
					resource.TestCheckResourceAttr(
						"argocd_cluster.cluster_metadata",
						"config.0.tls_client_config.0.insecure",
						strconv.FormatBool(isInsecure()),
					),
					resource.TestCheckResourceAttr(
						"argocd_cluster.cluster_metadata",
						"name",
						"https://kubernetes.default.svc.cluster.local",
					),
				),
			},
		},
	})
}

func TestAccArgoCDCluster_metadata(t *testing.T) {
	clusterName := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDClusterMetadata(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(
						"argocd_cluster.cluster_metadata",
						"metadata.0",
					),
				),
			},
			{
				ResourceName:            "argocd_cluster.cluster_metadata",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"config.0.bearer_token", "info", "config.0.tls_client_config.0.key_data"},
			},
			{
				Config: testAccArgoCDClusterMetadata_addLabels(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_cluster.cluster_metadata",
						"metadata.0.labels.test",
						"label",
					),
					resource.TestCheckNoResourceAttr(
						"argocd_cluster.cluster_metadata",
						"metadata.0.annotations",
					),
				),
			},
			{
				ResourceName:            "argocd_cluster.cluster_metadata",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"config.0.bearer_token", "info", "config.0.tls_client_config.0.key_data"},
			},
			{
				Config: testAccArgoCDClusterMetadata_addAnnotations(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_cluster.cluster_metadata",
						"metadata.0.labels.test",
						"label",
					),
					resource.TestCheckResourceAttr(
						"argocd_cluster.cluster_metadata",
						"metadata.0.annotations.test",
						"annotation",
					),
				),
			},
			{
				ResourceName:            "argocd_cluster.cluster_metadata",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"config.0.bearer_token", "info", "config.0.tls_client_config.0.key_data"},
			},
			{
				Config: testAccArgoCDClusterMetadata_removeLabels(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(
						"argocd_cluster.cluster_metadata",
						"metadata.0.labels.test",
					),
					resource.TestCheckResourceAttr(
						"argocd_cluster.cluster_metadata",
						"metadata.0.annotations.test",
						"annotation",
					),
				),
			},
			{
				ResourceName:            "argocd_cluster.cluster_metadata",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"config.0.bearer_token", "info", "config.0.tls_client_config.0.key_data"},
			},
		},
	})
}

func TestAccArgoCDCluster_invalidSameServer(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccArgoCDClusterTwiceWithSameServer(),
				ExpectError: regexp.MustCompile("cluster with server address .* already exists"),
			},
			{
				Config:      testAccArgoCDClusterTwiceWithSameServerNoNames(),
				ExpectError: regexp.MustCompile("cluster with server address .* already exists"),
			},
			{
				Config:      testAccArgoCDClusterTwiceWithSameLogicalServer(),
				ExpectError: regexp.MustCompile("cluster with server address .* already exists"),
			},
		},
	})
}

func TestAccArgoCDCluster_outsideDeletion(t *testing.T) {
	clusterName := acctest.RandString(10)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDClusterMetadata(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_cluster.cluster_metadata",
						"info.0.connection_state.0.status",
						"Successful",
					),
					resource.TestCheckResourceAttr(
						"argocd_cluster.cluster_metadata",
						"config.0.tls_client_config.0.insecure",
						strconv.FormatBool(isInsecure()),
					),
					resource.TestCheckResourceAttr(
						"argocd_cluster.cluster_metadata",
						"name",
						clusterName,
					),
				),
			},
			{
				PreConfig: func() {
					// delete cluster and validate refresh generates a plan
					// (non-regression test for https://github.com/oboukili/terraform-provider-argocd/issues/266)
					si, err := getServerInterface()
					if err != nil {
						t.Error(fmt.Errorf("failed to get server interface: %s", err.Error()))
					}
					ctx, cancel := context.WithTimeout(t.Context(), 120*time.Second)
					defer cancel()
					_, err = si.ClusterClient.Delete(ctx, &cluster.ClusterQuery{Name: clusterName})
					if err != nil {
						t.Error(fmt.Errorf("failed to delete cluster '%s': %s", clusterName, err.Error()))
					}
				},
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccArgoCDClusterMetadata(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_cluster.cluster_metadata",
						"info.0.connection_state.0.status",
						"Successful",
					),
					resource.TestCheckResourceAttr(
						"argocd_cluster.cluster_metadata",
						"config.0.tls_client_config.0.insecure",
						strconv.FormatBool(isInsecure()),
					),
					resource.TestCheckResourceAttr(
						"argocd_cluster.cluster_metadata",
						"name",
						clusterName,
					),
				),
			},
		},
	})
}

func TestAccArgoCDCluster_namespacesErrorWhenEmpty(t *testing.T) {
	name := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccArgoCDClusterNamespacesContainsEmptyString(name),
				ExpectError: regexp.MustCompile("namespaces: must contain non-empty strings"),
			},
			{
				Config:      testAccArgoCDClusterNamespacesContainsEmptyString_MultipleItems(name),
				ExpectError: regexp.MustCompile("namespaces: must contain non-empty strings"),
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
%s
  }
}
`, clusterName, getConfig())
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

func testAccArgoCDClusterProjectScope(clusterName, projectName string) string {
	return fmt.Sprintf(`
resource "argocd_cluster" "project_scope" {
  server = "https://kubernetes.default.svc.cluster.local"
  name   = "%s"
  project = "%s"
  config {
%s
  }
}
`, clusterName, projectName, getConfig())
}

func testAccArgoCDClusterMetadata(clusterName string) string {
	return fmt.Sprintf(`
resource "argocd_cluster" "cluster_metadata" {
  server = "https://kubernetes.default.svc.cluster.local"
  name   = "%s"
  config {
%s
  }
}
`, clusterName, getConfig())
}

func testAccArgoCDClusterMetadataNoName() string {
	return fmt.Sprintf(`
resource "argocd_cluster" "cluster_metadata" {
  server = "https://kubernetes.default.svc.cluster.local"
  config {
%s
  }
}
`, getConfig())
}

func testAccArgoCDClusterTwiceWithSameServer() string {
	return fmt.Sprintf(`
resource "argocd_cluster" "cluster_one_same_server" {
  server = "https://kubernetes.default.svc.cluster.local"
  name   = "foo"
  config {
%s
  }
}
resource "argocd_cluster" "cluster_two_same_server" {
  server = "https://kubernetes.default.svc.cluster.local"
  name   = "bar"
  config {
%s
  }
}`, getConfig(), getConfig())
}

func testAccArgoCDClusterTwiceWithSameServerNoNames() string {
	return fmt.Sprintf(`
resource "argocd_cluster" "cluster_one_no_name" {
  server = "https://kubernetes.default.svc.cluster.local"
  config {
%s
  }
}
resource "argocd_cluster" "cluster_two_no_name" {
  server = "https://kubernetes.default.svc.cluster.local"
  config {
%s
  }
}
`, getConfig(), getConfig())
}

func testAccArgoCDClusterTwiceWithSameLogicalServer() string {
	return fmt.Sprintf(`
resource "argocd_cluster" "cluster_with_trailing_slash" {
  name = "server"
  server = "https://kubernetes.default.svc.cluster.local/"
  config {
%s
  }
}
resource "argocd_cluster" "cluster_with_no_trailing_slash" {
  name = "server"
  server = "https://kubernetes.default.svc.cluster.local"
  config {
%s
  }
}`, getConfig(), getConfig())
}

func testAccArgoCDClusterMetadata_addLabels(clusterName string) string {
	return fmt.Sprintf(`
resource "argocd_cluster" "cluster_metadata" {
  server = "https://kubernetes.default.svc.cluster.local"
  name   = "%s"
  metadata {
    labels = {
      test = "label"
    }
  }
  config {
%s
  }
}
`, clusterName, getConfig())
}

func testAccArgoCDClusterMetadata_addAnnotations(clusterName string) string {
	return fmt.Sprintf(`
resource "argocd_cluster" "cluster_metadata" {
  server = "https://kubernetes.default.svc.cluster.local"
  name   = "%s"
  metadata {
    labels = {
      test = "label"
    }
    annotations = {
      test = "annotation"
    }
  }
  config {
%s
  }
}
`, clusterName, getConfig())
}

func testAccArgoCDClusterMetadata_removeLabels(clusterName string) string {
	return fmt.Sprintf(`
resource "argocd_cluster" "cluster_metadata" {
  server = "https://kubernetes.default.svc.cluster.local"
  name   = "%s"
  metadata {
    annotations = {
      test = "annotation"
    }
  }
  config {
%s
  }
}
`, clusterName, getConfig())
}

func testAccArgoCDClusterNamespacesContainsEmptyString(clusterName string) string {
	return fmt.Sprintf(`
resource "argocd_cluster" "simple" {
  server = "https://kubernetes.default.svc.cluster.local"
  name   = "%s"
  shard  = "1"
  namespaces = [""]
  config {
%s
  }
}
`, clusterName, getConfig())
}

func testAccArgoCDClusterNamespacesContainsEmptyString_MultipleItems(clusterName string) string {
	return fmt.Sprintf(`
resource "argocd_cluster" "simple" {
  server = "https://kubernetes.default.svc.cluster.local"
  name   = "%s"
  shard  = "1"
  namespaces = ["default", ""]
  config {
%s
  }
}
`, clusterName, getConfig())
}

// getInternalRestConfig returns the internal Kubernetes cluster REST config.
func getInternalRestConfig() (*rest.Config, error) {
	if testhelpers.GlobalTestEnv != nil {
		return testhelpers.GlobalTestEnv.RESTConfig, nil
	}

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

	rc := &rest.Config{}

	for key, cluster := range cfg.Clusters {
		if key == "kind-argocd" {
			authInfo := cfg.AuthInfos[key]
			rc.Host = cluster.Server
			rc.ServerName = cluster.TLSServerName
			rc.CAData = cluster.CertificateAuthorityData
			rc.CertData = authInfo.ClientCertificateData
			rc.KeyData = authInfo.ClientKeyData

			return rc, nil
		}
	}

	return nil, fmt.Errorf("could not find a kind-argocd cluster from the current ~/.kube/config file")
}

// build & init ArgoCD server interface
func getServerInterface() (*provider.ServerInterface, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	insecure, err := strconv.ParseBool(os.Getenv("ARGOCD_INSECURE"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse 'ARGOCD_INSECURE' env var to bool: %s", err.Error())
	}

	si := provider.NewServerInterface(provider.ArgoCDProviderConfig{
		ServerAddr: types.StringValue(os.Getenv("ARGOCD_SERVER")),
		Insecure:   types.BoolValue(insecure),
		Username:   types.StringValue(os.Getenv("ARGOCD_AUTH_USERNAME")),
		Password:   types.StringValue(os.Getenv("ARGOCD_AUTH_PASSWORD")),
	})

	diag := si.InitClients(ctx)
	if diag.HasError() {
		return nil, fmt.Errorf("failed to init clients: %v", diag.Errors())
	}

	return si, nil
}

func getConfig() string {
	if testhelpers.GlobalTestEnv != nil {
		r := testhelpers.GlobalTestEnv.RESTConfig

		return fmt.Sprintf(`
    tls_client_config {
      insecure = false
      ca_data = <<CA_DATA
%sCA_DATA
      cert_data = <<CERT_DATA
%sCERT_DATA
      key_data = <<KEY_DATA
%sKEY_DATA
    }`, string(r.CAData), string(r.CertData), string(r.KeyData))
	}

	return `
    # Uses Kind's bootstrap token whose ttl is 24 hours after cluster bootstrap.
    bearer_token = "abcdef.0123456789abcdef"
    tls_client_config {
      insecure = true
    }
`
}

func isInsecure() bool {
	return testhelpers.GlobalTestEnv == nil
}

package provider

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/argoproj-labs/terraform-provider-argocd/internal/testhelpers"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccArgoCDRepositoryCertificatesSSH(t *testing.T) {
	serverName := acctest.RandomWithPrefix("mywebsite")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDRepositoryCertificatesSSH(
					serverName,
					"ecdsa-sha2-nistp256",
					// gitlab's
					"AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFSMqzJeV9rUzU4kWitGjeR4PWSa29SPqJ1fVkhtj3Hw9xjLVXVYrU9QlYWrOLXBpQ6KWjbjTDTdDkoohFzgbEY=",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("argocd_repository_certificate.simple", "ssh.0.server_name", serverName),
					resource.TestCheckResourceAttr("argocd_repository_certificate.simple", "ssh.0.cert_subtype", "ecdsa-sha2-nistp256"),
					resource.TestCheckResourceAttrSet("argocd_repository_certificate.simple", "ssh.0.cert_info"),
				),
			},
			// same, no diff
			{
				Config: testAccArgoCDRepositoryCertificatesSSH(
					serverName,
					"ecdsa-sha2-nistp256",
					// gitlab's
					"AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFSMqzJeV9rUzU4kWitGjeR4PWSa29SPqJ1fVkhtj3Hw9xjLVXVYrU9QlYWrOLXBpQ6KWjbjTDTdDkoohFzgbEY=",
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// change only the cert_data => same id => diff
			{
				Config: testAccArgoCDRepositoryCertificatesSSH(
					serverName,
					"ecdsa-sha2-nistp256",
					// github's
					"AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEmKSENjQEezOmxkZMy7opKgwFB9nkt5YRrYMjNuG5N87uRgg6CLrbo5wAdT/y6v0mKV0U2w0WZ2YB/++Tpockg=",
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			// change cert_subtype & cert_data => changes id => diff
			{
				Config: testAccArgoCDRepositoryCertificatesSSH(
					serverName,
					"ssh-rsa",
					// github's
					"AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("argocd_repository_certificate.simple", "ssh.0.server_name", serverName),
					resource.TestCheckResourceAttr("argocd_repository_certificate.simple", "ssh.0.cert_subtype", "ssh-rsa"),
				),
			},
		},
	})
}

func TestAccArgoCDRepositoryCertificatesHttps(t *testing.T) {
	serverName := acctest.RandomWithPrefix("github")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDRepositoryCertificateHttps(
					serverName,
					// github's
					"-----BEGIN CERTIFICATE-----\nMIIFajCCBPCgAwIBAgIQBRiaVOvox+kD4KsNklVF3jAKBggqhkjOPQQDAzBWMQsw\nCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMTAwLgYDVQQDEydEaWdp\nQ2VydCBUTFMgSHlicmlkIEVDQyBTSEEzODQgMjAyMCBDQTEwHhcNMjIwMzE1MDAw\nMDAwWhcNMjMwMzE1MjM1OTU5WjBmMQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2Fs\naWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZyYW5jaXNjbzEVMBMGA1UEChMMR2l0SHVi\nLCBJbmMuMRMwEQYDVQQDEwpnaXRodWIuY29tMFkwEwYHKoZIzj0CAQYIKoZIzj0D\nAQcDQgAESrCTcYUh7GI/y3TARsjnANwnSjJLitVRgwgRI1JlxZ1kdZQQn5ltP3v7\nKTtYuDdUeEu3PRx3fpDdu2cjMlyA0aOCA44wggOKMB8GA1UdIwQYMBaAFAq8CCkX\njKU5bXoOzjPHLrPt+8N6MB0GA1UdDgQWBBR4qnLGcWloFLVZsZ6LbitAh0I7HjAl\nBgNVHREEHjAcggpnaXRodWIuY29tgg53d3cuZ2l0aHViLmNvbTAOBgNVHQ8BAf8E\nBAMCB4AwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMIGbBgNVHR8EgZMw\ngZAwRqBEoEKGQGh0dHA6Ly9jcmwzLmRpZ2ljZXJ0LmNvbS9EaWdpQ2VydFRMU0h5\nYnJpZEVDQ1NIQTM4NDIwMjBDQTEtMS5jcmwwRqBEoEKGQGh0dHA6Ly9jcmw0LmRp\nZ2ljZXJ0LmNvbS9EaWdpQ2VydFRMU0h5YnJpZEVDQ1NIQTM4NDIwMjBDQTEtMS5j\ncmwwPgYDVR0gBDcwNTAzBgZngQwBAgIwKTAnBggrBgEFBQcCARYbaHR0cDovL3d3\ndy5kaWdpY2VydC5jb20vQ1BTMIGFBggrBgEFBQcBAQR5MHcwJAYIKwYBBQUHMAGG\nGGh0dHA6Ly9vY3NwLmRpZ2ljZXJ0LmNvbTBPBggrBgEFBQcwAoZDaHR0cDovL2Nh\nY2VydHMuZGlnaWNlcnQuY29tL0RpZ2lDZXJ0VExTSHlicmlkRUNDU0hBMzg0MjAy\nMENBMS0xLmNydDAJBgNVHRMEAjAAMIIBfwYKKwYBBAHWeQIEAgSCAW8EggFrAWkA\ndgCt9776fP8QyIudPZwePhhqtGcpXc+xDCTKhYY069yCigAAAX+Oi8SRAAAEAwBH\nMEUCIAR9cNnvYkZeKs9JElpeXwztYB2yLhtc8bB0rY2ke98nAiEAjiML8HZ7aeVE\nP/DkUltwIS4c73VVrG9JguoRrII7gWMAdwA1zxkbv7FsV78PrUxtQsu7ticgJlHq\nP+Eq76gDwzvWTAAAAX+Oi8R7AAAEAwBIMEYCIQDNckqvBhup7GpANMf0WPueytL8\nu/PBaIAObzNZeNMpOgIhAMjfEtE6AJ2fTjYCFh/BNVKk1mkTwBTavJlGmWomQyaB\nAHYAs3N3B+GEUPhjhtYFqdwRCUp5LbFnDAuH3PADDnk2pZoAAAF/jovErAAABAMA\nRzBFAiEA9Uj5Ed/XjQpj/MxQRQjzG0UFQLmgWlc73nnt3CJ7vskCICqHfBKlDz7R\nEHdV5Vk8bLMBW1Q6S7Ga2SbFuoVXs6zFMAoGCCqGSM49BAMDA2gAMGUCMCiVhqft\n7L/stBmv1XqSRNfE/jG/AqKIbmjGTocNbuQ7kt1Cs7kRg+b3b3C9Ipu5FQIxAM7c\ntGKrYDGt0pH8iF6rzbp9Q4HQXMZXkNxg+brjWxnaOVGTDNwNH7048+s/hT9bUQ==\n-----END CERTIFICATE-----",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("argocd_repository_certificate.simple", "https.0.server_name", serverName),
					resource.TestCheckResourceAttr("argocd_repository_certificate.simple", "https.0.cert_subtype", "ecdsa"),
					resource.TestCheckResourceAttrSet("argocd_repository_certificate.simple", "https.0.cert_info"),
				),
			},
			{
				Config: testAccArgoCDRepositoryCertificateHttps(
					serverName,
					// gitlab's
					"-----BEGIN CERTIFICATE-----\nMIIGBzCCBO+gAwIBAgIQXCLSMilzZJR9TSABzbgKzzANBgkqhkiG9w0BAQsFADCB\njzELMAkGA1UEBhMCR0IxGzAZBgNVBAgTEkdyZWF0ZXIgTWFuY2hlc3RlcjEQMA4G\nA1UEBxMHU2FsZm9yZDEYMBYGA1UEChMPU2VjdGlnbyBMaW1pdGVkMTcwNQYDVQQD\nEy5TZWN0aWdvIFJTQSBEb21haW4gVmFsaWRhdGlvbiBTZWN1cmUgU2VydmVyIENB\nMB4XDTIxMDQxMjAwMDAwMFoXDTIyMDUxMTIzNTk1OVowFTETMBEGA1UEAxMKZ2l0\nbGFiLmNvbTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBANXnhcvOl289\n8oMglaax6bDz988oNMpXZCH6sI7Fzx9G/isEPObN6cyP+fjFa0dvwRmOHnepk2eo\nbzcECdgdBLCa7E29p7lLF0NFFTuIb52ew58fK/209XJ3amvjJ/m5rPP00uHrT+9v\nky2jkQUQszuC9R4vK+tfs2S5z9w6qh3hwIJecChzWKce8hRZdiO9S7ix/6ZNiAgw\nY2h8AiG0VruPOJ6PbNXOFUTsajK0EP8AzJfNDIjvWHjUOawR352m4eKxXvXm9knd\nB/w1gY90jmAQ9JIiyOm+QlmHwO+qQUpWYOxt5Xnb0Pp/RRHEtxDgjygQWajAwsxG\nobx6sCf6+qcCAwEAAaOCAtYwggLSMB8GA1UdIwQYMBaAFI2MXsRUrYrhd+mb+ZsF\n4bgBjWHhMB0GA1UdDgQWBBTFjbuGoOUrgk9Dhr35DblkBZCj1jAOBgNVHQ8BAf8E\nBAMCBaAwDAYDVR0TAQH/BAIwADAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUH\nAwIwSQYDVR0gBEIwQDA0BgsrBgEEAbIxAQICBzAlMCMGCCsGAQUFBwIBFhdodHRw\nczovL3NlY3RpZ28uY29tL0NQUzAIBgZngQwBAgEwgYQGCCsGAQUFBwEBBHgwdjBP\nBggrBgEFBQcwAoZDaHR0cDovL2NydC5zZWN0aWdvLmNvbS9TZWN0aWdvUlNBRG9t\nYWluVmFsaWRhdGlvblNlY3VyZVNlcnZlckNBLmNydDAjBggrBgEFBQcwAYYXaHR0\ncDovL29jc3Auc2VjdGlnby5jb20wggEEBgorBgEEAdZ5AgQCBIH1BIHyAPAAdgBG\npVXrdfqRIDC1oolp9PN9ESxBdL79SbiFq/L8cP5tRwAAAXjDcW8TAAAEAwBHMEUC\nIQCxf+r8/dbHJDrh0YTAKSwdR8VUxAT6kHN5/HLuOvSsKgIgY2jAAf/tr59/f0JX\nKvHaN4qIv54gtj+KsNo7N0d4xcEAdgDfpV6raIJPH2yt7rhfTj5a6s2iEqRqXo47\nEsAgRFwqcwAAAXjDcW7VAAAEAwBHMEUCID0jtWvtpO1yypP7i7SeZZb3dQ6QdLlD\nlXpvWhjqrQfdAiEA0gp8tTUwOt2XN01OVTUrDgb4wV5VbFtx1SSYNFREQxwweQYD\nVR0RBHIwcIIKZ2l0bGFiLmNvbYIPYXV0aC5naXRsYWIuY29tghRjdXN0b21lcnMu\nZ2l0bGFiLmNvbYIaZW1haWwuY3VzdG9tZXJzLmdpdGxhYi5jb22CD2dwcmQuZ2l0\nbGFiLmNvbYIOd3d3LmdpdGxhYi5jb20wDQYJKoZIhvcNAQELBQADggEBAD7lgx6z\ncZI+uLtr7fYWOZDtPChNy7YjAXVtDbrQ61D1lESUIZwyDF9/xCDMqMSe+It2+j+t\nT0PHkbz6zbJdUMQhQxW0RLMZUthPg66YLqRJuvBU7VdWHxhqjfFb9UZvxOzTGgmN\nMuzmdThtlhRacNCTxGO/AJfcAt13RbKyR30UtqHb883qAH6isQvYFsQmijXcJXiT\ntRbcJ1Dm/dI+57BCTYLp2WfBdg0Axla5QsApQ+ER5GZoY1m6H3+OWpX77IdCgXF+\nHMtKCn08QLVBjhLr3IkeKgrYJTR1IDmzRwGUuUVvn1iO9+W10GV02SMngdN4nFp3\nwoE3CsYogf1SfQM=\n-----END CERTIFICATE-----",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("argocd_repository_certificate.simple", "https.0.server_name", serverName),
					resource.TestCheckResourceAttr("argocd_repository_certificate.simple", "https.0.cert_subtype", "rsa"),
					resource.TestCheckResourceAttrSet("argocd_repository_certificate.simple", "https.0.cert_info"),
				),
			},
		},
	})
}

func TestAccArgoCDRepositoryCertificatesHttps_Crash(t *testing.T) {
	serverName := acctest.RandomWithPrefix("github")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDRepositoryCertificateHttps(
					serverName,
					// github's
					"-----BEGIN CERTIFICATE-----\nMIIFajCCBPCgAwIBAgIQBRiaVOvox+kD4KsNklVF3jAKBggqhkjOPQQDAzBWMQsw\nCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMTAwLgYDVQQDEydEaWdp\nQ2VydCBUTFMgSHlicmlkIEVDQyBTSEEzODQgMjAyMCBDQTEwHhcNMjIwMzE1MDAw\nMDAwWhcNMjMwMzE1MjM1OTU5WjBmMQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2Fs\naWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZyYW5jaXNjbzEVMBMGA1UEChMMR2l0SHVi\nLCBJbmMuMRMwEQYDVQQDEwpnaXRodWIuY29tMFkwEwYHKoZIzj0CAQYIKoZIzj0D\nAQcDQgAESrCTcYUh7GI/y3TARsjnANwnSjJLitVRgwgRI1JlxZ1kdZQQn5ltP3v7\nKTtYuDdUeEu3PRx3fpDdu2cjMlyA0aOCA44wggOKMB8GA1UdIwQYMBaAFAq8CCkX\njKU5bXoOzjPHLrPt+8N6MB0GA1UdDgQWBBR4qnLGcWloFLVZsZ6LbitAh0I7HjAl\nBgNVHREEHjAcggpnaXRodWIuY29tgg53d3cuZ2l0aHViLmNvbTAOBgNVHQ8BAf8E\nBAMCB4AwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMIGbBgNVHR8EgZMw\ngZAwRqBEoEKGQGh0dHA6Ly9jcmwzLmRpZ2ljZXJ0LmNvbS9EaWdpQ2VydFRMU0h5\nYnJpZEVDQ1NIQTM4NDIwMjBDQTEtMS5jcmwwRqBEoEKGQGh0dHA6Ly9jcmw0LmRp\nZ2ljZXJ0LmNvbS9EaWdpQ2VydFRMU0h5YnJpZEVDQ1NIQTM4NDIwMjBDQTEtMS5j\ncmwwPgYDVR0gBDcwNTAzBgZngQwBAgIwKTAnBggrBgEFBQcCARYbaHR0cDovL3d3\ndy5kaWdpY2VydC5jb20vQ1BTMIGFBggrBgEFBQcBAQR5MHcwJAYIKwYBBQUHMAGG\nGGh0dHA6Ly9vY3NwLmRpZ2ljZXJ0LmNvbTBPBggrBgEFBQcwAoZDaHR0cDovL2Nh\nY2VydHMuZGlnaWNlcnQuY29tL0RpZ2lDZXJ0VExTSHlicmlkRUNDU0hBMzg0MjAy\nMENBMS0xLmNydDAJBgNVHRMEAjAAMIIBfwYKKwYBBAHWeQIEAgSCAW8EggFrAWkA\ndgCt9776fP8QyIudPZwePhhqtGcpXc+xDCTKhYY069yCigAAAX+Oi8SRAAAEAwBH\nMEUCIAR9cNnvYkZeKs9JElpeXwztYB2yLhtc8bB0rY2ke98nAiEAjiML8HZ7aeVE\nP/DkUltwIS4c73VVrG9JguoRrII7gWMAdwA1zxkbv7FsV78PrUxtQsu7ticgJlHq\nP+Eq76gDwzvWTAAAAX+Oi8R7AAAEAwBIMEYCIQDNckqvBhup7GpANMf0WPueytL8\nu/PBaIAObzNZeNMpOgIhAMjfEtE6AJ2fTjYCFh/BNVKk1mkTwBTavJlGmWomQyaB\nAHYAs3N3B+GEUPhjhtYFqdwRCUp5LbFnDAuH3PADDnk2pZoAAAF/jovErAAABAMA\nRzBFAiEA9Uj5Ed/XjQpj/MxQRQjzG0UFQLmgWlc73nnt3CJ7vskCICqHfBKlDz7R\nEHdV5Vk8bLMBW1Q6S7Ga2SbFuoVXs6zFMAoGCCqGSM49BAMDA2gAMGUCMCiVhqft\n7L/stBmv1XqSRNfE/jG/AqKIbmjGTocNbuQ7kt1Cs7kRg+b3b3C9Ipu5FQIxAM7c\ntGKrYDGt0pH8iF6rzbp9Q4HQXMZXkNxg+brjWxnaOVGTDNwNH7048+s/hT9bUQ==\n-----END CERTIFICATE-----",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("argocd_repository_certificate.simple", "https.0.server_name", serverName),
					resource.TestCheckResourceAttr("argocd_repository_certificate.simple", "https.0.cert_subtype", "ecdsa"),
					resource.TestCheckResourceAttrSet("argocd_repository_certificate.simple", "https.0.cert_info"),
				),
			},
			{
				Config: testAccArgoCDRepositoryCertificateHttps(
					serverName,
					// gitlab's
					"-----BEGIN CERTIFICATE-----\nMIIGBzCCBO+gAwIBAgIQXCLSMilzZJR9TSABzbgKzzANBgkqhkiG9w0BAQsFADCB\njzELMAkGA1UEBhMCR0IxGzAZBgNVBAgTEkdyZWF0ZXIgTWFuY2hlc3RlcjEQMA4G\nA1UEBxMHU2FsZm9yZDEYMBYGA1UEChMPU2VjdGlnbyBMaW1pdGVkMTcwNQYDVQQD\nEy5TZWN0aWdvIFJTQSBEb21haW4gVmFsaWRhdGlvbiBTZWN1cmUgU2VydmVyIENB\nMB4XDTIxMDQxMjAwMDAwMFoXDTIyMDUxMTIzNTk1OVowFTETMBEGA1UEAxMKZ2l0\nbGFiLmNvbTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBANXnhcvOl289\n8oMglaax6bDz988oNMpXZCH6sI7Fzx9G/isEPObN6cyP+fjFa0dvwRmOHnepk2eo\nbzcECdgdBLCa7E29p7lLF0NFFTuIb52ew58fK/209XJ3amvjJ/m5rPP00uHrT+9v\nky2jkQUQszuC9R4vK+tfs2S5z9w6qh3hwIJecChzWKce8hRZdiO9S7ix/6ZNiAgw\nY2h8AiG0VruPOJ6PbNXOFUTsajK0EP8AzJfNDIjvWHjUOawR352m4eKxXvXm9knd\nB/w1gY90jmAQ9JIiyOm+QlmHwO+qQUpWYOxt5Xnb0Pp/RRHEtxDgjygQWajAwsxG\nobx6sCf6+qcCAwEAAaOCAtYwggLSMB8GA1UdIwQYMBaAFI2MXsRUrYrhd+mb+ZsF\n4bgBjWHhMB0GA1UdDgQWBBTFjbuGoOUrgk9Dhr35DblkBZCj1jAOBgNVHQ8BAf8E\nBAMCBaAwDAYDVR0TAQH/BAIwADAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUH\nAwIwSQYDVR0gBEIwQDA0BgsrBgEEAbIxAQICBzAlMCMGCCsGAQUFBwIBFhdodHRw\nczovL3NlY3RpZ28uY29tL0NQUzAIBgZngQwBAgEwgYQGCCsGAQUFBwEBBHgwdjBP\nBggrBgEFBQcwAoZDaHR0cDovL2NydC5zZWN0aWdvLmNvbS9TZWN0aWdvUlNBRG9t\nYWluVmFsaWRhdGlvblNlY3VyZVNlcnZlckNBLmNydDAjBggrBgEFBQcwAYYXaHR0\ncDovL29jc3Auc2VjdGlnby5jb20wggEEBgorBgEEAdZ5AgQCBIH1BIHyAPAAdgBG\npVXrdfqRIDC1oolp9PN9ESxBdL79SbiFq/L8cP5tRwAAAXjDcW8TAAAEAwBHMEUC\nIQCxf+r8/dbHJDrh0YTAKSwdR8VUxAT6kHN5/HLuOvSsKgIgY2jAAf/tr59/f0JX\nKvHaN4qIv54gtj+KsNo7N0d4xcEAdgDfpV6raIJPH2yt7rhfTj5a6s2iEqRqXo47\nEsAgRFwqcwAAAXjDcW7VAAAEAwBHMEUCID0jtWvtpO1yypP7i7SeZZb3dQ6QdLlD\nlXpvWhjqrQfdAiEA0gp8tTUwOt2XN01OVTUrDgb4wV5VbFtx1SSYNFREQxwweQYD\nVR0RBHIwcIIKZ2l0bGFiLmNvbYIPYXV0aC5naXRsYWIuY29tghRjdXN0b21lcnMu\nZ2l0bGFiLmNvbYIaZW1haWwuY3VzdG9tZXJzLmdpdGxhYi5jb22CD2dwcmQuZ2l0\nbGFiLmNvbYIOd3d3LmdpdGxhYi5jb20wDQYJKoZIhvcNAQELBQADggEBAD7lgx6z\ncZI+uLtr7fYWOZDtPChNy7YjAXVtDbrQ61D1lESUIZwyDF9/xCDMqMSe+It2+j+t\nT0PHkbz6zbJdUMQhQxW0RLMZUthPg66YLqRJuvBU7VdWHxhqjfFb9UZvxOzTGgmN\nMuzmdThtlhRacNCTxGO/AJfcAt13RbKyR30UtqHb883qAH6isQvYFsQmijXcJXiT\ntRbcJ1Dm/dI+57BCTYLp2WfBdg0Axla5QsApQ+ER5GZoY1m6H3+OWpX77IdCgXF+\nHMtKCn08QLVBjhLr3IkeKgrYJTR1IDmzRwGUuUVvn1iO9+W10GV02SMngdN4nFp3\nwoE3CsYogf1SfQM=\n-----END CERTIFICATE-----",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("argocd_repository_certificate.simple", "https.0.server_name", serverName),
					resource.TestCheckResourceAttr("argocd_repository_certificate.simple", "https.0.cert_subtype", "rsa"),
					resource.TestCheckResourceAttrSet("argocd_repository_certificate.simple", "https.0.cert_info"),
				),
			},
		},
	})
}

func TestAccArgoCDRepositoryCertificatesSSH_Invalid(t *testing.T) {
	certSubType := acctest.RandomWithPrefix("cert")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDRepositoryCertificatesSSH(
					"",
					certSubType,
					"",
				),
				ExpectError: regexp.MustCompile("Invalid hostname in request"),
			},
			{
				Config: testAccArgoCDRepositoryCertificatesSSH(
					"dummy_server",
					certSubType,
					"",
				),
				ExpectError: regexp.MustCompile("invalid entry in known_hosts data"),
			},
		},
	})
}

func TestAccArgoCDRepositoryCertificates_Empty(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccArgoCDRepositoryCertificates_Empty(),
				ExpectError: regexp.MustCompile("one of `https,ssh` must be specified"),
			},
		},
	})
}

func TestAccArgoCDRepositoryCertificatesSSH_Allow_Random_Subtype(t *testing.T) {
	certSubType := acctest.RandomWithPrefix("cert")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDRepositoryCertificatesSSH(
					"dummy_server",
					certSubType,
					"AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFSMqzJeV9rUzU4kWitGjeR4PWSa29SPqJ1fVkhtj3Hw9xjLVXVYrU9QlYWrOLXBpQ6KWjbjTDTdDkoohFzgbEY=",
				),
			},
		},
	})
}

func TestAccArgoCDRepositoryCertificatesSSH_WithApplication(t *testing.T) {
	// Skip if we're not in an acceptance test environment
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}

	appName := acctest.RandomWithPrefix("testacc")

	subtypesKeys, err := getSshKeysForHost("private-git-repository")
	assert.NoError(t, err)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDRepositoryCertificateCredentialsApplicationWithSSH(appName, subtypesKeys),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"argocd_application.simple",
						"metadata.0.uid",
					),
					resource.TestCheckResourceAttr(
						"argocd_repository.private",
						"connection_state_status",
						"Successful",
					)),
			},
		},
	})
}

func TestAccArgoCDRepositoryCertificatesSSH_CannotUpdateExisting(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDRepositoryCertificatesSSH(
					"github.com",
					"ssh-rsa",
					// github's
					"AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==",
				),
				ExpectError: regexp.MustCompile("already exist and upsert was not specified"),
			},
		},
	})
}

func TestAccArgoCDRepositoryCertificatesSSH_CannotUpdateExisting_MultipleAtOnce(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDRepositoryCertificateSSH_Duplicated(
					"github.com",
					"ssh-rsaaa",
					// github's
					"AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==",
				),
				ExpectError: regexp.MustCompile("already exist and upsert was not specified"),
			},
		},
	})
}

func TestAccArgoCDRepositoryCertificatesHttps_CannotUpdateExisting_MultipleAtOnce(t *testing.T) {
	host := "github.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccArgoCDRepositoryCertificateHttps_Duplicated(
					host,
					// github's
					"-----BEGIN CERTIFICATE-----\nMIIFajCCBPCgAwIBAgIQBRiaVOvox+kD4KsNklVF3jAKBggqhkjOPQQDAzBWMQsw\nCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMTAwLgYDVQQDEydEaWdp\nQ2VydCBUTFMgSHlicmlkIEVDQyBTSEEzODQgMjAyMCBDQTEwHhcNMjIwMzE1MDAw\nMDAwWhcNMjMwMzE1MjM1OTU5WjBmMQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2Fs\naWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZyYW5jaXNjbzEVMBMGA1UEChMMR2l0SHVi\nLCBJbmMuMRMwEQYDVQQDEwpnaXRodWIuY29tMFkwEwYHKoZIzj0CAQYIKoZIzj0D\nAQcDQgAESrCTcYUh7GI/y3TARsjnANwnSjJLitVRgwgRI1JlxZ1kdZQQn5ltP3v7\nKTtYuDdUeEu3PRx3fpDdu2cjMlyA0aOCA44wggOKMB8GA1UdIwQYMBaAFAq8CCkX\njKU5bXoOzjPHLrPt+8N6MB0GA1UdDgQWBBR4qnLGcWloFLVZsZ6LbitAh0I7HjAl\nBgNVHREEHjAcggpnaXRodWIuY29tgg53d3cuZ2l0aHViLmNvbTAOBgNVHQ8BAf8E\nBAMCB4AwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMIGbBgNVHR8EgZMw\ngZAwRqBEoEKGQGh0dHA6Ly9jcmwzLmRpZ2ljZXJ0LmNvbS9EaWdpQ2VydFRMU0h5\nYnJpZEVDQ1NIQTM4NDIwMjBDQTEtMS5jcmwwRqBEoEKGQGh0dHA6Ly9jcmw0LmRp\nZ2ljZXJ0LmNvbS9EaWdpQ2VydFRMU0h5YnJpZEVDQ1NIQTM4NDIwMjBDQTEtMS5j\ncmwwPgYDVR0gBDcwNTAzBgZngQwBAgIwKTAnBggrBgEFBQcCARYbaHR0cDovL3d3\ndy5kaWdpY2VydC5jb20vQ1BTMIGFBggrBgEFBQcBAQR5MHcwJAYIKwYBBQUHMAGG\nGGh0dHA6Ly9vY3NwLmRpZ2ljZXJ0LmNvbTBPBggrBgEFBQcwAoZDaHR0cDovL2Nh\nY2VydHMuZGlnaWNlcnQuY29tL0RpZ2lDZXJ0VExTSHlicmlkRUNDU0hBMzg0MjAy\nMENBMS0xLmNydDAJBgNVHRMEAjAAMIIBfwYKKwYBBAHWeQIEAgSCAW8EggFrAWkA\ndgCt9776fP8QyIudPZwePhhqtGcpXc+xDCTKhYY069yCigAAAX+Oi8SRAAAEAwBH\nMEUCIAR9cNnvYkZeKs9JElpeXwztYB2yLhtc8bB0rY2ke98nAiEAjiML8HZ7aeVE\nP/DkUltwIS4c73VVrG9JguoRrII7gWMAdwA1zxkbv7FsV78PrUxtQsu7ticgJlHq\nP+Eq76gDwzvWTAAAAX+Oi8R7AAAEAwBIMEYCIQDNckqvBhup7GpANMf0WPueytL8\nu/PBaIAObzNZeNMpOgIhAMjfEtE6AJ2fTjYCFh/BNVKk1mkTwBTavJlGmWomQyaB\nAHYAs3N3B+GEUPhjhtYFqdwRCUp5LbFnDAuH3PADDnk2pZoAAAF/jovErAAABAMA\nRzBFAiEA9Uj5Ed/XjQpj/MxQRQjzG0UFQLmgWlc73nnt3CJ7vskCICqHfBKlDz7R\nEHdV5Vk8bLMBW1Q6S7Ga2SbFuoVXs6zFMAoGCCqGSM49BAMDA2gAMGUCMCiVhqft\n7L/stBmv1XqSRNfE/jG/AqKIbmjGTocNbuQ7kt1Cs7kRg+b3b3C9Ipu5FQIxAM7c\ntGKrYDGt0pH8iF6rzbp9Q4HQXMZXkNxg+brjWxnaOVGTDNwNH7048+s/hT9bUQ==\n-----END CERTIFICATE-----",
				),
				ExpectError: regexp.MustCompile(fmt.Sprintf("https certificate for '%s' already exist.", host)),
			},
		},
	})
}

func testAccArgoCDRepositoryCertificates_Empty() string {
	return `
resource "argocd_repository_certificate" "simple" {
}
`
}

func testAccArgoCDRepositoryCertificatesSSH(serverName, cert_subtype, cert_data string) string {
	return fmt.Sprintf(`
resource "argocd_repository_certificate" "simple" {
  ssh {
	server_name  = "%s"
	cert_subtype = "%s"
	cert_data    = <<EOT
%s
EOT
  }
}
`, serverName, cert_subtype, cert_data)
}

func testAccArgoCDRepositoryCertificateSSH_Duplicated(serverName, cert_subtype, cert_data string) string {
	return fmt.Sprintf(`
resource "argocd_repository_certificate" "simple" {
  ssh {
	server_name  = "%s"
	cert_subtype = "%s"
	cert_data    = <<EOT
%s
EOT
  }
}

resource "argocd_repository_certificate" "simple2" {
	ssh {
	  server_name  = "%s"
	  cert_subtype = "%s"
	  cert_data    = <<EOT
  %s
  EOT
	}
  }
`, serverName, cert_subtype, cert_data, serverName, cert_subtype, cert_data)
}

func testAccArgoCDRepositoryCertificateHttps(serverName, cert_data string) string {
	return fmt.Sprintf(`
resource "argocd_repository_certificate" "simple" {
  https {
    server_name  = "%s"
    cert_data    = <<EOT
%s
EOT
  }
}
`, serverName, cert_data)
}

func testAccArgoCDRepositoryCertificateHttps_Duplicated(serverName, cert_data string) string {
	return fmt.Sprintf(`
resource "argocd_repository_certificate" "simple" {
  https {
    server_name  = "%s"
    cert_data    = <<EOT
%s
EOT
  }
}

resource "argocd_repository_certificate" "simple2" {
  https {
    server_name  = "%s"
    cert_data    = <<EOT
%s
EOT
  }
}
`, serverName, cert_data, serverName, cert_data)
}

func testAccArgoCDRepositoryCertificateCredentialsApplicationWithSSH(random string, subtypesKeys []string) string {
	return fmt.Sprintf(`
resource "argocd_repository_certificate" "private-git-repository_0" {
	ssh {
		server_name  = "private-git-repository.argocd.svc.cluster.local"
		cert_subtype = "%s"
		cert_data    = <<EOT
		%s
EOT
	}
}
resource "argocd_repository_certificate" "private-git-repository_1" {
	ssh {
		server_name  = "private-git-repository.argocd.svc.cluster.local"
		cert_subtype = "%s"
		cert_data    = <<EOT
		%s
EOT
	}
}
resource "argocd_repository_certificate" "private-git-repository_2" {
	ssh {
		server_name  = "private-git-repository.argocd.svc.cluster.local"
		cert_subtype = "%s"
		cert_data    = <<EOT
		%s
EOT
	}
}

resource "argocd_repository" "private" {
  repo       = "git@private-git-repository.argocd.svc.cluster.local:~/project-1.git"
  insecure   = false
  depends_on = [
	  argocd_repository_credentials.private, 
	  argocd_repository_certificate.private-git-repository_0, 
	  argocd_repository_certificate.private-git-repository_1, 
	  argocd_repository_certificate.private-git-repository_2
  ]
}
 
resource "argocd_repository_credentials" "private" {
  url             = "git@private-git-repository.argocd.svc.cluster.local"
  username        = "git"
  ssh_private_key = "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW\nQyNTUxOQAAACCGe6Vx0gbKqKCI0wIplfgK5JBjCDO3bhtU3sZfLoeUZgAAAJB9cNEifXDR\nIgAAAAtzc2gtZWQyNTUxOQAAACCGe6Vx0gbKqKCI0wIplfgK5JBjCDO3bhtU3sZfLoeUZg\nAAAEAJeUrObjoTbGO1Sq4TXHl/j4RJ5aKMC1OemWuHmLK7XYZ7pXHSBsqooIjTAimV+Ark\nkGMIM7duG1Texl8uh5RmAAAAC3Rlc3RAYXJnb2NkAQI=\n-----END OPENSSH PRIVATE KEY-----"
}

resource "argocd_application" "simple" {
  metadata {
    name      = "%s"
    namespace = "argocd"
    labels = {
      acceptance = "true"
    }
  }

  spec {
    source {
      repo_url        = argocd_repository.private.repo
	  path = "."
    }

    destination {
      server    = "https://kubernetes.default.svc"
      namespace = "default"
    }
  }
}
`, subtypesKeys[0], subtypesKeys[1], subtypesKeys[2], subtypesKeys[3], subtypesKeys[4], subtypesKeys[5], random)
}

// Return an array like :
// [0] = ssh-rsa
// [1] = AAAAB3NzaC1y...
// [2] = ecdsa-sha2-nistp256
// [3] = AAAAB3NzaC1y...
// etc
func getSshKeysForHost(host string) ([]string, error) {
	app := "kubectl"
	args := []string{
		"exec",
		"-n",
		"argocd",
		"deploy/argocd-server",
		"--",
		"ssh-keyscan",
		host,
	}

	var err error

	var output []byte

	if testhelpers.GlobalTestEnv != nil {
		args = append([]string{app}, args...)

		output, err = testhelpers.GlobalTestEnv.ExecInK3s(context.Background(), args...)
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}

		n := strings.Split(string(output), "`")
		output = []byte(n[1])
	} else {
		cmd := exec.Command(app, args...)

		output, err = cmd.Output()
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
	}

	re, _ := regexp.Compile(`(?m)^private-git-repository (?P<subtype>[^\s]+) (?P<key>.+)$`)
	matches := re.FindAllStringSubmatch(string(output), 3)

	subTypesKeys := make([]string, 0)
	for _, match := range matches {
		subTypesKeys = append(subTypesKeys, match[1])
		subTypesKeys = append(subTypesKeys, match[2])
	}

	return subTypesKeys, nil
}

// TestAccArgoCDRepositoryCertificate_SSHConsistency tests consistency of SSH certificate fields
func TestAccArgoCDRepositoryCertificate_SSHConsistency(t *testing.T) {
	serverName := acctest.RandomWithPrefix("ssh-test")

	config := testAccArgoCDRepositoryCertificatesSSH(
		serverName,
		"ssh-rsa",
		"AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==",
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository_certificate.simple",
						"ssh.0.server_name",
						serverName,
					),
					resource.TestCheckResourceAttr(
						"argocd_repository_certificate.simple",
						"ssh.0.cert_subtype",
						"ssh-rsa",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_repository_certificate.simple",
						"ssh.0.cert_info",
					),
				),
			},
			{
				// Apply the same configuration again to test for consistency
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository_certificate.simple",
						"ssh.0.server_name",
						serverName,
					),
					resource.TestCheckResourceAttr(
						"argocd_repository_certificate.simple",
						"ssh.0.cert_subtype",
						"ssh-rsa",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_repository_certificate.simple",
						"ssh.0.cert_info",
					),
				),
			},
		},
	})
}

// TestAccArgoCDRepositoryCertificate_HTTPSConsistency tests consistency of HTTPS certificate fields
func TestAccArgoCDRepositoryCertificate_HTTPSConsistency(t *testing.T) {
	serverName := acctest.RandomWithPrefix("https-test")
	certData := "-----BEGIN CERTIFICATE-----\nMIIFajCCBPCgAwIBAgIQBRiaVOvox+kD4KsNklVF3jAKBggqhkjOPQQDAzBWMQsw\nCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMTAwLgYDVQQDEydEaWdp\nQ2VydCBUTFMgSHlicmlkIEVDQyBTSEEzODQgMjAyMCBDQTEwHhcNMjIwMzE1MDAw\nMDAwWhcNMjMwMzE1MjM1OTU5WjBmMQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2Fs\naWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZyYW5jaXNjbzEVMBMGA1UEChMMR2l0SHVi\nLCBJbmMuMRMwEQYDVQQDEwpnaXRodWIuY29tMFkwEwYHKoZIzj0CAQYIKoZIzj0D\nAQcDQgAESrCTcYUh7GI/y3TARsjnANwnSjJLitVRgwgRI1JlxZ1kdZQQn5ltP3v7\nKTtYuDdUeEu3PRx3fpDdu2cjMlyA0aOCA44wggOKMB8GA1UdIwQYMBaAFAq8CCkX\njKU5bXoOzjPHLrPt+8N6MB0GA1UdDgQWBBR4qnLGcWloFLVZsZ6LbitAh0I7HjAl\nBgNVHREEHjAcggpnaXRodWIuY29tgg53d3cuZ2l0aHViLmNvbTAOBgNVHQ8BAf8E\nBAMCB4AwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMIGbBgNVHR8EgZMw\ngZAwRqBEoEKGQGh0dHA6Ly9jcmwzLmRpZ2ljZXJ0LmNvbS9EaWdpQ2VydFRMU0h5\nYnJpZEVDQ1NIQTM4NDIwMjBDQTEtMS5jcmwwRqBEoEKGQGh0dHA6Ly9jcmw0LmRp\nZ2ljZXJ0LmNvbS9EaWdpQ2VydFRMU0h5YnJpZEVDQ1NIQTM4NDIwMjBDQTEtMS5j\ncmwwPgYDVR0gBDcwNTAzBgZngQwBAgIwKTAnBggrBgEFBQcCARYbaHR0cDovL3d3\ndy5kaWdpY2VydC5jb20vQ1BTMIGFBggrBgEFBQcBAQR5MHcwJAYIKwYBBQUHMAGG\nGGh0dHA6Ly9vY3NwLmRpZ2ljZXJ0LmNvbTBPBggrBgEFBQcwAoZDaHR0cDovL2Nh\nY2VydHMuZGlnaWNlcnQuY29tL0RpZ2lDZXJ0VExTSHlicmlkRUNDU0hBMzg0MjAy\nMENBMS0xLmNydDAJBgNVHRMEAjAAMIIBfwYKKwYBBAHWeQIEAgSCAW8EggFrAWkA\ndgCt9776fP8QyIudPZwePhhqtGcpXc+xDCTKhYY069yCigAAAX+Oi8SRAAAEAwBH\nMEUCIAR9cNnvYkZeKs9JElpeXwztYB2yLhtc8bB0rY2ke98nAiEAjiML8HZ7aeVE\nP/DkUltwIS4c73VVrG9JguoRrII7gWMAdwA1zxkbv7FsV78PrUxtQsu7ticgJlHq\nP+Eq76gDwzvWTAAAAX+Oi8R7AAAEAwBIMEYCIQDNckqvBhup7GpANMf0WPueytL8\nu/PBaIAObzNZeNMpOgIhAMjfEtE6AJ2fTjYCFh/BNVKk1mkTwBTavJlGmWomQyaB\nAHYAs3N3B+GEUPhjhtYFqdwRCUp5LbFnDAuH3PADDnk2pZoAAAF/jovErAAABAMA\nRzBFAiEA9Uj5Ed/XjQpj/MxQRQjzG0UFQLmgWlc73nnt3CJ7vskCICqHfBKlDz7R\nEHdV5Vk8bLMBW1Q6S7Ga2SbFuoVXs6zFMAoGCCqGSM49BAMDA2gAMGUCMCiVhqft\n7L/stBmv1XqSRNfE/jG/AqKIbmjGTocNbuQ7kt1Cs7kRg+b3b3C9Ipu5FQIxAM7c\ntGKrYDGt0pH8iF6rzbp9Q4HQXMZXkNxg+brjWxnaOVGTDNwNH7048+s/hT9bUQ==\n-----END CERTIFICATE-----"

	config := testAccArgoCDRepositoryCertificateHttps(serverName, certData)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository_certificate.simple",
						"https.0.server_name",
						serverName,
					),
					resource.TestCheckResourceAttrWith(
						"argocd_repository_certificate.simple",
						"https.0.cert_data",
						func(value string) error {
							// Not yet sure why the impl is suffixing with newline. Adding a newline only makes the test fail,
							// since it'll add yet another newline.
							require.Contains(t, value, certData)
							return nil
						},
					),
					resource.TestCheckResourceAttr(
						"argocd_repository_certificate.simple",
						"https.0.cert_subtype",
						"ecdsa",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_repository_certificate.simple",
						"https.0.cert_info",
					),
				),
			},
			{
				// Apply the same configuration again to test for consistency
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"argocd_repository_certificate.simple",
						"https.0.server_name",
						serverName,
					),
					resource.TestCheckResourceAttrWith(
						"argocd_repository_certificate.simple",
						"https.0.cert_data",
						func(value string) error {
							// Not yet sure why the impl is suffixing with newline. Adding a newline only makes the test fail,
							// since it'll add yet another newline.
							require.Contains(t, value, certData)
							return nil
						},
					),
					resource.TestCheckResourceAttr(
						"argocd_repository_certificate.simple",
						"https.0.cert_subtype",
						"ecdsa",
					),
					resource.TestCheckResourceAttrSet(
						"argocd_repository_certificate.simple",
						"https.0.cert_info",
					),
				),
			},
		},
	})
}

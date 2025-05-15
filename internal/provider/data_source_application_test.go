package provider

import (
	"fmt"
	"testing"

	"github.com/argoproj-labs/terraform-provider-argocd/internal/features"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccArgoCDApplicationDataSource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t); testAccPreCheckFeatureSupported(t, features.MultipleApplicationSources) },
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"argocd": {
						VersionConstraint: "~> 5.0",
						Source:            "argoproj-labs/argocd",
					},
				},
				Config: `
resource "argocd_project" "foo" {
	metadata {
		name      = "foo"
		namespace = "argocd"
	}

	spec {
		description       = "project with source namespace"
		source_repos      = ["*"]
		source_namespaces = ["mynamespace-1"]

		destination {
			server    = "https://kubernetes.default.svc"
			namespace = "mynamespace-1"
		}
	}
}

resource "argocd_application" "foo" {
	metadata {
		name      = "foo"
		namespace = "mynamespace-1"
		labels = {
			acceptance = "true"
		}
	}

	spec {
		destination {
		  server    = "https://kubernetes.default.svc"
		  namespace = "mynamespace-1"
		}

		ignore_difference {
			group               = "apps"
			kind                = "Deployment"
			jq_path_expressions = [".spec.replicas"]
			json_pointers       = ["/spec/replicas"]
		}

		info {
			name = "foo"
			value = "foo"
		}

		project                = argocd_project.foo.metadata[0].name
		revision_history_limit = 1
	
		source {
			repo_url        = "https://opensearch-project.github.io/helm-charts"
			chart           = "opensearch"
			target_revision = "3.0.0"

			helm {
				parameter {
					name = "replicas"
					value = "1"
				}
	
				parameter {
					name = "singleNode"
					value = "true"
				}

				parameter {
					name = "persistence.enabled"
					value = "false"
				}

				values = <<-EOT
				  extraEnvs:
				    - name: "DISABLE_SECURITY_PLUGIN"
				      value: "true"
				EOT
			}
		}
	
		source {
			repo_url        = "https://github.com/argoproj/argo-cd.git"
			path            = "test/e2e/testdata/guestbook"
			target_revision = "HEAD"
		}

		sync_policy {
			automated {
				allow_empty = true
				prune       = true
				self_heal   = true
			}

			retry {
				backoff {
					duration     = "30s"
					factor       = "2"
					max_duration = "2m"
				}

				limit = "5"
			}

			sync_options = ["ApplyOutOfSyncOnly=true"]
		}
	}

	wait = true
}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("argocd_application.foo", "metadata.0.uid"),
					resource.TestCheckResourceAttr("argocd_application.foo", "metadata.0.name", "foo"),
					resource.TestCheckResourceAttr("argocd_application.foo", "metadata.0.namespace", "mynamespace-1"),
					resource.TestCheckResourceAttrSet("argocd_application.foo", "metadata.0.labels.%"),
					resource.TestCheckResourceAttr("argocd_application.foo", "spec.0.destination.0.server", "https://kubernetes.default.svc"),
					resource.TestCheckResourceAttr("argocd_application.foo", "spec.0.destination.0.namespace", "mynamespace-1"),
					resource.TestCheckResourceAttr("argocd_application.foo", "spec.0.ignore_difference.0.group", "apps"),
					resource.TestCheckResourceAttr("argocd_application.foo", "spec.0.ignore_difference.0.kind", "Deployment"),
					resource.TestCheckResourceAttr("argocd_application.foo", "spec.0.ignore_difference.0.jq_path_expressions.0", ".spec.replicas"),
					resource.TestCheckResourceAttr("argocd_application.foo", "spec.0.ignore_difference.0.json_pointers.0", "/spec/replicas"),
					resource.TestCheckResourceAttr("argocd_application.foo", "spec.0.info.0.name", "foo"),
					resource.TestCheckResourceAttr("argocd_application.foo", "spec.0.info.0.value", "foo"),
					resource.TestCheckResourceAttr("argocd_application.foo", "spec.0.project", "foo"),
					resource.TestCheckResourceAttr("argocd_application.foo", "spec.0.revision_history_limit", "1"),
					resource.TestCheckResourceAttr("argocd_application.foo", "spec.0.source.0.repo_url", "https://opensearch-project.github.io/helm-charts"),
					resource.TestCheckResourceAttr("argocd_application.foo", "spec.0.source.0.chart", "opensearch"),
					resource.TestCheckResourceAttr("argocd_application.foo", "spec.0.source.0.target_revision", "3.0.0"),
					resource.TestCheckResourceAttr("argocd_application.foo", "spec.0.source.1.repo_url", "https://github.com/argoproj/argo-cd.git"),
					resource.TestCheckResourceAttr("argocd_application.foo", "spec.0.source.1.path", "test/e2e/testdata/guestbook"),
					resource.TestCheckResourceAttr("argocd_application.foo", "spec.0.source.1.target_revision", "HEAD"),
					resource.TestCheckResourceAttr("argocd_application.foo", "spec.0.sync_policy.0.automated.0.allow_empty", "true"),
					resource.TestCheckResourceAttr("argocd_application.foo", "spec.0.sync_policy.0.automated.0.prune", "true"),
					resource.TestCheckResourceAttr("argocd_application.foo", "spec.0.sync_policy.0.automated.0.self_heal", "true"),
					resource.TestCheckResourceAttr("argocd_application.foo", "spec.0.sync_policy.0.retry.0.backoff.0.duration", "30s"),
					resource.TestCheckResourceAttr("argocd_application.foo", "spec.0.sync_policy.0.retry.0.backoff.0.factor", "2"),
					resource.TestCheckResourceAttr("argocd_application.foo", "spec.0.sync_policy.0.retry.0.backoff.0.max_duration", "2m"),
					resource.TestCheckResourceAttr("argocd_application.foo", "spec.0.sync_policy.0.retry.0.limit", "5"),
					resource.TestCheckResourceAttr("argocd_application.foo", "spec.0.sync_policy.0.sync_options.0", "ApplyOutOfSyncOnly=true"),
				),
			},
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config: `
data "argocd_application" "foo" {
	metadata = {
		name      = "foo"
		namespace = "mynamespace-1"
	}
}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.argocd_application.foo", "metadata.uid"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "metadata.name", "foo"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "metadata.namespace", "mynamespace-1"),
					resource.TestCheckResourceAttrSet("data.argocd_application.foo", "metadata.labels.%"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "spec.destination.server", "https://kubernetes.default.svc"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "spec.destination.namespace", "mynamespace-1"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "spec.ignore_differences.0.group", "apps"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "spec.ignore_differences.0.kind", "Deployment"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "spec.ignore_differences.0.jq_path_expressions.0", ".spec.replicas"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "spec.ignore_differences.0.json_pointers.0", "/spec/replicas"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "spec.info.name", "foo"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "spec.info.value", "foo"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "spec.project", "foo"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "spec.revision_history_limit", "1"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "spec.sources.0.repo_url", "https://opensearch-project.github.io/helm-charts"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "spec.sources.0.chart", "opensearch"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "spec.sources.0.target_revision", "3.0.0"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "spec.sources.1.repo_url", "https://github.com/argoproj/argo-cd.git"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "spec.sources.1.path", "test/e2e/testdata/guestbook"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "spec.sources.1.target_revision", "HEAD"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "spec.sync_policy.automated.allow_empty", "true"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "spec.sync_policy.automated.prune", "true"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "spec.sync_policy.automated.self_heal", "true"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "spec.sync_policy.retry.backoff.duration", "30s"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "spec.sync_policy.retry.backoff.factor", "2"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "spec.sync_policy.retry.backoff.max_duration", "2m"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "spec.sync_policy.retry.limit", "5"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "spec.sync_policy.sync_options.0", "ApplyOutOfSyncOnly=true"),
					resource.TestCheckResourceAttrSet("data.argocd_application.foo", "status.conditions.%"),
					resource.TestCheckResourceAttr("data.argocd_application.foo", "status.health.status", "Healthy"),
					resource.TestCheckResourceAttrSet("data.argocd_application.foo", "status.operation_state"),
					resource.TestCheckResourceAttrSet("data.argocd_application.foo", "status.reconciled_at"),
					resource.TestCheckResourceAttrSet("data.argocd_application.foo", "status.resources.%"),
					resource.TestCheckResourceAttrSet("data.argocd_application.foo", "status.summary"),
					resource.TestCheckResourceAttrSet("data.argocd_application.foo", "status.sync"),
				),
				ExpectNonEmptyPlan: true,
				PlanOnly:           true,
			},
		},
	})
}

func TestAccArgoCDApplicationDataSource_Directory(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"argocd": {
						VersionConstraint: "~> 5.0",
						Source:            "argoproj-labs/argocd",
					},
				},
				Config: `
resource "argocd_application" "directory" {
	metadata {
		name      = "directory"
		namespace = "argocd"
	}

	spec {
		destination {
			server    = "https://kubernetes.default.svc"
			namespace = "directory"
		}

		source {
			repo_url        = "https://github.com/solo-io/gloo"
			path            = "install/helm/gloo"
			target_revision = "v1.4.2"

			directory {
				jsonnet {
					ext_var {
						name  = "somename"
						value = "somevalue"
						code  = false
					}

					libs = ["vendor", "foo"]

					tla {
						name  = "yetanothername"
						value = "yetanothervalue"
						code  = true
					}
				}

				recurse = false
			}
		}
	}
}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("argocd_application.directory", "metadata.0.uid"),
					resource.TestCheckResourceAttr("argocd_application.directory", "spec.0.source.0.directory.0.jsonnet.0.ext_var.0.name", "somename"),
					resource.TestCheckResourceAttr("argocd_application.directory", "spec.0.source.0.directory.0.jsonnet.0.ext_var.0.value", "somevalue"),
					resource.TestCheckResourceAttr("argocd_application.directory", "spec.0.source.0.directory.0.jsonnet.0.ext_var.0.code", "false"),
					resource.TestCheckResourceAttr("argocd_application.directory", "spec.0.source.0.directory.0.jsonnet.0.libs.0", "vendor"),
					resource.TestCheckResourceAttr("argocd_application.directory", "spec.0.source.0.directory.0.jsonnet.0.tla.0.name", "yetanothername"),
					resource.TestCheckResourceAttr("argocd_application.directory", "spec.0.source.0.directory.0.jsonnet.0.tla.0.value", "yetanothervalue"),
					resource.TestCheckResourceAttr("argocd_application.directory", "spec.0.source.0.directory.0.jsonnet.0.tla.0.code", "true"),
					resource.TestCheckResourceAttr("argocd_application.directory", "spec.0.source.0.directory.0.recurse", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config: `
data "argocd_application" "directory" {
	metadata = {
		name = "directory"
	}
}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.argocd_application.directory", "metadata.uid"),
					resource.TestCheckResourceAttr("data.argocd_application.directory", "spec.sources.0.directory.jsonnet.0.name", "image.tag"),
					resource.TestCheckResourceAttr("data.argocd_application.directory", "spec.sources.0.directory.jsonnet.ext_vars.0.name", "somename"),
					resource.TestCheckResourceAttr("data.argocd_application.directory", "spec.sources.0.directory.jsonnet.ext_vars.0.value", "somevalue"),
					resource.TestCheckResourceAttr("data.argocd_application.directory", "spec.sources.0.directory.jsonnet.ext_vars.0.code", "false"),
					resource.TestCheckResourceAttr("data.argocd_application.directory", "spec.sources.0.directory.jsonnet.libs.0", "vendor"),
					resource.TestCheckResourceAttr("data.argocd_application.directory", "spec.sources.0.directory.jsonnet.tlas.0.name", "yetanothername"),
					resource.TestCheckResourceAttr("data.argocd_application.directory", "spec.sources.0.directory.jsonnet.tlas.0.value", "yetanothervalue"),
					resource.TestCheckResourceAttr("data.argocd_application.directory", "spec.sources.0.directory.jsonnet.tlas.0.code", "true"),
					resource.TestCheckResourceAttr("data.argocd_application.directory", "spec.sources.0.directory.recurse", "false"),
				),
				ExpectNonEmptyPlan: true,
				PlanOnly:           true,
			},
		},
	})
}

func TestAccArgoCDApplicationDataSource_Helm(t *testing.T) {
	helmValues := `
ingress:
  enabled: true
  path: /
  hosts:
    - mydomain.example.com
  annotations:
    kubernetes.io/ingress.class: nginx
    kubernetes.io/tls-acme: "true"
  labels: {}
  tls:
    - secretName: mydomain-tls
      hosts:
        - mydomain.example.com
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"argocd": {
						VersionConstraint: "~> 5.0",
						Source:            "argoproj-labs/argocd",
					},
				},
				Config: fmt.Sprintf(`
resource "argocd_application" "helm" {
	metadata {
		name      = "helm"
		namespace = "argocd"
	}

	spec {
		destination {
			server    = "https://kubernetes.default.svc"
			namespace = "helm"
		}

		source {
			repo_url        = "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami"
			chart           = "redis"
			target_revision = "16.9.11"
			
			helm {
				ignore_missing_value_files = true

				# file_parameter {
				# 	name = "foo"
				# 	path = "values.yaml"
				# }

				parameter {
					force_string = true
					name         = "image.tag"
					value        = "6.2.5"
				}

				pass_credentials = true
				release_name     = "testing"
				skip_crds        = true
				value_files      = ["values.yaml"]
				values = <<EOT
%s
EOT
			}
		}
	}
}
				`, helmValues),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("argocd_application.helm", "metadata.0.uid"),
					resource.TestCheckResourceAttr("argocd_application.helm", "metadata.0.name", "helm"),
					resource.TestCheckResourceAttr("argocd_application.helm", "metadata.0.namespace", "argocd"),
					resource.TestCheckResourceAttr("argocd_application.helm", "spec.0.destination.0.server", "https://kubernetes.default.svc"),
					resource.TestCheckResourceAttr("argocd_application.helm", "spec.0.destination.0.namespace", "helm"),
					resource.TestCheckResourceAttr("argocd_application.helm", "spec.0.source.0.repo_url", "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami"),
					resource.TestCheckResourceAttr("argocd_application.helm", "spec.0.source.0.chart", "redis"),
					resource.TestCheckResourceAttr("argocd_application.helm", "spec.0.source.0.target_revision", "16.9.11"),
					resource.TestCheckResourceAttr("argocd_application.helm", "spec.0.source.0.helm.0.ignore_missing_value_files", "true"),
					resource.TestCheckResourceAttr("argocd_application.helm", "spec.0.source.0.helm.0.parameter.0.force_string", "true"),
					resource.TestCheckResourceAttr("argocd_application.helm", "spec.0.source.0.helm.0.parameter.0.name", "image.tag"),
					resource.TestCheckResourceAttr("argocd_application.helm", "spec.0.source.0.helm.0.parameter.0.value", "6.2.5"),
					resource.TestCheckResourceAttr("argocd_application.helm", "spec.0.source.0.helm.0.pass_credentials", "true"),
					resource.TestCheckResourceAttr("argocd_application.helm", "spec.0.source.0.helm.0.release_name", "testing"),
					resource.TestCheckResourceAttr("argocd_application.helm", "spec.0.source.0.helm.0.skip_crds", "true"),
					resource.TestCheckResourceAttr("argocd_application.helm", "spec.0.source.0.helm.0.value_files.0", "values.yaml"),
					resource.TestCheckResourceAttr("argocd_application.helm", "spec.0.source.0.helm.0.values", helmValues+"\n"),
				),
			},
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config: `
data "argocd_application" "helm" {
	metadata = {
		name = "helm"
	}
}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.argocd_application.helm", "metadata.uid"),
					resource.TestCheckResourceAttr("data.argocd_application.helm", "metadata.name", "helm"),
					resource.TestCheckResourceAttr("data.argocd_application.helm", "metadata.namespace", "argocd"),
					resource.TestCheckResourceAttr("data.argocd_application.helm", "spec.destination.server", "https://kubernetes.default.svc"),
					resource.TestCheckResourceAttr("data.argocd_application.helm", "spec.destination.namespace", "helm"),
					resource.TestCheckResourceAttr("data.argocd_application.helm", "spec.sources.0.repo_url", "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami"),
					resource.TestCheckResourceAttr("data.argocd_application.helm", "spec.sources.0.chart", "redis"),
					resource.TestCheckResourceAttr("data.argocd_application.helm", "spec.sources.0.target_revision", "16.9.11"),
					resource.TestCheckResourceAttr("data.argocd_application.helm", "spec.sources.0.helm.ignore_missing_value_files", "true"),
					resource.TestCheckResourceAttr("data.argocd_application.helm", "spec.sources.0.helm.parameters.0.force_string", "true"),
					resource.TestCheckResourceAttr("data.argocd_application.helm", "spec.sources.0.helm.parameters.0.name", "image.tag"),
					resource.TestCheckResourceAttr("data.argocd_application.helm", "spec.sources.0.helm.parameters.0.value", "6.2.5"),
					resource.TestCheckResourceAttr("data.argocd_application.helm", "spec.sources.0.helm.pass_credentials", "true"),
					resource.TestCheckResourceAttr("data.argocd_application.helm", "spec.sources.0.helm.release_name", "testing"),
					resource.TestCheckResourceAttr("data.argocd_application.helm", "spec.sources.0.helm.skip_crds", "true"),
					resource.TestCheckResourceAttr("data.argocd_application.helm", "spec.sources.0.helm.value_files.0", "values.yaml"),
					resource.TestCheckResourceAttr("data.argocd_application.helm", "spec.sources.0.helm.values", helmValues+"\n"),
				),
				ExpectNonEmptyPlan: true,
				PlanOnly:           true,
			},
		},
	})
}

func TestAccArgoCDApplicationDataSource_Kustomize(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"argocd": {
						VersionConstraint: "~> 5.0",
						Source:            "argoproj-labs/argocd",
					},
				},
				Config: `
resource "argocd_application" "kustomize" {
	metadata {
		name      = "kustomize"
		namespace = "argocd"
	}

	spec {
		destination {
			server    = "https://kubernetes.default.svc"
			namespace = "kustomize"
		}

		source {
			repo_url        = "https://github.com/kubernetes-sigs/kustomize"
			path            = "examples/helloWorld"
			target_revision = "release-kustomize-v3.7"

			kustomize {
				common_annotations = {
					"this.is.a.common" = "anno-tation"
				}

				common_labels = {
					"another.io/one" = "true" 
				}

				images      = ["hashicorp/terraform:light"]
				name_prefix = "foo-"
				name_suffix = "-bar"
				# version     = "v4.5.7"
			}
		}
	}
}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("argocd_application.kustomize", "metadata.0.uid"),
					resource.TestCheckResourceAttr("argocd_application.kustomize", "metadata.0.name", "kustomize"),
					resource.TestCheckResourceAttr("argocd_application.kustomize", "metadata.0.namespace", "argocd"),
					resource.TestCheckResourceAttr("argocd_application.kustomize", "spec.0.destination.0.server", "https://kubernetes.default.svc"),
					resource.TestCheckResourceAttr("argocd_application.kustomize", "spec.0.destination.0.namespace", "kustomize"),
					resource.TestCheckResourceAttr("argocd_application.kustomize", "spec.0.source.0.repo_url", "https://github.com/kubernetes-sigs/kustomize"),
					resource.TestCheckResourceAttr("argocd_application.kustomize", "spec.0.source.0.path", "examples/helloWorld"),
					resource.TestCheckResourceAttr("argocd_application.kustomize", "spec.0.source.0.target_revision", "release-kustomize-v3.7"),
					resource.TestCheckResourceAttrSet("argocd_application.kustomize", "spec.0.source.0.kustomize.0.common_annotations.%"),
					resource.TestCheckResourceAttrSet("argocd_application.kustomize", "spec.0.source.0.kustomize.0.common_labels.%"),
					resource.TestCheckResourceAttr("argocd_application.kustomize", "spec.0.source.0.kustomize.0.images.0", "hashicorp/terraform:light"),
					resource.TestCheckResourceAttr("argocd_application.kustomize", "spec.0.source.0.kustomize.0.name_prefix", "foo-"),
					resource.TestCheckResourceAttr("argocd_application.kustomize", "spec.0.source.0.kustomize.0.name_suffix", "-bar"),
					// resource.TestCheckResourceAttr("argocd_application.kustomize", "spec.0.source.0.kustomize.0.version", "v4.5.7"),
				),
			},
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config: `
data "argocd_application" "kustomize" {
	metadata = {
		name = "kustomize"
	}
}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.argocd_application.kustomize", "metadata.uid"),
					resource.TestCheckResourceAttr("data.argocd_application.kustomize", "metadata.name", "kustomize"),
					resource.TestCheckResourceAttr("data.argocd_application.kustomize", "metadata.namespace", "argocd"),
					resource.TestCheckResourceAttr("data.argocd_application.kustomize", "spec.destination.server", "https://kubernetes.default.svc"),
					resource.TestCheckResourceAttr("data.argocd_application.kustomize", "spec.destination.namespace", "kustomize"),
					resource.TestCheckResourceAttr("data.argocd_application.kustomize", "spec.sources.0.repo_url", "https://github.com/kubernetes-sigs/kustomize"),
					resource.TestCheckResourceAttr("data.argocd_application.kustomize", "spec.sources.0.path", "examples/helloWorld"),
					resource.TestCheckResourceAttr("data.argocd_application.kustomize", "spec.sources.0.target_revision", "release-kustomize-v3.7"),
					resource.TestCheckResourceAttrSet("argocd_application.kustomize", "spec.0.source.0.kustomize.common_annotations.%"),
					resource.TestCheckResourceAttrSet("argocd_application.kustomize", "spec.0.source.0.kustomize.common_labels.%"),
					resource.TestCheckResourceAttr("argocd_application.kustomize", "spec.0.source.0.kustomize.images.0", "hashicorp/terraform:light"),
					resource.TestCheckResourceAttr("argocd_application.kustomize", "spec.0.source.0.kustomize.name_prefix", "foo-"),
					resource.TestCheckResourceAttr("argocd_application.kustomize", "spec.0.source.0.kustomize.name_suffix", "-bar"),
					// resource.TestCheckResourceAttr("argocd_application.kustomize", "spec.0.source.0.kustomize.version", "v4.5.7"),
				),
				ExpectNonEmptyPlan: true,
				PlanOnly:           true,
			},
		},
	})
}

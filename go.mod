module github.com/oboukili/terraform-provider-argocd

go 1.16

require (
	cloud.google.com/go/storage v1.14.0 // indirect
	github.com/Masterminds/semver v1.5.0
	github.com/apparentlymart/go-cidr v1.1.0 // indirect
	github.com/argoproj/argo-cd/v2 v2.4.12
	github.com/argoproj/gitops-engine v0.7.3
	github.com/argoproj/pkg v0.11.1-0.20211203175135-36c59d8fafe0
	github.com/aws/aws-sdk-go v1.38.65 // indirect
	github.com/cristalhq/jwt/v3 v3.1.0
	github.com/golang/protobuf v1.5.2
	github.com/hashicorp/go-getter v1.5.4 // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/hcl/v2 v2.8.2 // indirect
	github.com/hashicorp/terraform-json v0.13.0 // indirect
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.7.1
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/robfig/cron v1.2.0
	github.com/stretchr/testify v1.7.1
	github.com/ulikunitz/xz v0.5.10 // indirect
	golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e
	k8s.io/apimachinery v0.23.3
	k8s.io/client-go v11.0.1-0.20190816222228-6d55c1b1f1ca+incompatible
	modernc.org/mathutil v1.0.0
)

replace (
	// https://github.com/golang/go/issues/33546#issuecomment-519656923
	github.com/go-check/check => github.com/go-check/check v0.0.0-20180628173108-788fd7840127

	github.com/golang/protobuf => github.com/golang/protobuf v1.4.2
	github.com/gorilla/websocket => github.com/gorilla/websocket v1.4.2
	github.com/grpc-ecosystem/grpc-gateway => github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/improbable-eng/grpc-web => github.com/improbable-eng/grpc-web v0.0.0-20181111100011-16092bd1d58a

	// Avoid CVE-2022-28948
	gopkg.in/yaml.v3 => gopkg.in/yaml.v3 v3.0.1

	// https://github.com/kubernetes/kubernetes/issues/79384#issuecomment-505627280
	k8s.io/api => k8s.io/api v0.23.1
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.23.1
	k8s.io/apimachinery => k8s.io/apimachinery v0.23.1
	k8s.io/apiserver => k8s.io/apiserver v0.23.1
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.23.1
	k8s.io/client-go => k8s.io/client-go v0.23.1
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.23.1
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.23.1
	k8s.io/code-generator => k8s.io/code-generator v0.23.1
	k8s.io/component-base => k8s.io/component-base v0.23.1
	k8s.io/component-helpers => k8s.io/component-helpers v0.23.1
	k8s.io/controller-manager => k8s.io/controller-manager v0.23.1
	k8s.io/cri-api => k8s.io/cri-api v0.23.1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.23.1
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.23.1
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.23.1
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.23.1
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.23.1
	k8s.io/kubectl => k8s.io/kubectl v0.23.1
	k8s.io/kubelet => k8s.io/kubelet v0.23.1
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.23.1
	k8s.io/metrics => k8s.io/metrics v0.23.1
	k8s.io/mount-utils => k8s.io/mount-utils v0.23.1
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.23.1
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.23.1
)

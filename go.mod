module github.com/oboukili/terraform-provider-argocd

go 1.14

require (
	github.com/Masterminds/semver v1.5.0
	github.com/argoproj/argo-cd v1.6.2
	github.com/argoproj/gitops-engine v0.1.3
	github.com/argoproj/pkg v0.0.0-20200319004004-f46beff7cd54
	github.com/cristalhq/jwt/v3 v3.0.2
	github.com/golang/protobuf v1.3.4
	github.com/hashicorp/terraform-plugin-sdk v1.14.0
	github.com/robfig/cron v1.1.0
	github.com/stretchr/testify v1.5.1
	golang.org/x/crypto v0.0.0-20190820162420-60c769a6c586
	k8s.io/apimachinery v0.16.6
	modernc.org/mathutil v1.0.0
)

replace (
	k8s.io/api => k8s.io/api v0.16.6
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.16.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.6
	k8s.io/apiserver => k8s.io/apiserver v0.16.6
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.16.6
	k8s.io/client-go => k8s.io/client-go v0.16.6
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.16.6
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.16.6
	k8s.io/code-generator => k8s.io/code-generator v0.16.6
	k8s.io/component-base => k8s.io/component-base v0.16.6
	k8s.io/cri-api => k8s.io/cri-api v0.16.6
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.16.6
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.16.6
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.16.6
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.16.6
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.16.6
	k8s.io/kubectl => k8s.io/kubectl v0.16.6
	k8s.io/kubelet => k8s.io/kubelet v0.16.6
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.16.6
	k8s.io/metrics => k8s.io/metrics v0.16.6
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.16.6
)

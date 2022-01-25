module github.com/oboukili/terraform-provider-argocd

go 1.16

require (
	cloud.google.com/go/storage v1.14.0 // indirect
	github.com/Masterminds/semver v1.5.0
	github.com/apparentlymart/go-cidr v1.1.0 // indirect
	github.com/argoproj/argo-cd/v2 v2.1.2
	github.com/argoproj/gitops-engine v0.4.1
	github.com/argoproj/pkg v0.9.1
	github.com/aws/aws-sdk-go v1.38.65 // indirect
	github.com/cristalhq/jwt/v3 v3.1.0
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.4.3
	github.com/hashicorp/go-cty v1.4.1-0.20200414143053-d3edf31b6320 // indirect
	github.com/hashicorp/go-getter v1.5.4 // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/hcl/v2 v2.8.2 // indirect
	github.com/hashicorp/terraform-json v0.13.0 // indirect
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.7.1
	github.com/klauspost/compress v1.13.1 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/robfig/cron v1.1.0
	github.com/stretchr/testify v1.7.0
	github.com/ulikunitz/xz v0.5.10 // indirect
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/sys v0.0.0-20210616094352-59db8d763f22 // indirect
	golang.org/x/tools v0.1.3 // indirect
	google.golang.org/api v0.44.0-impersonate-preview // indirect
	google.golang.org/grpc v1.38.0 // indirect
	google.golang.org/protobuf v1.26.0-rc.1 // indirect
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v11.0.1-0.20190816222228-6d55c1b1f1ca+incompatible
	modernc.org/mathutil v1.0.0
)

replace (
	github.com/golang/protobuf => github.com/golang/protobuf v1.4.2
	github.com/gorilla/websocket => github.com/gorilla/websocket v1.4.2
	github.com/grpc-ecosystem/grpc-gateway => github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/improbable-eng/grpc-web => github.com/improbable-eng/grpc-web v0.0.0-20181111100011-16092bd1d58a

	k8s.io/api => k8s.io/api v0.21.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.21.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.0
	k8s.io/apiserver => k8s.io/apiserver v0.21.0
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.21.0
	k8s.io/client-go => k8s.io/client-go v0.21.0
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.21.0
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.21.0
	k8s.io/code-generator => k8s.io/code-generator v0.21.0
	k8s.io/component-base => k8s.io/component-base v0.21.0
	k8s.io/component-helpers => k8s.io/component-helpers v0.21.0
	k8s.io/controller-manager => k8s.io/controller-manager v0.21.0
	k8s.io/cri-api => k8s.io/cri-api v0.21.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.21.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.21.0
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.21.0
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.21.0
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.21.0
	k8s.io/kubectl => k8s.io/kubectl v0.21.0
	k8s.io/kubelet => k8s.io/kubelet v0.21.0
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.21.0
	k8s.io/metrics => k8s.io/metrics v0.21.0
	k8s.io/mount-utils => k8s.io/mount-utils v0.21.0
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.21.0
)

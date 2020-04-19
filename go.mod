module github.com/oboukili/terraform-provider-argocd

go 1.14

require (
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/argoproj/argo-cd v1.5.2
	github.com/argoproj/pkg v0.0.0-20200319004004-f46beff7cd54 // indirect
	github.com/coreos/go-oidc v2.2.1+incompatible // indirect
	github.com/go-openapi/spec v0.19.7 // indirect
	github.com/go-redis/cache v6.4.0+incompatible // indirect
	github.com/go-redis/redis v6.15.7+incompatible // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.4.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.14.3 // indirect
	github.com/hashicorp/terraform-plugin-sdk v1.9.1
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/mitchellh/mapstructure v1.2.2
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/sirupsen/logrus v1.5.0 // indirect
	github.com/spf13/cobra v1.0.0 // indirect
	github.com/square/go-jose v2.5.0+incompatible
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	golang.org/x/crypto v0.0.0-20200414173820-0848c9571904 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	google.golang.org/genproto v0.0.0-20200417142217-fb6d0575620b // indirect
	google.golang.org/grpc v1.28.1 // indirect
	gopkg.in/square/go-jose.v2 v2.5.0 // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1 // indirect
	k8s.io/apimachinery v0.17.5
	k8s.io/client-go v0.17.5 // indirect
	k8s.io/kube-openapi v0.0.0-20200413232311-afe0b5e9f729 // indirect
)

replace (
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.0
	k8s.io/api => k8s.io/api v0.17.5
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.5
	k8s.io/client-go => k8s.io/client-go v0.17.5
)

package features

import (
	"github.com/Masterminds/semver/v3"
	"github.com/jmespath/go-jmespath"
)

type Feature int64

const (
	ExecLogsPolicy Feature = iota
	ProjectSourceNamespaces
	MultipleApplicationSources
	ApplicationSet
	ApplicationSetProgressiveSync
	ManagedNamespaceMetadata
	ApplicationSetApplicationsSyncPolicy
	ApplicationSetIgnoreApplicationDifferences
	ApplicationSetTemplatePatch
	ApplicationKustomizePatches
	ProjectDestinationServiceAccounts
	ProjectFineGrainedPolicy
	ApplicationSourceName
)

type FeatureConstraint struct {
	// Name is a human-readable name for the feature.
	Name string
	// MinVersion is the minimum ArgoCD version that supports this feature.
	MinVersion *semver.Version
	// RequiredSettings is a list of JMESPath expressions to evaluate (to true) against the ArgoCD server's settings for this feature to be used.
	RequiredSettings *[]*jmespath.JMESPath
}

var ConstraintsMap = map[Feature]FeatureConstraint{
	ExecLogsPolicy:                             {"exec/logs RBAC policy", semver.MustParse("2.4.4"), nil},
	ProjectSourceNamespaces:                    {"project source namespaces", semver.MustParse("2.5.0"), nil},
	MultipleApplicationSources:                 {"multiple application sources", semver.MustParse("2.6.3"), nil}, // Whilst the feature was introduced in 2.6.0 there was a bug that affects refresh of applications (and hence `wait` within this provider) that was only fixed in https://github.com/argoproj/argo-cd/pull/12576
	ApplicationSet:                             {"application sets", semver.MustParse("2.5.0"), nil},
	ApplicationSetProgressiveSync:              {"progressive sync (`strategy`)", semver.MustParse("2.6.0"), nil},
	ManagedNamespaceMetadata:                   {"managed namespace metadsata", semver.MustParse("2.6.0"), nil},
	ApplicationSetApplicationsSyncPolicy:       {"application set level application sync policy", semver.MustParse("2.8.0"), nil},
	ApplicationSetIgnoreApplicationDifferences: {"application set ignore application differences", semver.MustParse("2.9.0"), nil},
	ApplicationSetTemplatePatch:                {"application set template patch", semver.MustParse("2.10.0"), nil},
	ApplicationKustomizePatches:                {"application kustomize patches", semver.MustParse("2.9.0"), nil},
	ProjectDestinationServiceAccounts:          {"project destination service accounts", semver.MustParse("2.13.0"), &[]*jmespath.JMESPath{jmespath.MustCompile("impersonationEnabled")}},
	ProjectFineGrainedPolicy:                   {"fine-grained policy in project", semver.MustParse("2.12.0"), nil},
}

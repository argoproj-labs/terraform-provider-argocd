package features

import (
	"github.com/Masterminds/semver/v3"
)

type Feature int64

const (
	ExecLogsPolicy Feature = iota
	ProjectSourceNamespaces
	MultipleApplicationSources
	ApplicationSet
	ApplicationSetProgressiveSync
	ManagedNamespaceMetadata
)

type FeatureConstraint struct {
	Name       string
	MinVersion *semver.Version
}

var ConstraintsMap = map[Feature]FeatureConstraint{
	ExecLogsPolicy:                {"exec/logs RBAC policy", semver.MustParse("2.4.4")},
	ProjectSourceNamespaces:       {"project source namespaces", semver.MustParse("2.5.0")},
	MultipleApplicationSources:    {"multiple application sources", semver.MustParse("2.6.3")}, // Whilst the feature was introduced in 2.6.0 there was a bug that affects refresh of applications (and hence `wait` within this provider) that was only fixed in https://github.com/argoproj/argo-cd/pull/12576
	ApplicationSet:                {"application sets", semver.MustParse("2.5.0")},
	ApplicationSetProgressiveSync: {"progressive sync (`strategy`)", semver.MustParse("2.6.0")},
	ManagedNamespaceMetadata:      {"managed namespace metadsata", semver.MustParse("2.6.0")},
}

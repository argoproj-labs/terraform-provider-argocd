package argocd

import (
	"encoding/json"
	"fmt"
	"reflect"

	application "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/oboukili/terraform-provider-argocd/internal/features"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func expandApplicationSet(d *schema.ResourceData, featureMultipleApplicationSourcesSupported bool) (metadata meta.ObjectMeta, spec application.ApplicationSetSpec, err error) {
	metadata = expandMetadata(d)
	spec, err = expandApplicationSetSpec(d, featureMultipleApplicationSourcesSupported)

	return
}

func expandApplicationSetSpec(d *schema.ResourceData, featureMultipleApplicationSourcesSupported bool) (spec application.ApplicationSetSpec, err error) {
	s := d.Get("spec.0").(map[string]interface{})

	if v, ok := s["generator"].([]interface{}); ok && len(v) > 0 {
		spec.Generators, err = expandApplicationSetGenerators(v, featureMultipleApplicationSourcesSupported)
		if err != nil {
			return
		}
	}

	spec.GoTemplate = s["go_template"].(bool)

	if v, ok := s["strategy"].([]interface{}); ok && len(v) > 0 {
		spec.Strategy, err = expandApplicationSetStrategy(v[0].(map[string]interface{}))
		if err != nil {
			return
		}
	}

	if v, ok := s["sync_policy"].([]interface{}); ok && len(v) > 0 {
		spec.SyncPolicy = expandApplicationSetSyncPolicy(v[0].(map[string]interface{}))
	}

	if v, ok := s["template"].([]interface{}); ok && len(v) > 0 {
		spec.Template, err = expandApplicationSetTemplate(v[0], featureMultipleApplicationSourcesSupported)
		if err != nil {
			return
		}
	}

	return spec, nil
}

func expandApplicationSetGenerators(g []interface{}, featureMultipleApplicationSourcesSupported bool) ([]application.ApplicationSetGenerator, error) {
	asgs := make([]application.ApplicationSetGenerator, len(g))

	for i, v := range g {
		v := v.(map[string]interface{})

		var g *application.ApplicationSetGenerator

		var err error

		if asg, ok := v["clusters"].([]interface{}); ok && len(asg) > 0 {
			g, err = expandApplicationSetClustersGenerator(asg[0], featureMultipleApplicationSourcesSupported)
		} else if asg, ok = v["cluster_decision_resource"].([]interface{}); ok && len(asg) > 0 {
			g, err = expandApplicationSetClusterDecisionResourceGenerator(asg[0], featureMultipleApplicationSourcesSupported)
		} else if asg, ok = v["git"].([]interface{}); ok && len(asg) > 0 {
			g, err = expandApplicationSetGitGenerator(asg[0], featureMultipleApplicationSourcesSupported)
		} else if asg, ok = v["list"].([]interface{}); ok && len(asg) > 0 {
			g, err = expandApplicationSetListGenerator(asg[0], featureMultipleApplicationSourcesSupported)
		} else if asg, ok = v["matrix"].([]interface{}); ok && len(asg) > 0 {
			g, err = expandApplicationSetMatrixGenerator(asg[0], featureMultipleApplicationSourcesSupported)
		} else if asg, ok = v["merge"].([]interface{}); ok && len(asg) > 0 {
			g, err = expandApplicationSetMergeGenerator(asg[0], featureMultipleApplicationSourcesSupported)
		} else if asg, ok = v["scm_provider"].([]interface{}); ok && len(asg) > 0 {
			g, err = expandApplicationSetSCMProviderGenerator(asg[0], featureMultipleApplicationSourcesSupported)
		} else if asg, ok = v["pull_request"].([]interface{}); ok && len(asg) > 0 {
			g, err = expandApplicationSetPullRequestGeneratorGenerator(asg[0], featureMultipleApplicationSourcesSupported)
		}

		if err != nil {
			return nil, err
		}

		if s, ok := v["selector"].([]interface{}); ok && len(s) > 0 {
			ls := expandLabelSelector(s)
			g.Selector = &ls
		}

		asgs[i] = *g
	}

	return asgs, nil
}

func expandApplicationSetClustersGenerator(cg interface{}, featureMultipleApplicationSourcesSupported bool) (*application.ApplicationSetGenerator, error) {
	asg := &application.ApplicationSetGenerator{
		Clusters: &application.ClusterGenerator{},
	}

	c := cg.(map[string]interface{})

	if v, ok := c["selector"]; ok {
		asg.Clusters.Selector = expandLabelSelector(v.([]interface{}))
	}

	if v, ok := c["template"].([]interface{}); ok && len(v) > 0 {
		temp, err := expandApplicationSetTemplate(v[0], featureMultipleApplicationSourcesSupported)
		if err != nil {
			return nil, err
		}

		asg.Clusters.Template = temp
	}

	if v, ok := c["values"]; ok {
		asg.Clusters.Values = expandStringMap(v.(map[string]interface{}))
	}

	return asg, nil
}

func expandApplicationSetClusterDecisionResourceGenerator(cdrg interface{}, featureMultipleApplicationSourcesSupported bool) (*application.ApplicationSetGenerator, error) {
	c := cdrg.(map[string]interface{})

	asg := &application.ApplicationSetGenerator{
		ClusterDecisionResource: &application.DuckTypeGenerator{
			ConfigMapRef: c["config_map_ref"].(string),
			Name:         c["name"].(string),
		},
	}

	if v, ok := c["label_selector"]; ok {
		asg.ClusterDecisionResource.LabelSelector = expandLabelSelector(v.([]interface{}))
	}

	if v, ok := c["requeue_after_seconds"].(string); ok && len(v) > 0 {
		ras, err := convertStringToInt64Pointer(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert requeue_after_seconds to *int64: %w", err)
		}

		asg.ClusterDecisionResource.RequeueAfterSeconds = ras
	}

	if v, ok := c["template"].([]interface{}); ok && len(v) > 0 {
		temp, err := expandApplicationSetTemplate(v[0], featureMultipleApplicationSourcesSupported)
		if err != nil {
			return nil, err
		}

		asg.ClusterDecisionResource.Template = temp
	}

	if v, ok := c["values"]; ok {
		asg.ClusterDecisionResource.Values = expandStringMap(v.(map[string]interface{}))
	}

	return asg, nil
}

func expandApplicationSetGitGenerator(gg interface{}, featureMultipleApplicationSourcesSupported bool) (*application.ApplicationSetGenerator, error) {
	g := gg.(map[string]interface{})

	asg := &application.ApplicationSetGenerator{
		Git: &application.GitGenerator{
			RepoURL:  g["repo_url"].(string),
			Revision: g["revision"].(string),
		},
	}

	if v, ok := g["directory"].([]interface{}); ok && len(v) > 0 {
		for _, d := range v {
			d := d.(map[string]interface{})

			dir := application.GitDirectoryGeneratorItem{
				Path: d["path"].(string),
			}

			if e, ok := d["exclude"].(bool); ok {
				dir.Exclude = e
			}

			asg.Git.Directories = append(asg.Git.Directories, dir)
		}
	}

	if v, ok := g["file"].([]interface{}); ok && len(v) > 0 {
		for _, f := range v {
			f := f.(map[string]interface{})

			file := application.GitFileGeneratorItem{
				Path: f["path"].(string),
			}

			asg.Git.Files = append(asg.Git.Files, file)
		}
	}

	if v, ok := g["template"].([]interface{}); ok && len(v) > 0 {
		temp, err := expandApplicationSetTemplate(v[0], featureMultipleApplicationSourcesSupported)
		if err != nil {
			return nil, err
		}

		asg.Git.Template = temp
	}

	return asg, nil
}

func expandApplicationSetListGenerator(lg interface{}, featureMultipleApplicationSourcesSupported bool) (*application.ApplicationSetGenerator, error) {
	asg := &application.ApplicationSetGenerator{
		List: &application.ListGenerator{},
	}

	l := lg.(map[string]interface{})

	e := l["elements"].([]interface{})

	for _, v := range e {
		data, err := json.Marshal(v)
		if err != nil {
			return asg, fmt.Errorf("failed to marshal list generator value: %w", err)
		}

		asg.List.Elements = append(asg.List.Elements, apiextensionsv1.JSON{
			Raw: data,
		})
	}

	if v, ok := l["template"].([]interface{}); ok && len(v) > 0 {
		temp, err := expandApplicationSetTemplate(v[0], featureMultipleApplicationSourcesSupported)
		if err != nil {
			return nil, err
		}

		asg.List.Template = temp
	}

	return asg, nil
}

func expandApplicationSetMatrixGenerator(mg interface{}, featureMultipleApplicationSourcesSupported bool) (*application.ApplicationSetGenerator, error) {
	asg := &application.ApplicationSetGenerator{}
	m := mg.(map[string]interface{})

	gs := m["generator"].([]interface{})

	asgs, err := expandApplicationSetGenerators(gs, featureMultipleApplicationSourcesSupported)
	if err != nil {
		return nil, err
	}

	ngs := make([]application.ApplicationSetNestedGenerator, len(asgs))
	for i, g := range asgs {
		ngs[i] = application.ApplicationSetNestedGenerator{
			ClusterDecisionResource: g.ClusterDecisionResource,
			Clusters:                g.Clusters,
			Git:                     g.Git,
			List:                    g.List,
			PullRequest:             g.PullRequest,
			SCMProvider:             g.SCMProvider,
		}

		if g.Matrix != nil {
			json, err := json.Marshal(g.Matrix)
			if err != nil {
				return asg, fmt.Errorf("failed to marshal nested matrix generator to json: %w", err)
			}

			ngs[i].Matrix = &apiextensionsv1.JSON{
				Raw: json,
			}
		}

		if g.Merge != nil {
			json, err := json.Marshal(g.Merge)
			if err != nil {
				return asg, fmt.Errorf("failed to marshal nested merge generator to json: %w", err)
			}

			ngs[i].Merge = &apiextensionsv1.JSON{
				Raw: json,
			}
		}
	}

	asg.Matrix = &application.MatrixGenerator{
		Generators: ngs,
	}

	if v, ok := m["template"].([]interface{}); ok && len(v) > 0 {
		temp, err := expandApplicationSetTemplate(v[0], featureMultipleApplicationSourcesSupported)
		if err != nil {
			return nil, err
		}

		asg.Matrix.Template = temp
	}

	return asg, nil
}

func expandApplicationSetMergeGenerator(mg interface{}, featureMultipleApplicationSourcesSupported bool) (*application.ApplicationSetGenerator, error) {
	asg := &application.ApplicationSetGenerator{
		Merge: &application.MergeGenerator{},
	}

	m := mg.(map[string]interface{})

	mks := m["merge_keys"].([]interface{})
	for _, k := range mks {
		asg.Merge.MergeKeys = append(asg.Merge.MergeKeys, k.(string))
	}

	gs := m["generator"].([]interface{})

	asgs, err := expandApplicationSetGenerators(gs, featureMultipleApplicationSourcesSupported)
	if err != nil {
		return nil, err
	}

	ngs := make([]application.ApplicationSetNestedGenerator, len(asgs))
	for i, g := range asgs {
		ngs[i] = application.ApplicationSetNestedGenerator{
			ClusterDecisionResource: g.ClusterDecisionResource,
			Clusters:                g.Clusters,
			Git:                     g.Git,
			List:                    g.List,
			PullRequest:             g.PullRequest,
			SCMProvider:             g.SCMProvider,
		}

		if g.Matrix != nil {
			json, err := json.Marshal(g.Matrix)
			if err != nil {
				return asg, fmt.Errorf("failed to marshal nested matrix generator to json: %w", err)
			}

			ngs[i].Matrix = &apiextensionsv1.JSON{
				Raw: json,
			}
		}

		if g.Merge != nil {
			json, err := json.Marshal(g.Merge)
			if err != nil {
				return asg, fmt.Errorf("failed to marshal nested merge generator to json: %w", err)
			}

			ngs[i].Merge = &apiextensionsv1.JSON{
				Raw: json,
			}
		}
	}

	asg.Merge.Generators = ngs

	if v, ok := m["template"].([]interface{}); ok && len(v) > 0 {
		temp, err := expandApplicationSetTemplate(v[0], featureMultipleApplicationSourcesSupported)
		if err != nil {
			return nil, err
		}

		asg.Merge.Template = temp
	}

	return asg, nil
}

func expandApplicationSetPullRequestGeneratorGenerator(mg interface{}, featureMultipleApplicationSourcesSupported bool) (*application.ApplicationSetGenerator, error) {
	asg := &application.ApplicationSetGenerator{
		PullRequest: &application.PullRequestGenerator{},
	}

	m := mg.(map[string]interface{})

	if v, ok := m["bitbucket_server"].([]interface{}); ok && len(v) > 0 {
		asg.PullRequest.BitbucketServer = expandApplicationSetPullRequestGeneratorBitbucketServer(v[0].(map[string]interface{}))
	} else if v, ok := m["gitea"].([]interface{}); ok && len(v) > 0 {
		asg.PullRequest.Gitea = expandApplicationSetPullRequestGeneratorGitea(v[0].(map[string]interface{}))
	} else if v, ok := m["github"].([]interface{}); ok && len(v) > 0 {
		asg.PullRequest.Github = expandApplicationSetPullRequestGeneratorGithub(v[0].(map[string]interface{}))
	} else if v, ok := m["gitlab"].([]interface{}); ok && len(v) > 0 {
		asg.PullRequest.GitLab = expandApplicationSetPullRequestGeneratorGitlab(v[0].(map[string]interface{}))
	}

	if v, ok := m["filter"].([]interface{}); ok && len(v) > 0 {
		asg.PullRequest.Filters = expandApplicationSetPullRequestGeneratorFilters(v)
	}

	if v, ok := m["requeue_after_seconds"].(string); ok && v != "" {
		ras, err := convertStringToInt64Pointer(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert requeue_after_seconds to *int64: %w", err)
		}

		asg.PullRequest.RequeueAfterSeconds = ras
	}

	if v, ok := m["template"].([]interface{}); ok && len(v) > 0 {
		temp, err := expandApplicationSetTemplate(v[0], featureMultipleApplicationSourcesSupported)
		if err != nil {
			return nil, err
		}

		asg.PullRequest.Template = temp
	}

	return asg, nil
}

func expandApplicationSetPullRequestGeneratorBitbucketServer(bs map[string]interface{}) *application.PullRequestGeneratorBitbucketServer {
	spgbs := &application.PullRequestGeneratorBitbucketServer{
		API:     bs["api"].(string),
		Project: bs["project"].(string),
		Repo:    bs["repo"].(string),
	}

	if v, ok := bs["basic_auth"].([]interface{}); ok && len(v) > 0 {
		ba := v[0].(map[string]interface{})

		spgbs.BasicAuth = &application.BasicAuthBitbucketServer{
			Username: ba["username"].(string),
		}

		if pr, ok := ba["password_ref"].([]interface{}); ok && len(pr) > 0 {
			spgbs.BasicAuth.PasswordRef = expandSecretRef(pr[0].(map[string]interface{}))
		}
	}

	return spgbs
}

func expandApplicationSetPullRequestGeneratorGitea(g map[string]interface{}) *application.PullRequestGeneratorGitea {
	prgg := &application.PullRequestGeneratorGitea{
		API:      g["api"].(string),
		Insecure: g["insecure"].(bool),
		Owner:    g["owner"].(string),
		Repo:     g["repo"].(string),
	}

	if v, ok := g["token_ref"].([]interface{}); ok && len(v) > 0 {
		prgg.TokenRef = expandSecretRef(v[0].(map[string]interface{}))
	}

	return prgg
}

func expandApplicationSetPullRequestGeneratorGithub(g map[string]interface{}) *application.PullRequestGeneratorGithub {
	spgg := &application.PullRequestGeneratorGithub{
		API:           g["api"].(string),
		AppSecretName: g["app_secret_name"].(string),
		Owner:         g["owner"].(string),
		Repo:          g["repo"].(string),
	}

	if v, ok := g["labels"].([]interface{}); ok && len(v) > 0 {
		for _, l := range v {
			spgg.Labels = append(spgg.Labels, l.(string))
		}
	}

	if v, ok := g["token_ref"].([]interface{}); ok && len(v) > 0 {
		spgg.TokenRef = expandSecretRef(v[0].(map[string]interface{}))
	}

	return spgg
}

func expandApplicationSetPullRequestGeneratorGitlab(g map[string]interface{}) *application.PullRequestGeneratorGitLab {
	spgg := &application.PullRequestGeneratorGitLab{
		API:              g["api"].(string),
		Project:          g["project"].(string),
		PullRequestState: g["pull_request_state"].(string),
	}

	if v, ok := g["labels"].([]interface{}); ok && len(v) > 0 {
		for _, l := range v {
			spgg.Labels = append(spgg.Labels, l.(string))
		}
	}

	if v, ok := g["token_ref"].([]interface{}); ok && len(v) > 0 {
		spgg.TokenRef = expandSecretRef(v[0].(map[string]interface{}))
	}

	return spgg
}

func expandApplicationSetPullRequestGeneratorFilters(fs []interface{}) []application.PullRequestGeneratorFilter {
	prgfs := make([]application.PullRequestGeneratorFilter, len(fs))

	for i, v := range fs {
		f := v.(map[string]interface{})
		spgf := application.PullRequestGeneratorFilter{}

		if bm, ok := f["branch_match"].(string); ok && bm != "" {
			spgf.BranchMatch = &bm
		}

		prgfs[i] = spgf
	}

	return prgfs
}

func expandApplicationSetSCMProviderGenerator(mg interface{}, featureMultipleApplicationSourcesSupported bool) (*application.ApplicationSetGenerator, error) {
	m := mg.(map[string]interface{})

	asg := &application.ApplicationSetGenerator{
		SCMProvider: &application.SCMProviderGenerator{
			CloneProtocol: m["clone_protocol"].(string),
		},
	}

	if v, ok := m["azure_devops"].([]interface{}); ok && len(v) > 0 {
		asg.SCMProvider.AzureDevOps = expandApplicationSetSCMProviderAzureDevOps(v[0].(map[string]interface{}))
	} else if v, ok := m["bitbucket_cloud"].([]interface{}); ok && len(v) > 0 {
		asg.SCMProvider.Bitbucket = expandApplicationSetSCMProviderBitbucket(v[0].(map[string]interface{}))
	} else if v, ok := m["bitbucket_server"].([]interface{}); ok && len(v) > 0 {
		asg.SCMProvider.BitbucketServer = expandApplicationSetSCMProviderBitbucketServer(v[0].(map[string]interface{}))
	} else if v, ok := m["gitea"].([]interface{}); ok && len(v) > 0 {
		asg.SCMProvider.Gitea = expandApplicationSetSCMProviderGitea(v[0].(map[string]interface{}))
	} else if v, ok := m["github"].([]interface{}); ok && len(v) > 0 {
		asg.SCMProvider.Github = expandApplicationSetSCMProviderGithub(v[0].(map[string]interface{}))
	} else if v, ok := m["gitlab"].([]interface{}); ok && len(v) > 0 {
		asg.SCMProvider.Gitlab = expandApplicationSetSCMProviderGitlab(v[0].(map[string]interface{}))
	}

	if v, ok := m["filter"].([]interface{}); ok && len(v) > 0 {
		asg.SCMProvider.Filters = expandApplicationSetSCMProviderGeneratorFilters(v)
	}

	if v, ok := m["requeue_after_seconds"].(string); ok && v != "" {
		ras, err := convertStringToInt64Pointer(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert requeue_after_seconds to *int64: %w", err)
		}

		asg.SCMProvider.RequeueAfterSeconds = ras
	}

	if v, ok := m["template"].([]interface{}); ok && len(v) > 0 {
		temp, err := expandApplicationSetTemplate(v[0], featureMultipleApplicationSourcesSupported)
		if err != nil {
			return nil, err
		}

		asg.SCMProvider.Template = temp
	}

	return asg, nil
}

func expandApplicationSetSCMProviderAzureDevOps(ado map[string]interface{}) *application.SCMProviderGeneratorAzureDevOps {
	spgado := &application.SCMProviderGeneratorAzureDevOps{
		AllBranches:  ado["all_branches"].(bool),
		API:          ado["api"].(string),
		Organization: ado["organization"].(string),
		TeamProject:  ado["team_project"].(string),
	}

	if v, ok := ado["access_token_ref"].([]interface{}); ok && len(v) > 0 {
		spgado.AccessTokenRef = expandSecretRef(v[0].(map[string]interface{}))
	}

	return spgado
}

func expandApplicationSetSCMProviderBitbucket(b map[string]interface{}) *application.SCMProviderGeneratorBitbucket {
	spgb := &application.SCMProviderGeneratorBitbucket{
		AllBranches: b["all_branches"].(bool),
		Owner:       b["owner"].(string),
		User:        b["user"].(string),
	}

	if v, ok := b["app_password_ref"].([]interface{}); ok && len(v) > 0 {
		spgb.AppPasswordRef = expandSecretRef(v[0].(map[string]interface{}))
	}

	return spgb
}

func expandApplicationSetSCMProviderBitbucketServer(bs map[string]interface{}) *application.SCMProviderGeneratorBitbucketServer {
	spgbs := &application.SCMProviderGeneratorBitbucketServer{
		AllBranches: bs["all_branches"].(bool),
		API:         bs["api"].(string),
		Project:     bs["project"].(string),
	}

	if v, ok := bs["basic_auth"].([]interface{}); ok && len(v) > 0 {
		ba := v[0].(map[string]interface{})

		spgbs.BasicAuth = &application.BasicAuthBitbucketServer{
			Username: ba["username"].(string),
		}

		if pr, ok := ba["password_ref"].([]interface{}); ok && len(pr) > 0 {
			spgbs.BasicAuth.PasswordRef = expandSecretRef(pr[0].(map[string]interface{}))
		}
	}

	return spgbs
}

func expandApplicationSetSCMProviderGitea(g map[string]interface{}) *application.SCMProviderGeneratorGitea {
	spgg := &application.SCMProviderGeneratorGitea{
		AllBranches: g["all_branches"].(bool),
		API:         g["api"].(string),
		Insecure:    g["insecure"].(bool),
		Owner:       g["owner"].(string),
	}

	if v, ok := g["token_ref"].([]interface{}); ok && len(v) > 0 {
		spgg.TokenRef = expandSecretRef(v[0].(map[string]interface{}))
	}

	return spgg
}

func expandApplicationSetSCMProviderGithub(g map[string]interface{}) *application.SCMProviderGeneratorGithub {
	spgg := &application.SCMProviderGeneratorGithub{
		AllBranches:   g["all_branches"].(bool),
		API:           g["api"].(string),
		Organization:  g["organization"].(string),
		AppSecretName: g["app_secret_name"].(string),
	}

	if v, ok := g["token_ref"].([]interface{}); ok && len(v) > 0 {
		spgg.TokenRef = expandSecretRef(v[0].(map[string]interface{}))
	}

	return spgg
}

func expandApplicationSetSCMProviderGitlab(g map[string]interface{}) *application.SCMProviderGeneratorGitlab {
	spgg := &application.SCMProviderGeneratorGitlab{
		AllBranches:      g["all_branches"].(bool),
		API:              g["api"].(string),
		IncludeSubgroups: g["include_subgroups"].(bool),
		Group:            g["group"].(string),
	}

	if v, ok := g["token_ref"].([]interface{}); ok && len(v) > 0 {
		spgg.TokenRef = expandSecretRef(v[0].(map[string]interface{}))
	}

	return spgg
}

func expandApplicationSetSCMProviderGeneratorFilters(fs []interface{}) []application.SCMProviderGeneratorFilter {
	spgfs := make([]application.SCMProviderGeneratorFilter, len(fs))

	for i, v := range fs {
		f := v.(map[string]interface{})
		spgf := application.SCMProviderGeneratorFilter{}

		if bm, ok := f["branch_match"].(string); ok && bm != "" {
			spgf.BranchMatch = &bm
		}

		if lm, ok := f["label_match"].(string); ok && lm != "" {
			spgf.LabelMatch = &lm
		}

		if pdne, ok := f["paths_do_not_exist"].([]interface{}); ok && len(pdne) > 0 {
			for _, p := range pdne {
				spgf.PathsDoNotExist = append(spgf.PathsDoNotExist, p.(string))
			}
		}

		if pe, ok := f["paths_exist"].([]interface{}); ok && len(pe) > 0 {
			for _, p := range pe {
				spgf.PathsExist = append(spgf.PathsExist, p.(string))
			}
		}

		if rm, ok := f["repository_match"].(string); ok && rm != "" {
			spgf.RepositoryMatch = &rm
		}

		spgfs[i] = spgf
	}

	return spgfs
}

func expandApplicationSetStrategy(sp map[string]interface{}) (*application.ApplicationSetStrategy, error) {
	s := &application.ApplicationSetStrategy{
		Type: sp["type"].(string),
	}

	if v, ok := sp["rolling_sync"].([]interface{}); ok && len(v) > 0 {
		rs, err := expandApplicationSetRolloutStrategy(v[0].(map[string]interface{}))
		if err != nil {
			return nil, err
		}

		s.RollingSync = rs
	}

	return s, nil
}

func expandApplicationSetRolloutStrategy(rs map[string]interface{}) (*application.ApplicationSetRolloutStrategy, error) {
	asrs := &application.ApplicationSetRolloutStrategy{}

	if s, ok := rs["step"].([]interface{}); ok && len(s) > 0 {
		ss, err := expandApplicationSetRolloutSteps(s)
		if err != nil {
			return nil, err
		}

		asrs.Steps = ss
	}

	return asrs, nil
}

func expandApplicationSetRolloutSteps(rss []interface{}) ([]application.ApplicationSetRolloutStep, error) {
	if len(rss) == 0 || rss[0] == nil {
		return []application.ApplicationSetRolloutStep{}, nil
	}

	asrss := make([]application.ApplicationSetRolloutStep, len(rss))

	for i, rs := range rss {
		rs := rs.(map[string]interface{})

		asrs := application.ApplicationSetRolloutStep{}

		if v, ok := rs["match_expressions"].([]interface{}); ok && len(v) > 0 {
			asrs.MatchExpressions = expandApplicationMatchExpressions(v)
		}

		if v, ok := rs["max_update"]; ok {
			mu, err := expandIntOrString(v.(string))
			if err != nil {
				return nil, fmt.Errorf("could not expand max_update: %w", err)
			}

			asrs.MaxUpdate = mu
		}

		asrss[i] = asrs
	}

	return asrss, nil
}

func expandApplicationMatchExpressions(mes []interface{}) []application.ApplicationMatchExpression {
	asrss := make([]application.ApplicationMatchExpression, len(mes))

	for i, me := range mes {
		me := me.(map[string]interface{})
		asrss[i] = application.ApplicationMatchExpression{
			Key:      me["key"].(string),
			Operator: me["operator"].(string),
			Values:   sliceOfString(me["values"].(*schema.Set).List()),
		}
	}

	return asrss
}

func expandApplicationSetSyncPolicy(sp map[string]interface{}) *application.ApplicationSetSyncPolicy {
	return &application.ApplicationSetSyncPolicy{
		PreserveResourcesOnDeletion: sp["preserve_resources_on_deletion"].(bool),
	}
}

func expandApplicationSetTemplate(temp interface{}, featureMultipleApplicationSourcesSupported bool) (template application.ApplicationSetTemplate, err error) {
	t, ok := temp.(map[string]interface{})
	if !ok {
		return template, fmt.Errorf("could not expand application set template")
	}

	if v, ok := t["metadata"]; ok {
		template.ApplicationSetTemplateMeta, err = expandApplicationSetTemplateMeta(v.([]interface{})[0])
		if err != nil {
			return
		}
	}

	if v, ok := t["spec"]; ok {
		s := v.([]interface{})[0].(map[string]interface{})

		template.Spec, err = expandApplicationSpec(s)
		if err != nil {
			return
		}

		l := len(template.Spec.Sources)

		switch {
		case l == 1:
			template.Spec.Source = &template.Spec.Sources[0]
			template.Spec.Sources = nil
		case l > 1 && !featureMultipleApplicationSourcesSupported:
			f := features.ConstraintsMap[features.MultipleApplicationSources]
			return template, fmt.Errorf("%s is only supported from ArgoCD %s onwards", f.Name, f.MinVersion.String())
		}
	}

	return template, nil
}

func expandApplicationSetTemplateMeta(meta interface{}) (metadata application.ApplicationSetTemplateMeta, err error) {
	if meta == nil {
		return
	}

	m, ok := meta.(map[string]interface{})
	if !ok {
		return metadata, fmt.Errorf("could not expand application set template metadata")
	}

	if v, ok := m["annotations"].(map[string]interface{}); ok && len(v) > 0 {
		metadata.Annotations = expandStringMap(v)
	}

	if v, ok := m["labels"].(map[string]interface{}); ok && len(v) > 0 {
		metadata.Labels = expandStringMap(v)
	}

	if v, ok := m["name"]; ok {
		metadata.Name = v.(string)
	}

	if v, ok := m["namespace"]; ok {
		metadata.Namespace = v.(string)
	}

	if v, ok := m["finalizers"].([]interface{}); ok && len(v) > 0 {
		metadata.Finalizers = expandStringList(v)
	}

	return metadata, nil
}

func flattenApplicationSet(as *application.ApplicationSet, d *schema.ResourceData) error {
	fMetadata := flattenMetadata(as.ObjectMeta, d)
	if err := d.Set("metadata", fMetadata); err != nil {
		e, _ := json.MarshalIndent(fMetadata, "", "\t")
		return fmt.Errorf("error persisting metadata: %s\n%s", err, e)
	}

	fSpec, err := flattenApplicationSetSpec(as.Spec)
	if err != nil {
		return err
	}

	if err := d.Set("spec", fSpec); err != nil {
		e, _ := json.MarshalIndent(fSpec, "", "\t")
		return fmt.Errorf("error persisting spec: %s\n%s", err, e)
	}

	return nil
}

func flattenApplicationSetSpec(s application.ApplicationSetSpec) ([]map[string]interface{}, error) {
	generators := make([]interface{}, len(s.Generators))

	for i, g := range s.Generators {
		generator, err := flattenGenerator(g)
		if err != nil {
			return nil, err
		}

		generators[i] = generator
	}

	spec := map[string]interface{}{
		"generator":   generators,
		"go_template": s.GoTemplate,
		"template":    flattenApplicationSetTemplate(s.Template),
	}

	if s.Strategy != nil {
		spec["strategy"] = flattenApplicationSetStrategy(*s.Strategy)
	}

	if s.SyncPolicy != nil {
		spec["sync_policy"] = flattenApplicationSetSyncPolicy(*s.SyncPolicy)
	}

	return []map[string]interface{}{spec}, nil
}

func flattenGenerator(g application.ApplicationSetGenerator) (map[string]interface{}, error) {
	generator := map[string]interface{}{}

	if g.Clusters != nil {
		generator["clusters"] = flattenApplicationSetClusterGenerator(g.Clusters)
	} else if g.ClusterDecisionResource != nil {
		generator["cluster_decision_resource"] = flattenApplicationSetClusterDecisionResourceGenerator(g.ClusterDecisionResource)
	} else if g.Git != nil {
		generator["git"] = flattenApplicationSetGitGenerator(g.Git)
	} else if g.List != nil {
		list, err := flattenApplicationSetListGenerator(g.List)
		if err != nil {
			return nil, err
		}

		generator["list"] = list
	} else if g.Matrix != nil {
		matrix, err := flattenApplicationSetMatrixGenerator(g.Matrix)
		if err != nil {
			return nil, err
		}

		generator["matrix"] = matrix
	} else if g.Merge != nil {
		matrix, err := flattenApplicationSetMergeGenerator(g.Merge)
		if err != nil {
			return nil, err
		}

		generator["merge"] = matrix
	} else if g.SCMProvider != nil {
		generator["scm_provider"] = flattenApplicationSetSCMProviderGenerator(g.SCMProvider)
	} else if g.PullRequest != nil {
		generator["pull_request"] = flattenApplicationSetPullRequestGenerator(g.PullRequest)
	}

	if g.Selector != nil {
		generator["selector"] = flattenLabelSelector(g.Selector)
	}

	return generator, nil
}

func flattenApplicationSetClusterGenerator(c *application.ClusterGenerator) []map[string]interface{} {
	g := map[string]interface{}{
		"enabled":  true,
		"selector": flattenLabelSelector(&c.Selector),
		"template": flattenApplicationSetTemplate(c.Template),
		"values":   c.Values,
	}

	return []map[string]interface{}{g}
}

func flattenApplicationSetClusterDecisionResourceGenerator(c *application.DuckTypeGenerator) []map[string]interface{} {
	g := map[string]interface{}{
		"config_map_ref": c.ConfigMapRef,
		"label_selector": flattenLabelSelector(&c.LabelSelector),
		"name":           c.Name,
		"template":       flattenApplicationSetTemplate(c.Template),
		"values":         c.Values,
	}

	if c.RequeueAfterSeconds != nil {
		g["requeue_after_seconds"] = convertInt64PointerToString(c.RequeueAfterSeconds)
	}

	return []map[string]interface{}{g}
}

func flattenApplicationSetGitGenerator(gg *application.GitGenerator) []map[string]interface{} {
	g := map[string]interface{}{
		"repo_url": gg.RepoURL,
		"revision": gg.Revision,
		"template": flattenApplicationSetTemplate(gg.Template),
	}

	if len(gg.Directories) > 0 {
		directories := make([]map[string]interface{}, len(gg.Directories))
		for i, d := range gg.Directories {
			directories[i] = map[string]interface{}{
				"path":    d.Path,
				"exclude": d.Exclude,
			}
		}

		g["directory"] = directories
	}

	if len(gg.Files) > 0 {
		files := make([]map[string]interface{}, len(gg.Files))
		for i, f := range gg.Files {
			files[i] = map[string]interface{}{
				"path": f.Path,
			}
		}

		g["file"] = files
	}

	return []map[string]interface{}{g}
}

func flattenApplicationSetListGenerator(lg *application.ListGenerator) ([]map[string]interface{}, error) {
	elements := make([]interface{}, len(lg.Elements))

	for i, e := range lg.Elements {
		element := make(map[string]interface{})

		err := json.Unmarshal(e.Raw, &element)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal list generator element: %w", err)
		}

		elements[i] = element
	}

	g := map[string]interface{}{
		"elements": elements,
		"template": flattenApplicationSetTemplate(lg.Template),
	}

	return []map[string]interface{}{g}, nil
}

func flattenApplicationSetMatrixGenerator(mg *application.MatrixGenerator) ([]map[string]interface{}, error) {
	generators := make([]interface{}, len(mg.Generators))

	for i, g := range mg.Generators {
		fg, err := flattenNestedGenerator(g)
		if err != nil {
			return nil, err
		}

		generators[i] = fg
	}

	g := map[string]interface{}{
		"generator": generators,
		"template":  flattenApplicationSetTemplate(mg.Template),
	}

	return []map[string]interface{}{g}, nil
}

func flattenApplicationSetMergeGenerator(mg *application.MergeGenerator) ([]map[string]interface{}, error) {
	generators := make([]interface{}, len(mg.Generators))

	for i, g := range mg.Generators {
		fg, err := flattenNestedGenerator(g)
		if err != nil {
			return nil, err
		}

		generators[i] = fg
	}

	g := map[string]interface{}{
		"merge_keys": mg.MergeKeys,
		"generator":  generators,
		"template":   flattenApplicationSetTemplate(mg.Template),
	}

	return []map[string]interface{}{g}, nil
}

func flattenApplicationSetPullRequestGenerator(prg *application.PullRequestGenerator) []map[string]interface{} {
	g := map[string]interface{}{}

	if prg.BitbucketServer != nil {
		g["bitbucket_server"] = flattenApplicationSetPullRequestGeneratorBitbucketServer(prg.BitbucketServer)
	} else if prg.Gitea != nil {
		g["gitea"] = flattenApplicationSetPullRequestGeneratorGitea(prg.Gitea)
	} else if prg.Github != nil {
		g["github"] = flattenApplicationSetPullRequestGeneratorGithub(prg.Github)
	} else if prg.GitLab != nil {
		g["gitlab"] = flattenApplicationSetPullRequestGeneratorGitlab(prg.GitLab)
	}

	if len(prg.Filters) > 0 {
		g["filter"] = flattenApplicationSetPullRequestGeneratorFilter(prg.Filters)
	}

	if prg.RequeueAfterSeconds != nil {
		g["requeue_after_seconds"] = convertInt64PointerToString(prg.RequeueAfterSeconds)
	}

	g["template"] = flattenApplicationSetTemplate(prg.Template)

	return []map[string]interface{}{g}
}

func flattenApplicationSetPullRequestGeneratorBitbucketServer(prgbs *application.PullRequestGeneratorBitbucketServer) []map[string]interface{} {
	bb := map[string]interface{}{
		"api":     prgbs.API,
		"project": prgbs.Project,
		"repo":    prgbs.Repo,
	}

	if prgbs.BasicAuth != nil {
		ba := map[string]interface{}{
			"username": prgbs.BasicAuth.Username,
		}

		if prgbs.BasicAuth.PasswordRef != nil {
			ba["password_ref"] = flattenSecretRef(*prgbs.BasicAuth.PasswordRef)
		}

		bb["basic_auth"] = []map[string]interface{}{ba}
	}

	return []map[string]interface{}{bb}
}

func flattenApplicationSetPullRequestGeneratorGitea(prgg *application.PullRequestGeneratorGitea) []map[string]interface{} {
	g := map[string]interface{}{
		"api":      prgg.API,
		"insecure": prgg.Insecure,
		"owner":    prgg.Owner,
		"repo":     prgg.Repo,
	}

	if prgg.TokenRef != nil {
		g["token_ref"] = flattenSecretRef(*prgg.TokenRef)
	}

	return []map[string]interface{}{g}
}

func flattenApplicationSetPullRequestGeneratorGithub(prgg *application.PullRequestGeneratorGithub) []map[string]interface{} {
	g := map[string]interface{}{
		"api":             prgg.API,
		"app_secret_name": prgg.AppSecretName,
		"owner":           prgg.Owner,
		"repo":            prgg.Repo,
	}

	if len(prgg.Labels) > 0 {
		g["labels"] = prgg.Labels
	}

	if prgg.TokenRef != nil {
		g["token_ref"] = flattenSecretRef(*prgg.TokenRef)
	}

	return []map[string]interface{}{g}
}

func flattenApplicationSetPullRequestGeneratorGitlab(prgg *application.PullRequestGeneratorGitLab) []map[string]interface{} {
	g := map[string]interface{}{
		"api":                prgg.API,
		"project":            prgg.Project,
		"pull_request_state": prgg.PullRequestState,
	}

	if len(prgg.Labels) > 0 {
		g["labels"] = prgg.Labels
	}

	if prgg.TokenRef != nil {
		g["token_ref"] = flattenSecretRef(*prgg.TokenRef)
	}

	return []map[string]interface{}{g}
}

func flattenApplicationSetPullRequestGeneratorFilter(spgfs []application.PullRequestGeneratorFilter) []map[string]interface{} {
	fs := make([]map[string]interface{}, len(spgfs))

	for i, v := range spgfs {
		fs[i] = map[string]interface{}{}

		if v.BranchMatch != nil {
			fs[i]["branch_match"] = *v.BranchMatch
		}
	}

	return fs
}

func flattenApplicationSetSCMProviderGenerator(spg *application.SCMProviderGenerator) []map[string]interface{} {
	g := map[string]interface{}{
		"clone_protocol": spg.CloneProtocol,
	}

	if spg.AzureDevOps != nil {
		g["azure_devops"] = flattenApplicationSetSCMProviderGeneratorAzureDevOps(spg.AzureDevOps)
	} else if spg.Bitbucket != nil {
		g["bitbucket_cloud"] = flattenApplicationSetSCMProviderGeneratorBitbucket(spg.Bitbucket)
	} else if spg.BitbucketServer != nil {
		g["bitbucket_server"] = flattenApplicationSetSCMProviderGeneratorBitbucketServer(spg.BitbucketServer)
	} else if spg.Gitea != nil {
		g["gitea"] = flattenApplicationSetSCMProviderGeneratorGitea(spg.Gitea)
	} else if spg.Github != nil {
		g["github"] = flattenApplicationSetSCMProviderGeneratorGithub(spg.Github)
	} else if spg.Gitlab != nil {
		g["gitlab"] = flattenApplicationSetSCMProviderGeneratorGitlab(spg.Gitlab)
	}

	if len(spg.Filters) > 0 {
		g["filter"] = flattenApplicationSetSCMProviderGeneratorFilter(spg.Filters)
	}

	if spg.RequeueAfterSeconds != nil {
		g["requeue_after_seconds"] = convertInt64PointerToString(spg.RequeueAfterSeconds)
	}

	g["template"] = flattenApplicationSetTemplate(spg.Template)

	return []map[string]interface{}{g}
}

func flattenApplicationSetSCMProviderGeneratorAzureDevOps(spgado *application.SCMProviderGeneratorAzureDevOps) []map[string]interface{} {
	a := map[string]interface{}{
		"all_branches": spgado.AllBranches,
		"api":          spgado.API,
		"organization": spgado.Organization,
		"team_project": spgado.TeamProject,
	}

	if spgado.AccessTokenRef != nil {
		a["access_token_ref"] = flattenSecretRef(*spgado.AccessTokenRef)
	}

	return []map[string]interface{}{a}
}

func flattenApplicationSetSCMProviderGeneratorBitbucket(spgb *application.SCMProviderGeneratorBitbucket) []map[string]interface{} {
	bb := map[string]interface{}{
		"all_branches": spgb.AllBranches,
		"owner":        spgb.Owner,
		"user":         spgb.User,
	}

	if spgb.AppPasswordRef != nil {
		bb["app_password_ref"] = flattenSecretRef(*spgb.AppPasswordRef)
	}

	return []map[string]interface{}{bb}
}

func flattenApplicationSetSCMProviderGeneratorBitbucketServer(spgbs *application.SCMProviderGeneratorBitbucketServer) []map[string]interface{} {
	bb := map[string]interface{}{
		"all_branches": spgbs.AllBranches,
		"api":          spgbs.API,
		"project":      spgbs.Project,
	}

	if spgbs.BasicAuth != nil {
		ba := map[string]interface{}{
			"username": spgbs.BasicAuth.Username,
		}

		if spgbs.BasicAuth.PasswordRef != nil {
			ba["password_ref"] = flattenSecretRef(*spgbs.BasicAuth.PasswordRef)
		}

		bb["basic_auth"] = []map[string]interface{}{ba}
	}

	return []map[string]interface{}{bb}
}

func flattenApplicationSetSCMProviderGeneratorGitea(spgg *application.SCMProviderGeneratorGitea) []map[string]interface{} {
	g := map[string]interface{}{
		"all_branches": spgg.AllBranches,
		"api":          spgg.API,
		"insecure":     spgg.Insecure,
		"owner":        spgg.Owner,
	}

	if spgg.TokenRef != nil {
		g["token_ref"] = flattenSecretRef(*spgg.TokenRef)
	}

	return []map[string]interface{}{g}
}

func flattenApplicationSetSCMProviderGeneratorGithub(spgg *application.SCMProviderGeneratorGithub) []map[string]interface{} {
	g := map[string]interface{}{
		"all_branches":    spgg.AllBranches,
		"api":             spgg.API,
		"app_secret_name": spgg.AppSecretName,
		"organization":    spgg.Organization,
	}

	if spgg.TokenRef != nil {
		g["token_ref"] = flattenSecretRef(*spgg.TokenRef)
	}

	return []map[string]interface{}{g}
}

func flattenApplicationSetSCMProviderGeneratorGitlab(spgg *application.SCMProviderGeneratorGitlab) []map[string]interface{} {
	g := map[string]interface{}{
		"all_branches":      spgg.AllBranches,
		"api":               spgg.API,
		"group":             spgg.Group,
		"include_subgroups": spgg.IncludeSubgroups,
	}

	if spgg.TokenRef != nil {
		g["token_ref"] = flattenSecretRef(*spgg.TokenRef)
	}

	return []map[string]interface{}{g}
}

func flattenApplicationSetSCMProviderGeneratorFilter(spgfs []application.SCMProviderGeneratorFilter) []map[string]interface{} {
	fs := make([]map[string]interface{}, len(spgfs))

	for i, v := range spgfs {
		fs[i] = map[string]interface{}{}

		if v.BranchMatch != nil {
			fs[i]["branch_match"] = *v.BranchMatch
		}

		if v.LabelMatch != nil {
			fs[i]["label_match"] = *v.LabelMatch
		}

		if len(v.PathsDoNotExist) > 0 {
			fs[i]["paths_do_not_exist"] = v.PathsDoNotExist
		}

		if len(v.PathsExist) > 0 {
			fs[i]["paths_exist"] = v.PathsExist
		}

		if v.RepositoryMatch != nil {
			fs[i]["repository_match"] = *v.RepositoryMatch
		}
	}

	return fs
}

func flattenNestedGenerator(g application.ApplicationSetNestedGenerator) (map[string]interface{}, error) {
	generator := map[string]interface{}{}

	if g.Clusters != nil {
		generator["clusters"] = flattenApplicationSetClusterGenerator(g.Clusters)
	} else if g.ClusterDecisionResource != nil {
		generator["cluster_decision_resource"] = flattenApplicationSetClusterDecisionResourceGenerator(g.ClusterDecisionResource)
	} else if g.Git != nil {
		generator["git"] = flattenApplicationSetGitGenerator(g.Git)
	} else if g.List != nil {
		list, err := flattenApplicationSetListGenerator(g.List)
		if err != nil {
			return nil, err
		}

		generator["list"] = list
	} else if g.Matrix != nil {
		mg, err := application.ToNestedMatrixGenerator(g.Matrix)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal nested matrix generator: %w", err)
		}

		matrix, err := flattenApplicationSetMatrixGenerator(mg.ToMatrixGenerator())
		if err != nil {
			return nil, err
		}

		generator["matrix"] = matrix
	} else if g.Merge != nil {
		mg, err := application.ToNestedMergeGenerator(g.Merge)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal nested matrix generator: %w", err)
		}

		merge, err := flattenApplicationSetMergeGenerator(mg.ToMergeGenerator())
		if err != nil {
			return nil, err
		}

		generator["merge"] = merge
	} else if g.SCMProvider != nil {
		generator["scm_provider"] = flattenApplicationSetSCMProviderGenerator(g.SCMProvider)
	} else if g.PullRequest != nil {
		generator["pull_request"] = flattenApplicationSetPullRequestGenerator(g.PullRequest)
	}

	if g.Selector != nil {
		generator["selector"] = flattenLabelSelector(g.Selector)
	}

	return generator, nil
}

func flattenApplicationSetStrategy(ass application.ApplicationSetStrategy) []map[string]interface{} {
	p := map[string]interface{}{
		"type": ass.Type,
	}

	if ass.RollingSync != nil {
		p["rolling_sync"] = flattenApplicationSetRolloutStrategy(*ass.RollingSync)
	}

	return []map[string]interface{}{p}
}

func flattenApplicationSetRolloutStrategy(asrs application.ApplicationSetRolloutStrategy) []map[string]interface{} {
	rs := map[string]interface{}{
		"step": flattenApplicationSetRolloutSteps(asrs.Steps),
	}

	return []map[string]interface{}{rs}
}

func flattenApplicationSetRolloutSteps(asrss []application.ApplicationSetRolloutStep) []map[string]interface{} {
	rss := make([]map[string]interface{}, len(asrss))

	for i, s := range asrss {
		rss[i] = map[string]interface{}{
			"match_expressions": flattenApplicationMatchExpression(s.MatchExpressions),
		}

		if s.MaxUpdate != nil {
			rss[i]["max_update"] = flattenIntOrString(s.MaxUpdate)
		}
	}

	return rss
}

func flattenApplicationMatchExpression(in []application.ApplicationMatchExpression) []map[string]interface{} {
	me := make([]map[string]interface{}, len(in))

	for i, n := range in {
		me[i] = map[string]interface{}{
			"key":      n.Key,
			"operator": n.Operator,
			"values":   newStringSet(schema.HashString, n.Values),
		}
	}

	return me
}

func flattenApplicationSetSyncPolicy(assp application.ApplicationSetSyncPolicy) []map[string]interface{} {
	p := map[string]interface{}{
		"preserve_resources_on_deletion": assp.PreserveResourcesOnDeletion,
	}

	return []map[string]interface{}{p}
}

func flattenApplicationSetTemplate(ast application.ApplicationSetTemplate) []map[string]interface{} {
	// Hack: Prior to ArgoCD 2.6.3, `Source` was not a pointer and as such a
	// zero value would be returned. However, this "zero" value means that the
	// `Template` is considered as non-zero in newer versions because the
	// pointer contains an object. To support versions of ArgoCD prior to 2.6.3,
	// we need to explicitly set the pointer to nil.
	if ast.Spec.Source != nil && ast.Spec.Source.IsZero() {
		ast.Spec.Source = nil
	}

	if reflect.ValueOf(ast).IsZero() {
		return nil
	}

	t := map[string]interface{}{
		"metadata": flattenApplicationSetTemplateMetadata(ast.ApplicationSetTemplateMeta),
		"spec":     flattenApplicationSpec(ast.Spec),
	}

	return []map[string]interface{}{t}
}

func flattenApplicationSetTemplateMetadata(tm application.ApplicationSetTemplateMeta) []map[string]interface{} {
	m := map[string]interface{}{
		"annotations": tm.Annotations,
		"finalizers":  tm.Finalizers,
		"labels":      tm.Labels,
		"name":        tm.Name,
		"namespace":   tm.Namespace,
	}

	return []map[string]interface{}{m}
}

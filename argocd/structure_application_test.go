package argocd

import (
	"testing"

	application "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
)

func TestFlattenApplicationSpecSourcesPrecedenceOverSource(t *testing.T) {
	t.Parallel()

	// ArgoCD ignores `spec.source` whenever `spec.sources` is non-empty (see
	// (*ApplicationSpec).HasMultipleSources() upstream), but a live Application object can
	// still have both fields populated, e.g. after import. Flattening must reflect the
	// sources ArgoCD actually syncs.
	spec := application.ApplicationSpec{
		Source: &application.ApplicationSource{
			RepoURL: "https://github.com/bogus/single-source.git",
			Path:    "single",
		},
		Sources: application.ApplicationSources{
			{RepoURL: "https://github.com/bogus/multi-source-a.git", Path: "multi-source-a"},
			{RepoURL: "https://github.com/bogus/multi-source-b.git", Path: "multi-source-b"},
		},
	}

	flattened := flattenApplicationSpec(spec)
	if len(flattened) != 1 {
		t.Fatalf("expected exactly one flattened spec, got %d", len(flattened))
	}

	sources, ok := flattened[0]["source"].([]map[string]interface{})
	if !ok {
		t.Fatalf("expected source to be []map[string]interface{}, got %T", flattened[0]["source"])
	}

	if len(sources) != 2 {
		t.Fatalf("expected spec.sources (2 entries) to take precedence over spec.source, got %d entries: %v", len(sources), sources)
	}

	if sources[0]["repo_url"] != "https://github.com/bogus/multi-source-a.git" {
		t.Errorf("expected first source repo_url %q, got %q", "https://github.com/bogus/multi-source-a.git", sources[0]["repo_url"])
	}

	if sources[1]["repo_url"] != "https://github.com/bogus/multi-source-b.git" {
		t.Errorf("expected second source repo_url %q, got %q", "https://github.com/bogus/multi-source-b.git", sources[1]["repo_url"])
	}
}

func TestFlattenApplicationSpecFallsBackToSingularSource(t *testing.T) {
	t.Parallel()

	spec := application.ApplicationSpec{
		Source: &application.ApplicationSource{
			RepoURL: "https://github.com/bogus/single-source.git",
			Path:    "single",
		},
	}

	flattened := flattenApplicationSpec(spec)

	sources, ok := flattened[0]["source"].([]map[string]interface{})
	if !ok {
		t.Fatalf("expected source to be []map[string]interface{}, got %T", flattened[0]["source"])
	}

	if len(sources) != 1 {
		t.Fatalf("expected exactly one source, got %d", len(sources))
	}

	if sources[0]["repo_url"] != "https://github.com/bogus/single-source.git" {
		t.Errorf("expected source repo_url %q, got %q", "https://github.com/bogus/single-source.git", sources[0]["repo_url"])
	}
}

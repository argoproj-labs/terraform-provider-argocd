package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExpandProject_OptionalDestinationAndSourceRepos ensures that an
// argocd_project can be expanded without any "destination" blocks or
// "source_repos" entries, e.g. for use as a global project to inherit
// settings from (see https://github.com/argoproj-labs/terraform-provider-argocd/issues/623).
func TestExpandProject_OptionalDestinationAndSourceRepos(t *testing.T) {
	t.Parallel()

	data := &projectModel{
		Metadata: []objectMeta{
			{
				Name: types.StringValue("global"),
			},
		},
		Spec: []projectSpecModel{
			{
				Description: types.StringValue("global project with no destinations or source repos"),
			},
		},
	}

	objectMeta, spec, diags := expandProject(t.Context(), data)

	require.False(t, diags.HasError(), "expandProject should not error when destination and source_repos are omitted")
	assert.Equal(t, "global", objectMeta.Name)
	assert.Empty(t, spec.Destinations)
	assert.Empty(t, spec.SourceRepos)
}

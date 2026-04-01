package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestMetadataAnnotationsValidator(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		val         types.Map
		expectError bool
	}{
		"null map": {
			val: types.MapNull(types.StringType),
		},
		"unknown map": {
			val: types.MapUnknown(types.StringType),
		},
		"valid annotation key": {
			val: types.MapValueMust(types.StringType, map[string]attr.Value{
				"app.kubernetes.io/name": types.StringValue("myapp"),
			}),
		},
		"multiple valid keys": {
			val: types.MapValueMust(types.StringType, map[string]attr.Value{
				"app.kubernetes.io/name":    types.StringValue("myapp"),
				"this.is.a.valid.key/stuff": types.StringValue("value"),
			}),
		},
		"unknown element value": {
			val: types.MapValueMust(types.StringType, map[string]attr.Value{
				"app.kubernetes.io/name": types.StringUnknown(),
			}),
		},
		"mixed known and unknown values": {
			val: types.MapValueMust(types.StringType, map[string]attr.Value{
				"app.kubernetes.io/name":    types.StringValue("myapp"),
				"app.kubernetes.io/version": types.StringUnknown(),
			}),
		},
		"invalid annotation key": {
			val: types.MapValueMust(types.StringType, map[string]attr.Value{
				"-invalid": types.StringValue("value"),
			}),
			expectError: true,
		},
		"uppercase annotation key rejected": {
			val: types.MapValueMust(types.StringType, map[string]attr.Value{
				"App.Kubernetes.IO/Name": types.StringValue("myapp"),
			}),
			expectError: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := validator.MapRequest{
				Path:        path.Root("annotations"),
				ConfigValue: test.val,
			}

			resp := validator.MapResponse{}
			MetadataAnnotations().ValidateMap(context.Background(), req, &resp)
			assert.Equal(t, test.expectError, resp.Diagnostics.HasError())
		})
	}
}

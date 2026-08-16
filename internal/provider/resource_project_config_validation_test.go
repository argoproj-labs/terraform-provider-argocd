package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProjectResourceValidateConfigPolicies exercises the project resource's
// config validators against a synthetic configuration. Unlike the acceptance
// tests it proves the policy is rejected during validation rather than by the
// Argo CD API on apply, which is the whole point of validating client side.
func TestProjectResourceValidateConfigPolicies(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		policy      string
		expectError bool
	}{
		"valid policy": {
			policy: "p, proj:myproject:admin, applications, get, myproject/*, allow",
		},
		"unrecognised resource": {
			policy:      "p, proj:myproject:admin, applicat, get, myproject/*, allow",
			expectError: true,
		},
		"resource that is not project scoped": {
			policy:      "p, proj:myproject:admin, certificates, get, myproject/*, allow",
			expectError: true,
		},
		"subject does not match role": {
			policy:      "p, proj:myproject:bar, applications, get, myproject/*, allow",
			expectError: true,
		},
		"invalid effect": {
			policy:      "p, proj:myproject:admin, applications, get, myproject/*, whatever",
			expectError: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := validateProjectConfig(t, "myproject", "admin", test.policy)

			if test.expectError {
				assert.True(t, diags.HasError(), "expected a validation error, got none")
			} else {
				assert.False(t, diags.HasError(), "unexpected validation error: %v", diags.Errors())
			}
		})
	}
}

// TestProjectResourceValidateConfigUnknownProjectName ensures an unknown project
// name short circuits validation instead of reporting a bogus error, since both
// the expected subject and object are derived from it.
func TestProjectResourceValidateConfigUnknownProjectName(t *testing.T) {
	t.Parallel()

	diags := validateProjectConfig(t, tftypes.UnknownValue, "admin", "p, proj:myproject:admin, applications, get, myproject/*, allow")
	assert.False(t, diags.HasError(), "unexpected validation error: %v", diags.Errors())
}

func validateProjectConfig(t *testing.T, projectName any, roleName string, policy string) diag.Diagnostics {
	t.Helper()

	ctx := context.Background()

	r := &projectResource{}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	s := schemaResp.Schema

	objectType, ok := s.Type().TerraformType(ctx).(tftypes.Object)
	require.True(t, ok)

	metadataType, ok := objectType.AttributeTypes["metadata"].(tftypes.List)
	require.True(t, ok)

	metadataElement, ok := metadataType.ElementType.(tftypes.Object)
	require.True(t, ok)

	specType, ok := objectType.AttributeTypes["spec"].(tftypes.List)
	require.True(t, ok)

	specElement, ok := specType.ElementType.(tftypes.Object)
	require.True(t, ok)

	roleType, ok := specElement.AttributeTypes["role"].(tftypes.Set)
	require.True(t, ok)

	roleElement, ok := roleType.ElementType.(tftypes.Object)
	require.True(t, ok)

	policiesType, ok := roleElement.AttributeTypes["policies"].(tftypes.List)
	require.True(t, ok)

	metadata := tftypes.NewValue(metadataType, []tftypes.Value{
		objectWithNullDefaults(metadataElement, map[string]tftypes.Value{
			"name": tftypes.NewValue(tftypes.String, projectName),
		}),
	})

	role := objectWithNullDefaults(roleElement, map[string]tftypes.Value{
		"name": tftypes.NewValue(tftypes.String, roleName),
		"policies": tftypes.NewValue(policiesType, []tftypes.Value{
			tftypes.NewValue(tftypes.String, policy),
		}),
	})

	spec := tftypes.NewValue(specType, []tftypes.Value{
		objectWithNullDefaults(specElement, map[string]tftypes.Value{
			"role": tftypes.NewValue(roleType, []tftypes.Value{role}),
		}),
	})

	config := tfsdk.Config{
		Schema: s,
		Raw: objectWithNullDefaults(objectType, map[string]tftypes.Value{
			"metadata": metadata,
			"spec":     spec,
		}),
	}

	resp := &resource.ValidateConfigResponse{}

	for _, validator := range r.ConfigValidators(ctx) {
		validator.ValidateResource(ctx, resource.ValidateConfigRequest{Config: config}, resp)
	}

	return resp.Diagnostics
}

// objectWithNullDefaults builds an object value where every attribute not named
// in overrides is null, so a test only has to spell out what it cares about.
func objectWithNullDefaults(objectType tftypes.Object, overrides map[string]tftypes.Value) tftypes.Value {
	attributes := make(map[string]tftypes.Value, len(objectType.AttributeTypes))

	for name, attributeType := range objectType.AttributeTypes {
		if override, ok := overrides[name]; ok {
			attributes[name] = override
			continue
		}

		attributes[name] = tftypes.NewValue(attributeType, nil)
	}

	return tftypes.NewValue(objectType, attributes)
}

package argocd

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"testing"
)

func testResourceArgoCDProjectStateDataV0() map[string]interface{} {
	return map[string]interface{}{
		"spec": []map[string]interface{}{
			{
				"orphaned_resources": map[string]bool{"warn": true},
			},
		},
	}
}

func testResourceArgoCDProjectStateDataV1() map[string]interface{} {
	newOrphanedResources := schema.NewSet(
		schema.HashResource(&schema.Resource{
			Schema: map[string]*schema.Schema{
				"warn": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"ignore": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"group": {
								Type:         schema.TypeString,
								ValidateFunc: validateGroupName,
								Optional:     true,
							},
							"kind": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"name": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
			},
		}),
		[]interface{}{map[string]interface{}{"warn": true}},
	)
	return map[string]interface{}{
		"spec": []map[string]interface{}{
			{
				"orphaned_resources": newOrphanedResources,
			},
		},
	}
}

func TestResourceArgoCDProjectStateUpgradeV0(t *testing.T) {
	cases := []struct {
		name        string
		expected    map[string]interface{}
		sourceState map[string]interface{}
	}{
		{
			"source < v0.5.0",
			testResourceArgoCDProjectStateDataV1(),
			testResourceArgoCDProjectStateDataV0(),
		},
		{
			"source < v1.1.0, >= v0.5.0",
			testResourceArgoCDProjectStateDataV1(),
			testResourceArgoCDProjectStateDataV1(),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_actual, err := resourceArgoCDProjectStateUpgradeV0(tc.sourceState, nil)
			if err != nil {
				t.Fatalf("error migrating state: %s", err)
			}
			expected := tc.expected["spec"].([]map[string]interface{})[0]["orphaned_resources"].(*schema.Set)
			actual := _actual["spec"].([]map[string]interface{})[0]["orphaned_resources"].(*schema.Set)
			if !expected.HashEqual(actual) {
				t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
			}
		})
	}
}

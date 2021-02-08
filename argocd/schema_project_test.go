package argocd

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"reflect"
	"testing"
)

func orphanedResourcesSchemaSetFuncV1() schema.SchemaSetFunc {
	return schema.HashResource(&schema.Resource{
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
	})
}

func TestResourceArgoCDProjectStateUpgradeV0(t *testing.T) {
	type projectStateUpgradeTestCases []struct {
		name          string
		expectedState map[string]interface{}
		sourceState   map[string]interface{}
	}
	cases := projectStateUpgradeTestCases{
		{
			name: "source_<_v0.5.0_with_warn",
			sourceState: map[string]interface{}{
				"spec": []map[string]interface{}{
					{
						"orphaned_resources": map[string]bool{"warn": true},
					},
				},
			},
			expectedState: map[string]interface{}{
				"spec": []map[string]interface{}{
					{
						"orphaned_resources": schema.NewSet(
							orphanedResourcesSchemaSetFuncV1(),
							[]interface{}{map[string]interface{}{"warn": true}},
						),
					},
				},
			},
		},
		{
			name: "source_<_v0.5.0_without_orphaned_resources",
			sourceState: map[string]interface{}{
				"spec": []map[string]interface{}{
					{
						"source_repos": []string{"*"},
					},
				},
			},
			expectedState: map[string]interface{}{
				"spec": []map[string]interface{}{
					{
						"source_repos": []string{"*"},
					},
				},
			},
		},
		{
			name: "source_<_v0.5.0_with_empty_orphaned_resources",
			sourceState: map[string]interface{}{
				"spec": []map[string]interface{}{
					{
						"orphaned_resources": map[string]bool{},
					},
				},
			},
			expectedState: map[string]interface{}{
				"spec": []map[string]interface{}{
					{
						"orphaned_resources": schema.NewSet(
							orphanedResourcesSchemaSetFuncV1(),
							[]interface{}{map[string]interface{}{"warn": false}},
						),
					},
				},
			},
		},
		{
			name: "source_<_v1.1.0_>=_0.4.8_with_warn",
			sourceState: map[string]interface{}{
				"spec": []map[string]interface{}{
					{
						"cluster_resource_whitelist": []map[string]string{},
						"description":                "test",
						"destination": map[string]string{
							"namespace": "*",
							"server":    "https://testing.io",
						},
						"namespace_resource_blacklist": []map[string]string{},
						"orphaned_resources": map[string]bool{
							"warn": true,
						},
						"role":         []map[string]interface{}{},
						"source_repos": []string{"git@github.com:testing/test.git"},
						"sync_window":  []map[string]interface{}{},
					},
				},
			},
			expectedState: map[string]interface{}{
				"spec": []map[string]interface{}{
					{
						"cluster_resource_whitelist": []map[string]string{},
						"description":                "test",
						"destination": map[string]string{
							"namespace": "*",
							"server":    "https://testing.io",
						},
						"namespace_resource_blacklist": []map[string]string{},
						"orphaned_resources": schema.NewSet(
							orphanedResourcesSchemaSetFuncV1(),
							[]interface{}{map[string]interface{}{"warn": true}},
						),
						"role":         []map[string]interface{}{},
						"source_repos": []string{"git@github.com:testing/test.git"},
						"sync_window":  []map[string]interface{}{},
					},
				},
			},
		},
		{
			name: "source_<_v1.1.1_without_orphaned_resources",
			sourceState: map[string]interface{}{
				"spec": []map[string]interface{}{
					{
						"source_repos": []string{"*"},
					},
				},
			},
			expectedState: map[string]interface{}{
				"spec": []map[string]interface{}{
					{
						"source_repos": []string{"*"},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actualState, err := resourceArgoCDProjectStateUpgradeV0(tc.sourceState, nil)
			if err != nil {
				t.Fatalf("error migrating state: %s", err)
			}
			if !reflect.DeepEqual(actualState, tc.expectedState) {
				if expectedSet, ok := tc.expectedState["spec"].([]map[string]interface{})[0]["orphaned_resources"]; ok {

					actualSet := actualState["spec"].([]map[string]interface{})[0]["orphaned_resources"].(*schema.Set)

					if !expectedSet.(*schema.Set).HashEqual(actualSet) {
						t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expectedSet, actualSet)
					}
					// Cannot DeepEqual a pointer reference
					for k := range tc.expectedState["spec"].([]map[string]interface{})[0] {
						av := actualState["spec"].([]map[string]interface{})[0][k]
						ev := tc.expectedState["spec"].([]map[string]interface{})[0][k]
						if k != "orphaned_resources" && !reflect.DeepEqual(av, ev) {
							t.Fatalf("\n\n[maps] expected:\n\n%#v\n\ngot:\n\n%#v\n\n", tc.expectedState, actualState)
						}
					}
					for k, av := range actualState["spec"].([]map[string]interface{})[0] {
						ev := tc.expectedState["spec"].([]map[string]interface{})[0][k]
						if k != "orphaned_resources" && !reflect.DeepEqual(av, ev) {
							t.Fatalf("\n\n[maps] expected:\n\n%#v\n\ngot:\n\n%#v\n\n", tc.expectedState, actualState)
						}
					}
				} else {
					// Cannot DeepEqual a pointer reference
					for k := range tc.expectedState["spec"].([]map[string]interface{})[0] {
						av := actualState["spec"].([]map[string]interface{})[0][k]
						ev := tc.expectedState["spec"].([]map[string]interface{})[0][k]
						if k != "orphaned_resources" && !reflect.DeepEqual(av, ev) {
							t.Fatalf("\n\n[maps without set] expected:\n\n%#v\n\ngot:\n\n%#v\n\n", tc.expectedState, actualState)
						}
					}
					for k := range tc.sourceState["spec"].([]map[string]interface{})[0] {
						av := actualState["spec"].([]map[string]interface{})[0][k]
						ev := tc.expectedState["spec"].([]map[string]interface{})[0][k]
						if k != "orphaned_resources" && !reflect.DeepEqual(av, ev) {
							t.Fatalf("\n\n[maps] expected:\n\n%#v\n\ngot:\n\n%#v\n\n", tc.expectedState, actualState)
						}
					}
				}
			}

		})
	}
}

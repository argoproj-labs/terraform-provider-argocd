package argocd

import (
	"context"
	"reflect"
	"testing"
)

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
				"spec": []interface{}{
					map[string]interface{}{
						"orphaned_resources": map[string]bool{"warn": true},
					},
				},
			},
			expectedState: map[string]interface{}{
				"spec": []interface{}{
					map[string]interface{}{
						"orphaned_resources": []interface{}{map[string]bool{"warn": true}},
					},
				},
			},
		},
		{
			name: "source_<_v0.5.0_without_orphaned_resources",
			sourceState: map[string]interface{}{
				"spec": []interface{}{
					map[string]interface{}{
						"source_repos": []string{"*"},
					},
				},
			},
			expectedState: map[string]interface{}{
				"spec": []interface{}{
					map[string]interface{}{
						"source_repos": []string{"*"},
					},
				},
			},
		},
		{
			name: "source_<_v0.5.0_with_empty_orphaned_resources",
			sourceState: map[string]interface{}{
				"spec": []interface{}{
					map[string]interface{}{
						"orphaned_resources": map[string]bool{},
					},
				},
			},
			expectedState: map[string]interface{}{
				"spec": []interface{}{
					map[string]interface{}{
						"orphaned_resources": []interface{}{map[string]bool{"warn": false}},
					},
				},
			},
		},
		{
			name: "source_<_v1.1.0_>=_0.4.8_with_warn",
			sourceState: map[string]interface{}{
				"spec": []interface{}{
					map[string]interface{}{
						"cluster_resource_blacklist": []map[string]string{},
						"cluster_resource_whitelist": []map[string]string{},
						"description":                "test",
						"destination": map[string]string{
							"namespace": "*",
							"server":    "https://testing.io",
						},
						"namespace_resource_blacklist": []map[string]string{},
						"namespace_resource_whitelist": []map[string]string{},
						"orphaned_resources":           map[string]bool{"warn": true},
						"role":                         []map[string]interface{}{},
						"source_repos":                 []string{"git@github.com:testing/test.git"},
						"sync_window":                  []map[string]interface{}{},
					},
				},
			},
			expectedState: map[string]interface{}{
				"spec": []interface{}{
					map[string]interface{}{
						"cluster_resource_blacklist": []map[string]string{},
						"cluster_resource_whitelist": []map[string]string{},
						"description":                "test",
						"destination": map[string]string{
							"namespace": "*",
							"server":    "https://testing.io",
						},
						"namespace_resource_blacklist": []map[string]string{},
						"namespace_resource_whitelist": []map[string]string{},
						"orphaned_resources":           []interface{}{map[string]bool{"warn": true}},
						"role":                         []map[string]interface{}{},
						"source_repos":                 []string{"git@github.com:testing/test.git"},
						"sync_window":                  []map[string]interface{}{},
					},
				},
			},
		},
		{
			name: "source_<_v1.1.1_without_orphaned_resources",
			sourceState: map[string]interface{}{
				"spec": []interface{}{
					map[string]interface{}{
						"source_repos": []string{"*"},
					},
				},
			},
			expectedState: map[string]interface{}{
				"spec": []interface{}{
					map[string]interface{}{
						"source_repos": []string{"*"},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actualState, err := resourceArgoCDProjectStateUpgradeV0(context.TODO(), tc.sourceState, nil)
			if err != nil {
				t.Fatalf("error migrating state: %s", err)
			}
			if !reflect.DeepEqual(actualState, tc.expectedState) {
				t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", tc.expectedState, actualState)
			}
		})
	}
}

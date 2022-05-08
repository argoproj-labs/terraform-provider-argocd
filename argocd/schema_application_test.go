package argocd

import (
	"context"
	"reflect"
	"testing"
)

func TestUpgradeSchemaApplication_V0V1_Default_SkipCrds(t *testing.T) {
	v0 := map[string]interface{}{
		"spec": []interface{}{map[string]interface{}{
			"source": []interface{}{map[string]interface{}{
				"repo_url":        "https://charts.bitnami.com/bitnami",
				"chart":           "redis",
				"target_revision": "15.3.0",

				"helm": []interface{}{map[string]interface{}{
					"release_name": "testing",
				}},
			}},
			"destination": []interface{}{map[string]interface{}{
				"server":    "https://kubernetes.default.svc",
				"namespace": "default",
			}}},
		},
	}

	v1 := map[string]interface{}{
		"spec": []interface{}{map[string]interface{}{
			"source": []interface{}{map[string]interface{}{
				"repo_url":        "https://charts.bitnami.com/bitnami",
				"chart":           "redis",
				"target_revision": "15.3.0",

				"helm": []interface{}{map[string]interface{}{
					"release_name": "testing",
					"skip_crds":    false,
				}},
			}},
			"destination": []interface{}{map[string]interface{}{
				"server":    "https://kubernetes.default.svc",
				"namespace": "default",
			}}},
		},
	}

	actual, _ := resourceArgoCDApplicationStateUpgradeV0(context.TODO(), v0, nil)

	if !reflect.DeepEqual(v1, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", v1, actual)
	}
}

func TestUpgradeSchemaApplication_V0V1_Default_SkipCrds_NoChange(t *testing.T) {
	v0 := map[string]interface{}{
		"spec": []interface{}{map[string]interface{}{
			"source": []interface{}{map[string]interface{}{
				"repo_url":        "https://charts.bitnami.com/bitnami",
				"chart":           "redis",
				"target_revision": "15.3.0",
			}},
			"destination": []interface{}{map[string]interface{}{
				"server":    "https://kubernetes.default.svc",
				"namespace": "default",
			}}},
		},
	}

	actual, _ := resourceArgoCDApplicationStateUpgradeV0(context.TODO(), v0, nil)

	if !reflect.DeepEqual(v0, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", v0, actual)
	}
}

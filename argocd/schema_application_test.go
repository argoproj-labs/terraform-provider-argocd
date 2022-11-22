package argocd

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestUpgradeSchemaApplication_V0V1_Default_SkipCrds(t *testing.T) {
	v0 := map[string]interface{}{
		"spec": []interface{}{
			map[string]interface{}{
				"source": []interface{}{map[string]interface{}{
					"repo_url":        "https://charts.bitnami.com/bitnami",
					"chart":           "redis",
					"target_revision": "16.9.11",

					"helm": []interface{}{map[string]interface{}{
						"release_name": "testing",
					}},
				}},
				"destination": []interface{}{map[string]interface{}{
					"server":    "https://kubernetes.default.svc",
					"namespace": "default",
				}},
			},
		},
	}

	v1 := map[string]interface{}{
		"spec": []interface{}{
			map[string]interface{}{
				"source": []interface{}{map[string]interface{}{
					"repo_url":        "https://charts.bitnami.com/bitnami",
					"chart":           "redis",
					"target_revision": "16.9.11",

					"helm": []interface{}{map[string]interface{}{
						"release_name": "testing",
						"skip_crds":    false,
					}},
				}},
				"destination": []interface{}{map[string]interface{}{
					"server":    "https://kubernetes.default.svc",
					"namespace": "default",
				}},
			},
		},
	}

	actual, _ := resourceArgoCDApplicationStateUpgradeV0(context.TODO(), v0, nil)

	if !reflect.DeepEqual(v1, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", v1, actual)
	}
}

func TestUpgradeSchemaApplication_V0V1_Default_SkipCrds_NoChange(t *testing.T) {
	v0 := map[string]interface{}{
		"spec": []interface{}{
			map[string]interface{}{
				"source": []interface{}{map[string]interface{}{
					"repo_url":        "https://charts.bitnami.com/bitnami",
					"chart":           "redis",
					"target_revision": "16.9.11",
				}},
				"destination": []interface{}{map[string]interface{}{
					"server":    "https://kubernetes.default.svc",
					"namespace": "default",
				}},
			},
		},
	}

	actual, _ := resourceArgoCDApplicationStateUpgradeV0(context.TODO(), v0, nil)

	if !reflect.DeepEqual(v0, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", v0, actual)
	}
}

func TestUpgradeSchemaApplication_V1V2_Default_NoChange(t *testing.T) {
	v1 := map[string]interface{}{
		"spec": []interface{}{
			map[string]interface{}{
				"source": []interface{}{map[string]interface{}{
					"repo_url":        "https://charts.bitnami.com/bitnami",
					"chart":           "redis",
					"target_revision": "16.9.11",

					"helm": []interface{}{map[string]interface{}{
						"release_name": "testing",
						"skip_crds":    false,
					}},
				}},
				"destination": []interface{}{map[string]interface{}{
					"server":    "https://kubernetes.default.svc",
					"namespace": "default",
				}},
			},
		},
	}

	actual, _ := resourceArgoCDApplicationStateUpgradeV1(context.TODO(), v1, nil)

	if !reflect.DeepEqual(v1, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", v1, actual)
	}
}

func TestUpgradeSchemaApplication_V1V2_WithKsonnet(t *testing.T) {
	v1 := map[string]interface{}{
		"spec": []interface{}{
			map[string]interface{}{
				"source": []interface{}{map[string]interface{}{
					"repo_url":        "https://charts.bitnami.com/bitnami",
					"chart":           "redis",
					"target_revision": "16.9.11",

					"ksonnet": []interface{}{map[string]interface{}{
						"destination": []interface{}{map[string]interface{}{
							"namespace": "foo",
						}},
					}},
				}},
				"destination": []interface{}{map[string]interface{}{
					"server":    "https://kubernetes.default.svc",
					"namespace": "default",
				}},
			},
		},
	}

	_, err := resourceArgoCDApplicationStateUpgradeV1(context.TODO(), v1, nil)

	if err == nil || !strings.Contains(err.Error(), "'ksonnet' support has been removed") {
		t.Fatalf("\n\nexpected error during state migration was not found - err returned was: %v", err)
	}
}

func TestUpgradeSchemaApplication_V2V3_Default_NoChange(t *testing.T) {
	v2 := map[string]interface{}{
		"metadata": []interface{}{
			map[string]interface{}{
				"name":      "test",
				"namespace": "argocd",
			},
		},
		"spec": []interface{}{
			map[string]interface{}{
				"source": []interface{}{map[string]interface{}{
					"repo_url":        "https://charts.bitnami.com/bitnami",
					"chart":           "redis",
					"target_revision": "16.9.11",

					"helm": []interface{}{map[string]interface{}{
						"release_name": "testing",
						"skip_crds":    false,
					}},
				}},
				"destination": []interface{}{map[string]interface{}{
					"server":    "https://kubernetes.default.svc",
					"namespace": "default",
				}},
			},
		},
	}

	actual, _ := resourceArgoCDApplicationStateUpgradeV2(context.TODO(), v2, nil)

	if !reflect.DeepEqual(v2, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", v2, actual)
	}
}

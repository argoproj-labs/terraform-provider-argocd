package argocd

import (
	"reflect"
	"strings"
	"testing"
)

func TestUpgradeSchemaApplication_V0V1_Default_SkipCrds(t *testing.T) {
	t.Parallel()

	v0 := map[string]interface{}{
		"spec": []interface{}{
			map[string]interface{}{
				"source": []interface{}{map[string]interface{}{
					"repo_url":        "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami",
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
					"repo_url":        "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami",
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

	actual, _ := resourceArgoCDApplicationStateUpgradeV0(t.Context(), v0, nil)

	if !reflect.DeepEqual(v1, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", v1, actual)
	}
}

func TestUpgradeSchemaApplication_V0V1_Default_SkipCrds_NoChange(t *testing.T) {
	t.Parallel()

	v0 := map[string]interface{}{
		"spec": []interface{}{
			map[string]interface{}{
				"source": []interface{}{map[string]interface{}{
					"repo_url":        "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami",
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

	actual, _ := resourceArgoCDApplicationStateUpgradeV0(t.Context(), v0, nil)

	if !reflect.DeepEqual(v0, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", v0, actual)
	}
}

func TestUpgradeSchemaApplication_V1V2_Default_NoChange(t *testing.T) {
	t.Parallel()

	v1 := map[string]interface{}{
		"spec": []interface{}{
			map[string]interface{}{
				"source": []interface{}{map[string]interface{}{
					"repo_url":        "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami",
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

	actual, _ := resourceArgoCDApplicationStateUpgradeV1(t.Context(), v1, nil)

	if !reflect.DeepEqual(v1, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", v1, actual)
	}
}

func TestUpgradeSchemaApplication_V1V2_WithKsonnet(t *testing.T) {
	t.Parallel()

	v1 := map[string]interface{}{
		"spec": []interface{}{
			map[string]interface{}{
				"source": []interface{}{map[string]interface{}{
					"repo_url":        "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami",
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

	_, err := resourceArgoCDApplicationStateUpgradeV1(t.Context(), v1, nil)

	if err == nil || !strings.Contains(err.Error(), "'ksonnet' support has been removed") {
		t.Fatalf("\n\nexpected error during state migration was not found - err returned was: %v", err)
	}
}

func TestUpgradeSchemaApplication_V2V3_Default_NoChange(t *testing.T) {
	t.Parallel()

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
					"repo_url":        "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami",
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

	actual, _ := resourceArgoCDApplicationStateUpgradeV2(t.Context(), v2, nil)

	if !reflect.DeepEqual(v2, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", v2, actual)
	}
}
func TestUpgradeSchemaApplication_V3V4(t *testing.T) {
	t.Parallel()

	type stateUpgradeTestCases []struct {
		name          string
		sourceState   map[string]interface{}
		expectedState map[string]interface{}
	}

	cases := stateUpgradeTestCases{
		{
			name: "no sync policy",
			sourceState: map[string]interface{}{
				"metadata": []interface{}{
					map[string]interface{}{
						"name":      "test",
						"namespace": "argocd",
					},
				},
				"spec": []interface{}{
					map[string]interface{}{
						"source": []interface{}{map[string]interface{}{
							"repo_url":        "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami",
							"chart":           "redis",
							"target_revision": "16.9.11",
						}},
						"destination": []interface{}{map[string]interface{}{
							"server":    "https://kubernetes.default.svc",
							"namespace": "default",
						}},
					},
				},
			},
			expectedState: map[string]interface{}{
				"metadata": []interface{}{
					map[string]interface{}{
						"name":      "test",
						"namespace": "argocd",
					},
				},
				"spec": []interface{}{
					map[string]interface{}{
						"source": []interface{}{map[string]interface{}{
							"repo_url":        "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami",
							"chart":           "redis",
							"target_revision": "16.9.11",
						}},
						"destination": []interface{}{map[string]interface{}{
							"server":    "https://kubernetes.default.svc",
							"namespace": "default",
						}},
					},
				},
			},
		},
		{
			name: "full sync policy",
			sourceState: map[string]interface{}{
				"metadata": []interface{}{
					map[string]interface{}{
						"name":      "test",
						"namespace": "argocd",
					},
				},
				"spec": []interface{}{
					map[string]interface{}{
						"source": []interface{}{map[string]interface{}{
							"repo_url":        "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami",
							"chart":           "redis",
							"target_revision": "16.9.11",
						}},
						"destination": []interface{}{map[string]interface{}{
							"server":    "https://kubernetes.default.svc",
							"namespace": "default",
						}},
						"sync_policy": []interface{}{map[string]interface{}{
							"automated": map[string]interface{}{
								"prune":       true,
								"self_heal":   true,
								"allow_empty": true,
							},
							"sync_options": []string{
								"Validate=false",
							},
							"retry": []interface{}{map[string]interface{}{
								"limit": "5",
								"backoff": map[string]interface{}{
									"duration":     "30s",
									"max_duration": "2m",
									"factor":       "2",
								},
							}},
						}},
					},
				},
			},
			expectedState: map[string]interface{}{
				"metadata": []interface{}{
					map[string]interface{}{
						"name":      "test",
						"namespace": "argocd",
					},
				},
				"spec": []interface{}{
					map[string]interface{}{
						"source": []interface{}{map[string]interface{}{
							"repo_url":        "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami",
							"chart":           "redis",
							"target_revision": "16.9.11",
						}},
						"destination": []interface{}{map[string]interface{}{
							"server":    "https://kubernetes.default.svc",
							"namespace": "default",
						}},
						"sync_policy": []interface{}{map[string]interface{}{
							"automated": []map[string]interface{}{
								{
									"prune":       true,
									"self_heal":   true,
									"allow_empty": true,
								},
							},
							"sync_options": []string{
								"Validate=false",
							},
							"retry": []interface{}{map[string]interface{}{
								"limit": "5",
								"backoff": []map[string]interface{}{
									{
										"duration":     "30s",
										"max_duration": "2m",
										"factor":       "2",
									},
								},
							}},
						}},
					},
				},
			},
		},
		{
			name: "no automated block",
			sourceState: map[string]interface{}{
				"metadata": []interface{}{
					map[string]interface{}{
						"name":      "test",
						"namespace": "argocd",
					},
				},
				"spec": []interface{}{
					map[string]interface{}{
						"source": []interface{}{map[string]interface{}{
							"repo_url":        "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami",
							"chart":           "redis",
							"target_revision": "16.9.11",
						}},
						"destination": []interface{}{map[string]interface{}{
							"server":    "https://kubernetes.default.svc",
							"namespace": "default",
						}},
						"sync_policy": []interface{}{map[string]interface{}{
							"sync_options": []string{
								"Validate=false",
							},
							"retry": []interface{}{map[string]interface{}{
								"limit": "5",
								"backoff": map[string]interface{}{
									"duration":     "30s",
									"max_duration": "2m",
									"factor":       "2",
								},
							}},
						}},
					},
				},
			},
			expectedState: map[string]interface{}{
				"metadata": []interface{}{
					map[string]interface{}{
						"name":      "test",
						"namespace": "argocd",
					},
				},
				"spec": []interface{}{
					map[string]interface{}{
						"source": []interface{}{map[string]interface{}{
							"repo_url":        "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami",
							"chart":           "redis",
							"target_revision": "16.9.11",
						}},
						"destination": []interface{}{map[string]interface{}{
							"server":    "https://kubernetes.default.svc",
							"namespace": "default",
						}},
						"sync_policy": []interface{}{map[string]interface{}{
							"sync_options": []string{
								"Validate=false",
							},
							"retry": []interface{}{map[string]interface{}{
								"limit": "5",
								"backoff": []map[string]interface{}{
									{
										"duration":     "30s",
										"max_duration": "2m",
										"factor":       "2",
									},
								},
							}},
						}},
					},
				},
			},
		},
		{
			name: "blank automated block",
			sourceState: map[string]interface{}{
				"metadata": []interface{}{
					map[string]interface{}{
						"name":      "test",
						"namespace": "argocd",
					},
				},
				"spec": []interface{}{
					map[string]interface{}{
						"source": []interface{}{map[string]interface{}{
							"repo_url":        "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami",
							"chart":           "redis",
							"target_revision": "16.9.11",
						}},
						"destination": []interface{}{map[string]interface{}{
							"server":    "https://kubernetes.default.svc",
							"namespace": "default",
						}},
						"sync_policy": []interface{}{map[string]interface{}{
							"automated": map[string]interface{}{},
							"sync_options": []string{
								"Validate=false",
							},
							"retry": []interface{}{map[string]interface{}{
								"limit": "5",
								"backoff": map[string]interface{}{
									"duration":     "30s",
									"max_duration": "2m",
									"factor":       "2",
								},
							}},
						}},
					},
				},
			},
			expectedState: map[string]interface{}{
				"metadata": []interface{}{
					map[string]interface{}{
						"name":      "test",
						"namespace": "argocd",
					},
				},
				"spec": []interface{}{
					map[string]interface{}{
						"source": []interface{}{map[string]interface{}{
							"repo_url":        "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami",
							"chart":           "redis",
							"target_revision": "16.9.11",
						}},
						"destination": []interface{}{map[string]interface{}{
							"server":    "https://kubernetes.default.svc",
							"namespace": "default",
						}},
						"sync_policy": []interface{}{map[string]interface{}{
							"automated": []map[string]interface{}{{}},
							"sync_options": []string{
								"Validate=false",
							},
							"retry": []interface{}{map[string]interface{}{
								"limit": "5",
								"backoff": []map[string]interface{}{
									{
										"duration":     "30s",
										"max_duration": "2m",
										"factor":       "2",
									},
								},
							}},
						}},
					},
				},
			},
		},
		{
			name: "no backoff",
			sourceState: map[string]interface{}{
				"metadata": []interface{}{
					map[string]interface{}{
						"name":      "test",
						"namespace": "argocd",
					},
				},
				"spec": []interface{}{
					map[string]interface{}{
						"source": []interface{}{map[string]interface{}{
							"repo_url":        "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami",
							"chart":           "redis",
							"target_revision": "16.9.11",
						}},
						"destination": []interface{}{map[string]interface{}{
							"server":    "https://kubernetes.default.svc",
							"namespace": "default",
						}},
						"sync_policy": []interface{}{map[string]interface{}{
							"automated": map[string]interface{}{
								"prune":       true,
								"self_heal":   true,
								"allow_empty": true,
							},
							"sync_options": []string{
								"Validate=false",
							},
							"retry": []interface{}{map[string]interface{}{
								"limit": "5",
							}},
						}},
					},
				},
			},
			expectedState: map[string]interface{}{
				"metadata": []interface{}{
					map[string]interface{}{
						"name":      "test",
						"namespace": "argocd",
					},
				},
				"spec": []interface{}{
					map[string]interface{}{
						"source": []interface{}{map[string]interface{}{
							"repo_url":        "https://raw.githubusercontent.com/bitnami/charts/archive-full-index/bitnami",
							"chart":           "redis",
							"target_revision": "16.9.11",
						}},
						"destination": []interface{}{map[string]interface{}{
							"server":    "https://kubernetes.default.svc",
							"namespace": "default",
						}},
						"sync_policy": []interface{}{map[string]interface{}{
							"automated": []map[string]interface{}{
								{
									"prune":       true,
									"self_heal":   true,
									"allow_empty": true,
								},
							},
							"sync_options": []string{
								"Validate=false",
							},
							"retry": []interface{}{map[string]interface{}{
								"limit": "5",
							}},
						}},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actualState, err := resourceArgoCDApplicationStateUpgradeV3(t.Context(), tc.sourceState, nil)
			if err != nil {
				t.Fatalf("error migrating state: %s", err)
			}

			if !reflect.DeepEqual(actualState, tc.expectedState) {
				t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", tc.expectedState, actualState)
			}
		})
	}
}

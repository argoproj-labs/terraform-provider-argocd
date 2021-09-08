package argocd

import (
	application "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func expandApplicationDestinations(ds *schema.Set) (
	result []application.ApplicationDestination) {
	for _, dest := range ds.List() {
		result = append(result, expandApplicationDestination(dest))
	}
	return
}

func expandSyncWindows(sws []interface{}) (
	result []*application.SyncWindow) {
	for _, _sw := range sws {
		sw := _sw.(map[string]interface{})
		result = append(
			result,
			&application.SyncWindow{
				Applications: expandStringList(sw["applications"].([]interface{})),
				Clusters:     expandStringList(sw["clusters"].([]interface{})),
				Duration:     sw["duration"].(string),
				Kind:         sw["kind"].(string),
				ManualSync:   sw["manual_sync"].(bool),
				Namespaces:   expandStringList(sw["namespaces"].([]interface{})),
				Schedule:     sw["schedule"].(string),
			},
		)
	}
	return
}

func expandK8SGroupKind(groupKinds *schema.Set) (
	result []meta.GroupKind) {
	for _, _gk := range groupKinds.List() {
		gk := _gk.(map[string]interface{})
		result = append(result, meta.GroupKind{
			Group: gk["group"].(string),
			Kind:  gk["kind"].(string),
		})
	}
	return
}

func flattenK8SGroupKinds(gks []meta.GroupKind) (
	result []map[string]string) {
	for _, gk := range gks {
		result = append(result, map[string]string{
			"group": gk.Group,
			"kind":  gk.Kind,
		})
	}
	return
}

func flattenSyncWindows(sws application.SyncWindows) (
	result []map[string]interface{}) {
	for _, sw := range sws {
		result = append(result, map[string]interface{}{
			"applications": sw.Applications,
			"clusters":     sw.Clusters,
			"duration":     sw.Duration,
			"kind":         sw.Kind,
			"manual_sync":  sw.ManualSync,
			"namespaces":   sw.Namespaces,
			"schedule":     sw.Schedule,
		})
	}
	return
}

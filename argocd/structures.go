package argocd

import (
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func expandApplicationDestination(ds *schema.Set) (
	result []v1alpha1.ApplicationDestination) {
	for _, _dest := range ds.List() {
		dest := _dest.(map[string]interface{})
		result = append(
			result,
			v1alpha1.ApplicationDestination{
				Server:    dest["server"].(string),
				Namespace: dest["namespace"].(string),
			},
		)
	}
	return
}

func expandSyncWindows(sws []interface{}) (
	result []*v1alpha1.SyncWindow) {
	for _, _sw := range sws {
		sw := _sw.(map[string]interface{})
		result = append(
			result,
			&v1alpha1.SyncWindow{
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
	result []v1.GroupKind) {
	for _, _gk := range groupKinds.List() {
		gk := _gk.(map[string]interface{})
		result = append(result, v1.GroupKind{
			Group: gk["group"].(string),
			Kind:  gk["kind"].(string),
		})
	}
	return
}

func flattenApplicationDestinations(ds []v1alpha1.ApplicationDestination) (
	result []map[string]string) {
	for _, d := range ds {
		result = append(result, map[string]string{
			"server":    d.Server,
			"namespace": d.Namespace,
		})
	}
	return
}

func flattenK8SGroupKinds(gks []v1.GroupKind) (
	result []map[string]string) {
	for _, gk := range gks {
		result = append(result, map[string]string{
			"group": gk.Group,
			"kind":  gk.Kind,
		})
	}
	return
}

func flattenSyncWindows(sws v1alpha1.SyncWindows) (
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

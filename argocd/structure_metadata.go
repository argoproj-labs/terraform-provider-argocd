package argocd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func expandMetadata(d *schema.ResourceData) (meta meta.ObjectMeta) {
	m := d.Get("metadata.0").(map[string]interface{})

	if v, ok := m["annotations"].(map[string]interface{}); ok && len(v) > 0 {
		meta.Annotations = expandStringMap(m["annotations"].(map[string]interface{}))
	}

	if v, ok := m["labels"].(map[string]interface{}); ok && len(v) > 0 {
		meta.Labels = expandStringMap(m["labels"].(map[string]interface{}))
	}

	if v, ok := m["name"]; ok {
		meta.Name = v.(string)
	}

	if v, ok := m["namespace"]; ok {
		meta.Namespace = v.(string)
	}

	return meta
}

func flattenMetadata(meta meta.ObjectMeta, d *schema.ResourceData) []interface{} {
	m := map[string]interface{}{
		"generation":       meta.Generation,
		"name":             meta.Name,
		"namespace":        meta.Namespace,
		"resource_version": meta.ResourceVersion,
		"uid":              fmt.Sprintf("%v", meta.UID),
	}

	annotations := d.Get("metadata.0.annotations").(map[string]interface{})
	m["annotations"] = metadataRemoveInternalKeys(meta.Annotations, annotations)

	labels := d.Get("metadata.0.labels").(map[string]interface{})
	m["labels"] = metadataRemoveInternalKeys(meta.Labels, labels)

	return []interface{}{m}
}

func metadataRemoveInternalKeys(m map[string]string, d map[string]interface{}) map[string]string {
	for k := range m {
		if metadataIsInternalKey(k) && !isKeyInMap(k, d) {
			delete(m, k)
		}
	}

	return m
}

// argocdInternalKeys are metadata keys that Argo CD itself writes on the
// objects it manages. They are not part of the desired state, so they must not
// show up as drift unless they are explicitly set in the configuration.
var argocdInternalKeys = map[string]bool{
	// Set by the notifications controller.
	"notified.notifications.argoproj.io": true,
	// Set by the API server on every refresh request (UI "Refresh",
	// `argocd app get --refresh`) and removed by the application controller.
	"argocd.argoproj.io/refresh": true,
	// Set together with the refresh annotation since Argo CD 3.x, but only
	// removed by the source hydrator. It therefore stays on every Application
	// that does not use spec.sourceHydrator (or when the hydrator is disabled).
	"argocd.argoproj.io/hydrate": true,
}

func metadataIsInternalKey(annotationKey string) bool {
	if argocdInternalKeys[annotationKey] {
		return true
	}

	u, err := url.Parse("//" + annotationKey)
	if err != nil {
		return false
	}

	return strings.HasSuffix(u.Hostname(), "kubernetes.io")
}

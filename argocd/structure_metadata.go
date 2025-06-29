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

	if v, ok := m["finalizers"].([]interface{}); ok && len(v) > 0 {
		meta.Finalizers = expandStringList(v)
	}

	if v, ok := m["name"]; ok {
		meta.Name = v.(string)
	}

	if v, ok := m["namespace"]; ok {
		meta.Namespace = v.(string)
	}

	return meta
}

// expandMetadataForUpdate safely expands metadata for updates, merging user-configured
// finalizers with existing system finalizers to prevent accidental removal
func expandMetadataForUpdate(d *schema.ResourceData, existingMeta meta.ObjectMeta) (meta meta.ObjectMeta) {
	meta = expandMetadata(d)

	// Merge finalizers: keep existing system finalizers, add/update user finalizers
	if len(existingMeta.Finalizers) > 0 {
		userFinalizers := make(map[string]bool)
		for _, f := range meta.Finalizers {
			userFinalizers[f] = true
		}

		// Start with existing finalizers
		merged := make([]string, 0, len(existingMeta.Finalizers)+len(meta.Finalizers))
		for _, existing := range existingMeta.Finalizers {
			if userFinalizers[existing] {
				// User explicitly configured this finalizer, keep it
				merged = append(merged, existing)
				delete(userFinalizers, existing)
			} else {
				// System finalizer not configured by user, preserve it
				merged = append(merged, existing)
			}
		}

		// Add any new user finalizers
		for finalizer := range userFinalizers {
			merged = append(merged, finalizer)
		}

		meta.Finalizers = merged
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

	finalizers := d.Get("metadata.0.finalizers").([]interface{})
	m["finalizers"] = metadataFilterFinalizers(meta.Finalizers, finalizers)

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

func metadataIsInternalKey(annotationKey string) bool {
	u, err := url.Parse("//" + annotationKey)
	if err != nil {
		return false
	}

	return strings.HasSuffix(u.Hostname(), "kubernetes.io") || annotationKey == "notified.notifications.argoproj.io"
}

func metadataFilterFinalizers(apiFinalizers []string, configuredFinalizers []interface{}) []string {
	configured := make(map[string]bool)
	for _, v := range configuredFinalizers {
		if s, ok := v.(string); ok {
			configured[s] = true
		}
	}

	result := make([]string, 0)
	for _, finalizer := range apiFinalizers {
		// Only include finalizers that were explicitly configured by the user
		if configured[finalizer] {
			result = append(result, finalizer)
		}
	}
	return result
}

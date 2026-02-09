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

// expandMetadataForUpdate safely expands metadata for updates, using Terraform's
// state tracking to distinguish between user-managed and system-managed finalizers.
// This allows users to remove finalizers they previously set while preserving
// finalizers added by controllers or other systems.
func expandMetadataForUpdate(d *schema.ResourceData, existingMeta meta.ObjectMeta) (meta meta.ObjectMeta) {
	meta = expandMetadata(d)

	// Use GetChange to determine what Terraform previously managed vs what user wants now
	oldRaw, newRaw := d.GetChange("metadata.0.finalizers")
	oldFinalizers := toStringSet(oldRaw)
	newFinalizers := toStringSet(newRaw)

	merged := make([]string, 0, len(existingMeta.Finalizers)+len(newFinalizers))

	// Process existing finalizers from the API
	for _, existing := range existingMeta.Finalizers {
		if oldFinalizers[existing] && !newFinalizers[existing] {
			// This finalizer was previously managed by Terraform but user removed it
			// from config - don't preserve it (allow removal)
			continue
		}

		if !newFinalizers[existing] {
			// This finalizer exists on API but was never in Terraform state
			// It's a system-managed finalizer - preserve it
			merged = append(merged, existing)
		} else {
			// User wants this finalizer and it exists - keep it
			merged = append(merged, existing)
			delete(newFinalizers, existing)
		}
	}

	// Add any new finalizers the user configured that don't exist yet
	for finalizer := range newFinalizers {
		merged = append(merged, finalizer)
	}

	meta.Finalizers = merged

	return meta
}

// toStringSet converts an interface{} (expected to be []interface{}) to a map[string]bool set
func toStringSet(v interface{}) map[string]bool {
	result := make(map[string]bool)
	if v == nil {
		return result
	}

	if list, ok := v.([]interface{}); ok {
		for _, item := range list {
			if s, ok := item.(string); ok {
				result[s] = true
			}
		}
	}

	return result
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

	// Check if this is an import operation (no resource_version in state)
	isImport := d.Get("metadata.0.resource_version") == "" || d.Get("metadata.0.resource_version") == nil

	finalizers := d.Get("metadata.0.finalizers").([]interface{})
	if isImport {
		// During import, return all finalizers from API so user can see what exists
		m["finalizers"] = meta.Finalizers
	} else {
		// During normal read, filter to only show configured finalizers
		m["finalizers"] = metadataFilterFinalizers(meta.Finalizers, finalizers)
	}

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

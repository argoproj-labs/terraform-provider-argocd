package argocd

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func metadataSchema(name string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Standard Kubernetes object metadata. For more info see the [Kubernetes reference](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata).",
		Required:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: metadataFields(name),
		},
	}
}

func metadataFields(objectName string) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"annotations": {
			Type:         schema.TypeMap,
			Description:  fmt.Sprintf("An unstructured key value map stored with the %s that may be used to store arbitrary metadata. More info: http://kubernetes.io/docs/user-guide/annotations", objectName),
			Optional:     true,
			Elem:         &schema.Schema{Type: schema.TypeString},
			ValidateFunc: validateMetadataAnnotations,
		},
		"generation": {
			Type:        schema.TypeInt,
			Description: "A sequence number representing a specific generation of the desired state.",
			Computed:    true,
		},
		"labels": {
			Type:         schema.TypeMap,
			Description:  fmt.Sprintf("Map of string keys and values that can be used to organize and categorize (scope and select) the %s. May match selectors of replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels", objectName),
			Optional:     true,
			Elem:         &schema.Schema{Type: schema.TypeString},
			ValidateFunc: validateMetadataLabels(false),
		},
		"name": {
			Type:         schema.TypeString,
			Description:  fmt.Sprintf("Name of the %s, must be unique. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names", objectName),
			Optional:     true,
			ForceNew:     true,
			Computed:     true,
			ValidateFunc: validateMetadataName,
		},
		"namespace": {
			Type:         schema.TypeString,
			Description:  fmt.Sprintf("Namespace of the %s, must be unique. Cannot be updated. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/", objectName),
			Optional:     true,
			ForceNew:     true,
			Computed:     true,
			ValidateFunc: validateMetadataName,
		},
		"resource_version": {
			Type:        schema.TypeString,
			Description: fmt.Sprintf("An opaque value that represents the internal version of this %s that can be used by clients to determine when %s has changed. Read more: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency", objectName, objectName),
			Computed:    true,
		},
		"uid": {
			Type:        schema.TypeString,
			Description: fmt.Sprintf("The unique in time and space value for this %s. More info: http://kubernetes.io/docs/user-guide/identifiers#uids", objectName),
			Computed:    true,
		},
		"finalizers": {
			Type:        schema.TypeList,
			Description: "List of finalizers for the resource.",
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
	}
}

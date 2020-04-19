package argocd

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func metadataSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		MinItems:    1,
		MaxItems:    1,
		Description: "Kubernetes resource metadata. Required attributes: name, namespace.",
		Required:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"namespace": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"uid": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"labels": {
					Type:     schema.TypeMap,
					Computed: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"annotations": {
					Type:     schema.TypeMap,
					Computed: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"ClusterName": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"resource_version": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"generation": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"creation_timestamp": {
					Type:         schema.TypeString,
					Computed:     true,
					ValidateFunc: validation.IsRFC3339Time,
				},
				"deletion_timestamp": {
					Type:         schema.TypeString,
					Computed:     true,
					ValidateFunc: validation.IsRFC3339Time,
				},
				"finalizers": {
					Type:     schema.TypeSet,
					Set:      schema.HashString,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"deletion_grace_period_seconds": {
					Type:     schema.TypeString,
					Optional: true,
				},
				// TODO: add the rest
			},
		},
	}
}

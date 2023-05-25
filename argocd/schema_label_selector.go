package argocd

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func labelSelectorSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"match_expressions": matchExpressionsSchema(),
		"match_labels": {
			Type:        schema.TypeMap,
			Description: "A map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of `match_expressions`, whose key field is \"key\", the operator is \"In\", and the values array contains only \"value\". The requirements are ANDed.",
			Optional:    true,
		},
	}
}

func matchExpressionsSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "A list of label selector requirements. The requirements are ANDed.",
		Optional:    true,
		ForceNew:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key": {
					Type:        schema.TypeString,
					Description: "The label key that the selector applies to.",
					Optional:    true,
				},
				"operator": {
					Type:        schema.TypeString,
					Description: "A key's relationship to a set of values. Valid operators ard `In`, `NotIn`, `Exists` and `DoesNotExist`.",
					Optional:    true,
				},
				"values": {
					Type:        schema.TypeSet,
					Description: "An array of string values. If the operator is `In` or `NotIn`, the values array must be non-empty. If the operator is `Exists` or `DoesNotExist`, the values array must be empty. This array is replaced during a strategic merge patch.",
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
					Set:         schema.HashString,
				},
			},
		},
	}
}

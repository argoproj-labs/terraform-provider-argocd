package argocd

import (
	"fmt"
	"strconv"
	"strings"

	application "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func expandIntOrString(s string) (*intstr.IntOrString, error) {
	if len(s) == 0 {
		return nil, nil
	}

	if strings.HasSuffix(s, "%") {
		return &intstr.IntOrString{
			StrVal: s,
			Type:   intstr.String,
		}, nil
	}

	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to convert string to int32: %w", err)
	}

	return &intstr.IntOrString{
		IntVal: int32(i),
		Type:   intstr.Int,
	}, nil
}

func expandSecretRef(sr map[string]interface{}) *application.SecretRef {
	return &application.SecretRef{
		Key:        sr["key"].(string),
		SecretName: sr["secret_name"].(string),
	}
}

func flattenIntOrString(ios *intstr.IntOrString) string {
	if ios == nil {
		return ""
	}

	switch {
	case ios.StrVal != "":
		return ios.StrVal
	default:
		return strconv.Itoa(int(ios.IntVal))
	}
}

func flattenSecretRef(sr application.SecretRef) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"key":         sr.Key,
			"secret_name": sr.SecretName,
		},
	}
}

func newStringSet(f schema.SchemaSetFunc, in []string) *schema.Set {
	var out = make([]interface{}, len(in))

	for i, v := range in {
		out[i] = v
	}

	return schema.NewSet(f, out)
}

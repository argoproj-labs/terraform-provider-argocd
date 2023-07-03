package argocd

import (
	"fmt"
	"strconv"
	"strings"

	application "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func expandK8SGroupKind(groupKinds *schema.Set) (result []meta.GroupKind) {
	for _, _gk := range groupKinds.List() {
		gk := _gk.(map[string]interface{})

		result = append(result, meta.GroupKind{
			Group: gk["group"].(string),
			Kind:  gk["kind"].(string),
		})
	}

	return
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

func flattenK8SGroupKinds(gks []meta.GroupKind) (result []map[string]string) {
	for _, gk := range gks {
		result = append(result, map[string]string{
			"group": gk.Group,
			"kind":  gk.Kind,
		})
	}

	return
}

func flattenSecretRef(sr application.SecretRef) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"key":         sr.Key,
			"secret_name": sr.SecretName,
		},
	}
}

func flattenSyncWindows(sws application.SyncWindows) (result []map[string]interface{}) {
	for _, sw := range sws {
		result = append(result, map[string]interface{}{
			"applications": sw.Applications,
			"clusters":     sw.Clusters,
			"duration":     sw.Duration,
			"kind":         sw.Kind,
			"manual_sync":  sw.ManualSync,
			"namespaces":   sw.Namespaces,
			"schedule":     sw.Schedule,
			"timezone":     sw.TimeZone,
		})
	}

	return
}

func newStringSet(f schema.SchemaSetFunc, in []string) *schema.Set {
	var out = make([]interface{}, len(in))

	for i, v := range in {
		out[i] = v
	}

	return schema.NewSet(f, out)
}

package argocd

import (
	"fmt"
	application "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func expandApplicationSpec(d *schema.ResourceData) (
	spec application.ApplicationSpec,
	err error) {

	s := d.Get("spec.0").(map[string]interface{})

	if v, ok := s["project"]; ok {
		spec.Project = v.(string)
	}
	if v, ok := s["revision_history_limit"]; ok {
		pv := v.(int64)
		spec.RevisionHistoryLimit = &pv
	}
	if v, ok := s["info"]; ok {
		spec.Info = expandApplicationInfo(v.(*schema.Set))
	}
	if v, ok := s["ignore_differences"]; ok {
		spec.IgnoreDifferences = expandApplicationIgnoreDifferences(v.([]map[string]interface{}))
	}
	// TODO
	return spec, nil
}

func expandApplicationIgnoreDifferences(ids []map[string]interface{}) (
	result []application.ResourceIgnoreDifferences) {
	for _, id := range ids {
		var elem = application.ResourceIgnoreDifferences{}
		if v, ok := id["group"]; ok {
			elem.Group = v.(string)
		}
		if v, ok := id["kind"]; ok {
			elem.Kind = v.(string)
		}
		if v, ok := id["name"]; ok {
			elem.Name = v.(string)
		}
		if v, ok := id["namespace"]; ok {
			elem.Namespace = v.(string)
		}
		if v, ok := id["json_pointers"]; ok {
			jps := v.(*schema.Set).List()
			for _, jp := range jps {
				elem.JSONPointers = append(elem.JSONPointers, jp.(string))
			}
		}
		result = append(result, elem)
	}
	return
}

func expandApplicationInfo(infos *schema.Set) (
	result []application.Info) {
	for _, i := range infos.List() {
		result = append(result, application.Info{
			Name:  i.(map[string]string)["name"],
			Value: i.(map[string]string)["value"],
		})
	}
	return
}

func expandApplicationDestination(dest interface{}) (
	result application.ApplicationDestination) {
	d, ok := dest.(map[string]interface{})
	if !ok {
		panic(fmt.Errorf("could not expand application destination"))
	}
	return application.ApplicationDestination{
		Server:    d["server"].(string),
		Namespace: d["namespace"].(string),
	}
}

func flattenApplicationDestination(dest application.ApplicationDestination) map[string]string {
	return map[string]string{
		"server":    dest.Server,
		"namespace": dest.Namespace,
	}
}

func flattenApplicationSpec(s application.ApplicationSpec) (
	[]map[string]interface{},
	error) {
	spec := map[string]interface{}{
		"destination": flattenApplicationDestination(s.Destination),
		// TODO
	}
	return []map[string]interface{}{spec}, nil
}

func expandApplication(d *schema.ResourceData) (
	metadata meta.ObjectMeta,
	spec application.ApplicationSpec,
	err error) {
	metadata = expandMetadata(d)
	spec, err = expandApplicationSpec(d)
	return
}

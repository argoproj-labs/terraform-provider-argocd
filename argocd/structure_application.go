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

	return spec, nil
}

func expandApplicationDestination(dest interface{}) (
	result application.ApplicationDestination) {
	d, ok := dest.(map[string]interface{})
	if !ok {
		panic(fmt.Errorf("could not expand application destination"))
	}
	// TODO
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

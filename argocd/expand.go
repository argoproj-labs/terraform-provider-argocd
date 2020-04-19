package argocd

import (
	"fmt"
	argoCDAppv1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/mitchellh/mapstructure"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: needs comprehensive unit tests
func expandArgoCDProject(d *schema.ResourceData) (
	objectMeta metav1.ObjectMeta,
	spec argoCDAppv1.AppProjectSpec,
	err error) {

	// Initialize new mapstructure decoders
	md, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused: true,
		TagName:     "json",
		Result:      &objectMeta,
		// TODO: replace all keys to camelCase equivalents
		// TODO: convert generation field from string to int64
		// TODO: convert creation_timestamp field from string to Time
		//DecodeHook:
	})
	if err != nil {
		return
	}
	sd, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused: true,
		TagName:     "json",
		Result:      &spec,
		// TODO: replace all keys to camelCase equivalents
		//DecodeHook:
	})
	if err != nil {
		return
	}
	// Expand project metadata
	m := d.Get("metadata.0")
	if err := md.Decode(m); err != nil {
		return objectMeta, spec, fmt.Errorf("metadata expansion: %s | %v", err, m)
	}
	// Expand project spec
	s := d.Get("spec.0")
	if err := sd.Decode(s); err != nil {
		return objectMeta, spec, fmt.Errorf("spec expansion: %s | %v ", err, s)
	}
	return
}

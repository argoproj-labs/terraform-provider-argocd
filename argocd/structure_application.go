package argocd

import (
	"encoding/json"
	"fmt"

	application "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Expand

func expandApplication(d *schema.ResourceData) (
	metadata meta.ObjectMeta,
	spec application.ApplicationSpec,
	diags diag.Diagnostics) {

	metadata = expandMetadata(d)
	spec, diags = expandApplicationSpec(d)
	return
}

func expandApplicationSpec(d *schema.ResourceData) (
	spec application.ApplicationSpec,
	diags diag.Diagnostics) {

	s := d.Get("spec.0").(map[string]interface{})

	if v, ok := s["project"]; ok {
		spec.Project = v.(string)
	}
	if v, ok := s["revision_history_limit"]; ok {
		pv := int64(v.(int))
		spec.RevisionHistoryLimit = &pv
	}
	if v, ok := s["info"]; ok {
		spec.Info = expandApplicationInfo(v.(*schema.Set))
	}
	if v, ok := s["ignore_difference"]; ok {
		spec.IgnoreDifferences = expandApplicationIgnoreDifferences(v.([]interface{}))
	}
	if v, ok := s["sync_policy"]; ok {
		spec.SyncPolicy, diags = expandApplicationSyncPolicy(v.([]interface{}))
	}
	if v, ok := s["destination"]; ok {
		spec.Destination = expandApplicationDestination(v.(*schema.Set).List()[0])
	}
	if v, ok := s["source"]; ok {
		spec.Source = expandApplicationSource(v.([]interface{})[0])
	}
	return spec, diags
}

func expandApplicationSource(_as interface{}) (
	result application.ApplicationSource) {
	as := _as.(map[string]interface{})
	if v, ok := as["repo_url"]; ok {
		result.RepoURL = v.(string)
	}
	if v, ok := as["path"]; ok {
		result.Path = v.(string)
	}
	if v, ok := as["target_revision"]; ok {
		result.TargetRevision = v.(string)
	}
	if v, ok := as["chart"]; ok {
		result.Chart = v.(string)
	}
	if v, ok := as["helm"]; ok {
		result.Helm = expandApplicationSourceHelm(v.([]interface{}))
	}
	if v, ok := as["kustomize"]; ok {
		result.Kustomize = expandApplicationSourceKustomize(v.([]interface{}))
	}
	if v, ok := as["ksonnet"]; ok {
		result.Ksonnet = expandApplicationSourceKsonnet(v.([]interface{}))
	}
	if v, ok := as["directory"]; ok {
		result.Directory = expandApplicationSourceDirectory(v.([]interface{}))
	}
	if v, ok := as["plugin"]; ok {
		result.Plugin = expandApplicationSourcePlugin(v.([]interface{}))
	}
	return
}

func expandApplicationSourcePlugin(in []interface{}) *application.ApplicationSourcePlugin {
	if len(in) == 0 {
		return nil
	}
	a := in[0].(map[string]interface{})
	result := &application.ApplicationSourcePlugin{}
	if v, ok := a["name"]; ok {
		result.Name = v.(string)
	}
	if env, ok := a["env"]; ok {
		for _, v := range env.(*schema.Set).List() {
			result.Env = append(result.Env,
				&application.EnvEntry{
					Name:  v.(map[string]interface{})["name"].(string),
					Value: v.(map[string]interface{})["value"].(string),
				},
			)
		}
	}
	return result
}

func expandApplicationSourceDirectory(in []interface{}) *application.ApplicationSourceDirectory {
	if len(in) == 0 || in[0] == nil {
		return nil
	}
	a := in[0].(map[string]interface{})
	result := &application.ApplicationSourceDirectory{}
	if v, ok := a["recurse"]; ok {
		result.Recurse = v.(bool)
	}
	if aj, ok := a["jsonnet"].([]interface{}); ok {
		jsonnet := application.ApplicationSourceJsonnet{}
		if len(aj) > 0 && aj[0] != nil {
			j := aj[0].(map[string]interface{})
			if evs, ok := j["ext_var"].([]interface{}); ok && len(evs) > 0 {
				for _, v := range evs {
					if vv, ok := v.(map[string]interface{}); ok {
						jsonnet.ExtVars = append(jsonnet.ExtVars,
							application.JsonnetVar{
								Name:  vv["name"].(string),
								Value: vv["value"].(string),
								Code:  vv["code"].(bool),
							},
						)
					}
				}
			}

			if tlas, ok := j["tla"].(*schema.Set); ok && len(tlas.List()) > 0 {
				for _, v := range tlas.List() {
					if vv, ok := v.(map[string]interface{}); ok {
						jsonnet.TLAs = append(jsonnet.TLAs,
							application.JsonnetVar{
								Name:  vv["name"].(string),
								Value: vv["value"].(string),
								Code:  vv["code"].(bool),
							},
						)
					}
				}
			}
		}
		result.Jsonnet = jsonnet
	}
	return result
}

func expandApplicationSourceKsonnet(in []interface{}) *application.ApplicationSourceKsonnet {
	if len(in) == 0 {
		return nil
	}
	a := in[0].(map[string]interface{})
	result := &application.ApplicationSourceKsonnet{}
	if v, ok := a["environment"]; ok {
		result.Environment = v.(string)
	}
	if parameters, ok := a["parameters"]; ok {
		for _, _p := range parameters.(*schema.Set).List() {
			p := _p.(map[string]string)
			parameter := application.KsonnetParameter{}
			if v, ok := p["name"]; ok {
				parameter.Name = v
			}
			if v, ok := p["value"]; ok {
				parameter.Value = v
			}
			if v, ok := p["component"]; ok {
				parameter.Component = v
			}
			result.Parameters = append(result.Parameters, parameter)
		}
	}
	return result
}

func expandApplicationSourceKustomize(in []interface{}) *application.ApplicationSourceKustomize {
	if len(in) == 0 {
		return nil
	}
	a := in[0].(map[string]interface{})
	result := &application.ApplicationSourceKustomize{}
	if v, ok := a["name_prefix"]; ok {
		result.NamePrefix = v.(string)
	}
	if v, ok := a["name_suffix"]; ok {
		result.NameSuffix = v.(string)
	}
	if v, ok := a["version"]; ok {
		result.Version = v.(string)
	}
	if v, ok := a["images"]; ok {
		for _, i := range v.(*schema.Set).List() {
			result.Images = append(
				result.Images,
				application.KustomizeImage(i.(string)),
			)
		}
	}
	if cls, ok := a["common_labels"]; ok {
		result.CommonLabels = make(map[string]string, 0)
		for k, v := range cls.(map[string]interface{}) {
			result.CommonLabels[k] = v.(string)
		}
	}
	if cas, ok := a["common_annotations"]; ok {
		result.CommonAnnotations = make(map[string]string, 0)
		for k, v := range cas.(map[string]interface{}) {
			result.CommonAnnotations[k] = v.(string)
		}
	}
	return result
}

func expandApplicationSourceHelm(in []interface{}) *application.ApplicationSourceHelm {
	if len(in) == 0 {
		return nil
	}
	a := in[0].(map[string]interface{})
	result := &application.ApplicationSourceHelm{}
	if v, ok := a["value_files"]; ok {
		for _, vf := range v.([]interface{}) {
			result.ValueFiles = append(result.ValueFiles, vf.(string))
		}
	}
	if v, ok := a["values"]; ok {
		result.Values = v.(string)
	}
	if v, ok := a["release_name"]; ok {
		result.ReleaseName = v.(string)
	}
	if parameters, ok := a["parameter"]; ok {
		for _, _p := range parameters.(*schema.Set).List() {
			p := _p.(map[string]interface{})
			parameter := application.HelmParameter{}
			if v, ok := p["force_string"]; ok {
				parameter.ForceString = v.(bool)
			}
			if v, ok := p["name"]; ok {
				parameter.Name = v.(string)
			}
			if v, ok := p["value"]; ok {
				parameter.Value = v.(string)
			}
			result.Parameters = append(result.Parameters, parameter)
		}
	}
	return result
}

func expandApplicationSyncPolicy(_sp []interface{}) (*application.SyncPolicy, diag.Diagnostics) {
	if len(_sp) == 0 {
		return nil, nil
	}
	sp := _sp[0]
	if sp == nil {
		return &application.SyncPolicy{}, nil
	}
	var automated = &application.SyncPolicyAutomated{}
	var syncOptions application.SyncOptions
	var retry = &application.RetryStrategy{}
	var syncPolicy = &application.SyncPolicy{}

	if a, ok := sp.(map[string]interface{})["automated"].(map[string]interface{}); ok {
		if len(a) > 0 {
			for k, v := range a {
				if k == "prune" {
					automated.Prune = v.(bool)
				}
				if k == "self_heal" {
					automated.SelfHeal = v.(bool)
				}
				if k == "allow_empty" {
					automated.AllowEmpty = v.(bool)
				}
			}
			syncPolicy.Automated = automated
		}
	}
	if v, ok := sp.(map[string]interface{})["sync_options"]; ok {
		sOpts := v.([]interface{})
		if len(sOpts) > 0 {
			for _, sOpt := range sOpts {
				syncOptions = append(syncOptions, sOpt.(string))
			}
			syncPolicy.SyncOptions = syncOptions
		}
	}
	if _retry, ok := sp.(map[string]interface{})["retry"].([]interface{}); ok {
		if len(_retry) > 0 {
			r := (_retry[0]).(map[string]interface{})
			if len(r) > 0 {
				for k, v := range r {
					if k == "limit" {
						var err error
						retry.Limit, err = convertStringToInt64(v.(string))
						if err != nil {
							return nil, []diag.Diagnostic{
								{
									Severity: diag.Error,
									Summary:  "Error converting retry limit to integer",
									Detail:   err.Error(),
								},
							}
						}
					}
					if k == "backoff" {
						retry.Backoff = &application.Backoff{}
						for kb, vb := range v.(map[string]interface{}) {
							if kb == "duration" {
								retry.Backoff.Duration = vb.(string)
							}
							if kb == "max_duration" {
								retry.Backoff.MaxDuration = vb.(string)
							}
							if kb == "factor" {
								factor, err := convertStringToInt64Pointer(vb.(string))
								if err != nil {
									return nil, []diag.Diagnostic{
										{
											Severity: diag.Error,
											Summary:  "Error converting backoff factor to integer",
											Detail:   err.Error(),
										},
									}
								}
								retry.Backoff.Factor = factor
							}
						}
					}
				}
				syncPolicy.Retry = retry
			}
		}
	}
	return syncPolicy, nil
}

func expandApplicationIgnoreDifferences(ids []interface{}) (
	result []application.ResourceIgnoreDifferences) {
	for _, _id := range ids {
		id := _id.(map[string]interface{})
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
		if v, ok := id["jq_path_expressions"]; ok {
			jqpes := v.(*schema.Set).List()
			for _, jqpe := range jqpes {
				elem.JQPathExpressions = append(elem.JQPathExpressions, jqpe.(string))
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
		Name:      d["name"].(string),
	}
}

// Flatten

func flattenApplication(app *application.Application, d *schema.ResourceData) error {
	fMetadata := flattenMetadata(app.ObjectMeta, d)
	fSpec, err := flattenApplicationSpec(app.Spec)
	if err != nil {
		return err
	}
	if err := d.Set("spec", fSpec); err != nil {
		e, _ := json.MarshalIndent(fSpec, "", "\t")
		return fmt.Errorf("error persisting spec: %s\n%s", err, e)
	}
	if err := d.Set("metadata", fMetadata); err != nil {
		e, _ := json.MarshalIndent(fMetadata, "", "\t")
		return fmt.Errorf("error persisting metadata: %s\n%s", err, e)
	}
	return nil
}

func flattenApplicationSpec(s application.ApplicationSpec) (
	[]map[string]interface{},
	error) {
	spec := map[string]interface{}{
		"destination": flattenApplicationDestinations(
			[]application.ApplicationDestination{s.Destination},
		),
		"ignore_difference": flattenApplicationIgnoreDifferences(s.IgnoreDifferences),
		"info":              flattenApplicationInfo(s.Info),
		"project":           s.Project,
		"source": flattenApplicationSource(
			[]application.ApplicationSource{s.Source},
		),
		"sync_policy": flattenApplicationSyncPolicy(s.SyncPolicy),
	}
	if s.RevisionHistoryLimit != nil {
		spec["revision_history_limit"] = int(*s.RevisionHistoryLimit)
	}
	return []map[string]interface{}{spec}, nil
}

func flattenApplicationSyncPolicy(sp *application.SyncPolicy) []map[string]interface{} {
	if sp == nil {
		return nil
	}
	result := make(map[string]interface{}, 0)
	backoff := make(map[string]string, 0)
	if sp.Automated != nil {
		result["automated"] = map[string]bool{
			"prune":       sp.Automated.Prune,
			"self_heal":   sp.Automated.SelfHeal,
			"allow_empty": sp.Automated.AllowEmpty,
		}
	}
	result["sync_options"] = []string(sp.SyncOptions)
	if sp.Retry != nil {
		limit := convertInt64ToString(sp.Retry.Limit)
		if sp.Retry.Backoff != nil {
			backoff = map[string]string{
				"duration":     sp.Retry.Backoff.Duration,
				"max_duration": sp.Retry.Backoff.MaxDuration,
			}
			if sp.Retry.Backoff.Factor != nil {
				backoff["factor"] = convertInt64PointerToString(sp.Retry.Backoff.Factor)
			}
		}
		result["retry"] = []map[string]interface{}{
			{
				"limit":   limit,
				"backoff": backoff,
			},
		}
	}
	return []map[string]interface{}{result}
}

func flattenApplicationInfo(infos []application.Info) (
	result []map[string]string) {
	for _, i := range infos {
		result = append(result, map[string]string{
			"name":  i.Name,
			"value": i.Value,
		})
	}
	return
}

func flattenApplicationIgnoreDifferences(ids []application.ResourceIgnoreDifferences) (
	result []map[string]interface{}) {
	for _, id := range ids {
		result = append(result, map[string]interface{}{
			"group":               id.Group,
			"kind":                id.Kind,
			"name":                id.Name,
			"namespace":           id.Namespace,
			"json_pointers":       id.JSONPointers,
			"jq_path_expressions": id.JQPathExpressions,
		})
	}
	return
}

func flattenApplicationSource(source []application.ApplicationSource) (
	result []map[string]interface{}) {
	for _, s := range source {
		result = append(result, map[string]interface{}{
			"chart": s.Chart,
			"directory": flattenApplicationSourceDirectory(
				[]*application.ApplicationSourceDirectory{s.Directory},
			),
			"helm": flattenApplicationSourceHelm(
				[]*application.ApplicationSourceHelm{s.Helm},
			),
			"ksonnet": flattenApplicationSourceKsonnet(
				[]*application.ApplicationSourceKsonnet{s.Ksonnet},
			),
			"kustomize": flattenApplicationSourceKustomize(
				[]*application.ApplicationSourceKustomize{s.Kustomize},
			),
			"path": s.Path,
			"plugin": flattenApplicationSourcePlugin(
				[]*application.ApplicationSourcePlugin{s.Plugin},
			),
			"repo_url":        s.RepoURL,
			"target_revision": s.TargetRevision,
		})
	}
	return
}

func flattenApplicationSourcePlugin(as []*application.ApplicationSourcePlugin) (
	result []map[string]interface{}) {
	for _, a := range as {
		if a != nil {
			var env []map[string]string
			for _, e := range a.Env {
				env = append(env, map[string]string{
					"name":  e.Name,
					"value": e.Value,
				})
			}
			result = append(result, map[string]interface{}{
				"name": a.Name,
				"env":  env,
			})
		}
	}
	return
}

func flattenApplicationSourceKsonnet(as []*application.ApplicationSourceKsonnet) (
	result []map[string]interface{}) {
	for _, a := range as {
		if a != nil {
			var parameters []map[string]string
			for _, p := range a.Parameters {
				parameters = append(parameters,
					map[string]string{
						"component": p.Component,
						"name":      p.Name,
						"value":     p.Value,
					},
				)
			}
			result = append(result, map[string]interface{}{
				"environment": a.Environment,
				"parameters":  parameters,
			})
		}
	}
	return
}

func flattenApplicationSourceDirectory(as []*application.ApplicationSourceDirectory) (
	result []map[string]interface{}) {
	for _, a := range as {
		if a != nil {
			jsonnet := make(map[string][]interface{}, 0)
			for _, jev := range a.Jsonnet.ExtVars {
				jsonnet["ext_var"] = append(jsonnet["ext_var"], map[string]interface{}{
					"code":  jev.Code,
					"name":  jev.Name,
					"value": jev.Value,
				})
			}
			for _, jtla := range a.Jsonnet.TLAs {
				jsonnet["tla"] = append(jsonnet["tla"], map[string]interface{}{
					"code":  jtla.Code,
					"name":  jtla.Name,
					"value": jtla.Value,
				})
			}

			m := make(map[string]interface{})
			m["recurse"] = a.Recurse

			if len(jsonnet) > 0 {
				m["jsonnet"] = []map[string][]interface{}{jsonnet}
			}
			result = append(result, m)
		}
	}
	return
}

func flattenApplicationSourceKustomize(as []*application.ApplicationSourceKustomize) (
	result []map[string]interface{}) {
	for _, a := range as {
		if a != nil {
			var images []string
			for _, i := range a.Images {
				images = append(images, string(i))
			}
			result = append(result, map[string]interface{}{
				"common_annotations": a.CommonAnnotations,
				"common_labels":      a.CommonLabels,
				"images":             images,
				"name_prefix":        a.NamePrefix,
				"name_suffix":        a.NameSuffix,
				"version":            a.Version,
			})
		}
	}
	return
}

func flattenApplicationSourceHelm(as []*application.ApplicationSourceHelm) (
	result []map[string]interface{}) {
	for _, a := range as {
		if a != nil {
			var parameters []map[string]interface{}
			for _, p := range a.Parameters {
				parameters = append(parameters, map[string]interface{}{
					"force_string": p.ForceString,
					"name":         p.Name,
					"value":        p.Value,
				})
			}
			result = append(result, map[string]interface{}{
				"parameter":    parameters,
				"release_name": a.ReleaseName,
				"value_files":  a.ValueFiles,
				"values":       a.Values,
			})
		}
	}
	return
}

func flattenApplicationDestinations(ds []application.ApplicationDestination) (
	result []map[string]string) {
	for _, d := range ds {
		result = append(result, map[string]string{
			"namespace": d.Namespace,
			"server":    d.Server,
			"name":      d.Name,
		})
	}
	return
}

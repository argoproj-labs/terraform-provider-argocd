package argocd

import (
	"encoding/json"
	"fmt"
	application "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Expand

func expandApplication(d *schema.ResourceData) (
	metadata meta.ObjectMeta,
	spec application.ApplicationSpec,
	err error) {
	metadata = expandMetadata(d)
	spec, err = expandApplicationSpec(d)
	return
}

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
	if v, ok := s["sync_policy"]; ok {
		spec.SyncPolicy = expandApplicationSyncPolicy(v.([]map[string]interface{})[0])
	}
	if v, ok := s["destination"]; ok {
		spec.Destination = expandApplicationDestination(v.(*schema.Set))
	}
	if v, ok := s["source"]; ok {
		spec.Source = expandApplicationSource(v.([]map[string]interface{})[0])
	}
	return spec, nil
}

func expandApplicationSource(as map[string]interface{}) (
	result application.ApplicationSource) {
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
		result.Helm = expandApplicationSourceHelm(v.([]map[string]interface{})[0])
	}
	if v, ok := as["kustomize"]; ok {
		result.Kustomize = expandApplicationSourceKustomize(v.([]map[string]interface{})[0])
	}
	if v, ok := as["ksonnet"]; ok {
		result.Ksonnet = expandApplicationSourceKsonnet(v.([]map[string]interface{})[0])
	}
	if v, ok := as["directory"]; ok {
		result.Directory = expandApplicationSourceDirectory(v.([]map[string]interface{})[0])
	}
	if v, ok := as["plugin"]; ok {
		result.Plugin = expandApplicationSourcePlugin(v.([]map[string]interface{})[0])
	}
	return
}

func expandApplicationSourcePlugin(a map[string]interface{}) *application.ApplicationSourcePlugin {
	result := &application.ApplicationSourcePlugin{}
	if v, ok := a["name"]; ok {
		result.Name = v.(string)
	}
	if env, ok := a["env"]; ok {
		for _, v := range env.(*schema.Set).List() {
			result.Env = append(result.Env,
				&application.EnvEntry{
					Name:  v.(map[string]string)["name"],
					Value: v.(map[string]string)["value"],
				},
			)
		}
	}
	return result
}

func expandApplicationSourceDirectory(a map[string]interface{}) *application.ApplicationSourceDirectory {
	result := &application.ApplicationSourceDirectory{}
	if v, ok := a["recurse"]; ok {
		result.Recurse = v.(bool)
	}
	if _j, ok := a["jsonnet"]; ok {
		jsonnet := application.ApplicationSourceJsonnet{}
		j := _j.([]map[string][]map[string]interface{})

		if evs, ok := j[0]["ext_vars"]; ok && len(evs) > 0 {
			for _, v := range evs {
				jsonnet.ExtVars = append(jsonnet.ExtVars,
					application.JsonnetVar{
						Name:  v["name"].(string),
						Value: v["value"].(string),
						Code:  v["code"].(bool),
					},
				)
			}
		}
		if tlas, ok := j[0]["tlas"]; ok && len(tlas) > 0 {
			for _, v := range tlas {
				jsonnet.TLAs = append(jsonnet.TLAs,
					application.JsonnetVar{
						Name:  v["name"].(string),
						Value: v["value"].(string),
						Code:  v["code"].(bool),
					},
				)
			}
		}
		result.Jsonnet = jsonnet
	}
	return result
}

func expandApplicationSourceKsonnet(a map[string]interface{}) *application.ApplicationSourceKsonnet {
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

func expandApplicationSourceKustomize(a map[string]interface{}) *application.ApplicationSourceKustomize {
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
		for k, v := range cls.(map[string]string) {
			result.CommonLabels[k] = v
		}
	}
	return result
}

func expandApplicationSourceHelm(a map[string]interface{}) *application.ApplicationSourceHelm {
	result := &application.ApplicationSourceHelm{}
	if v, ok := a["values"]; ok {
		result.Values = v.(string)
	}
	if v, ok := a["release_name"]; ok {
		result.ReleaseName = v.(string)
	}
	if parameters, ok := a["parameters"]; ok {
		for _, p := range parameters.([]map[string]interface{}) {
			parameter := application.HelmParameter{}
			if v, ok := p["name"]; ok {
				parameter.Name = v.(string)
			}
			if v, ok := p["value"]; ok {
				parameter.Value = v.(string)
			}
			if v, ok := p["force_string"]; ok {
				parameter.ForceString = v.(bool)
			}
			result.Parameters = append(result.Parameters, parameter)
		}
	}
	return result
}

func expandApplicationSyncPolicy(sp map[string]interface{}) *application.SyncPolicy {
	var automated = &application.SyncPolicyAutomated{}
	var syncOptions application.SyncOptions

	if v, ok := sp["automated"]; ok {
		a := v.(map[string]bool)
		if prune, ok := a["prune"]; ok {
			automated.Prune = prune
		}
		if selfHeal, ok := a["self_heal"]; ok {
			automated.SelfHeal = selfHeal
		}
	}
	if v, ok := sp["sync_options"]; ok {
		sOpts := v.(*schema.Set).List()
		for _, sOpt := range sOpts {
			syncOptions = append(syncOptions, sOpt.(string))
		}
	}
	return &application.SyncPolicy{
		Automated:   automated,
		SyncOptions: syncOptions,
	}
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
		"ignore_differences":     flattenApplicationIgnoreDifferences(s.IgnoreDifferences),
		"info":                   flattenApplicationInfo(s.Info),
		"project":                s.Project,
		"revision_history_limit": *s.RevisionHistoryLimit,
		"source": flattenApplicationSource(
			[]application.ApplicationSource{s.Source},
		),
		"sync_policy": s.SyncPolicy,
	}
	return []map[string]interface{}{spec}, nil
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
			"group":         id.Group,
			"kind":          id.Kind,
			"name":          id.Name,
			"namespace":     id.Namespace,
			"json_pointers": id.JSONPointers,
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
	return
}

func flattenApplicationSourceKsonnet(as []*application.ApplicationSourceKsonnet) (
	result []map[string]interface{}) {
	for _, a := range as {
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
	return
}

func flattenApplicationSourceDirectory(as []*application.ApplicationSourceDirectory) (
	result []map[string]interface{}) {
	for _, a := range as {
		jsonnet := make(map[string][]interface{}, 0)
		for _, jev := range a.Jsonnet.ExtVars {
			jsonnet["ext_vars"] = append(jsonnet["ext_vars"], map[string]interface{}{
				"code":  jev.Code,
				"name":  jev.Name,
				"value": jev.Value,
			})
		}
		for _, jtla := range a.Jsonnet.TLAs {
			jsonnet["tlas"] = append(jsonnet["tlas"], map[string]interface{}{
				"code":  jtla.Code,
				"name":  jtla.Name,
				"value": jtla.Value,
			})
		}
		result = append(result, map[string]interface{}{
			"jsonnet": []map[string][]interface{}{jsonnet},
			"recurse": a.Recurse,
		})
	}
	return
}

func flattenApplicationSourceKustomize(as []*application.ApplicationSourceKustomize) (
	result []map[string]interface{}) {
	for _, a := range as {
		var images []string
		for _, i := range a.Images {
			images = append(images, string(i))
		}
		result = append(result, map[string]interface{}{
			"common_labels": a.CommonLabels,
			"images":        images,
			"name_prefix":   a.NamePrefix,
			"name_suffix":   a.NameSuffix,
			"version":       a.Version,
		})
	}
	return
}

func flattenApplicationSourceHelm(as []*application.ApplicationSourceHelm) (
	result []map[string]interface{}) {
	for _, a := range as {
		var parameters []map[string]interface{}
		for _, p := range a.Parameters {
			parameters = append(parameters, map[string]interface{}{
				"force_string": p.ForceString,
				"name":         p.Name,
				"value":        p.Value,
			})
		}
		result = append(result, map[string]interface{}{
			"parameters":   parameters,
			"release_name": a.ReleaseName,
			"value_files":  a.ValueFiles,
			"values":       a.Values,
		})
	}
	return
}

func flattenApplicationDestinations(ds []application.ApplicationDestination) (
	result []map[string]string) {
	for _, d := range ds {
		result = append(result, map[string]string{
			"namespace": d.Namespace,
			"server":    d.Server,
		})
	}
	return
}

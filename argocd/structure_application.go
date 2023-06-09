package argocd

import (
	"encoding/json"
	"fmt"

	application "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Expand

func expandApplication(d *schema.ResourceData) (metadata meta.ObjectMeta, spec application.ApplicationSpec, err error) {
	metadata = expandMetadata(d)
	spec, err = expandApplicationSpec(d.Get("spec.0").(map[string]interface{}))

	return
}

func expandApplicationSpec(s map[string]interface{}) (spec application.ApplicationSpec, err error) {
	if v, ok := s["project"]; ok {
		spec.Project = v.(string)
	}

	if v, ok := s["revision_history_limit"]; ok {
		pv := int64(v.(int))
		spec.RevisionHistoryLimit = &pv
	}

	if v, ok := s["info"]; ok {
		spec.Info, err = expandApplicationInfo(v.(*schema.Set))
		if err != nil {
			return
		}
	}

	if v, ok := s["ignore_difference"]; ok {
		spec.IgnoreDifferences = expandApplicationIgnoreDifferences(v.([]interface{}))
	}

	if v, ok := s["sync_policy"].([]interface{}); ok && len(v) > 0 {
		spec.SyncPolicy, err = expandApplicationSyncPolicy(v[0])
		if err != nil {
			return
		}
	}

	if v, ok := s["destination"]; ok {
		spec.Destination = expandApplicationDestination(v.(*schema.Set).List()[0])
	}

	if v, ok := s["source"].([]interface{}); ok && len(v) > 0 {
		spec.Sources = expandApplicationSource(v)
	}

	return spec, nil
}

func expandApplicationSource(_ass []interface{}) []application.ApplicationSource {
	ass := make([]application.ApplicationSource, len(_ass))

	for i, v := range _ass {
		as := v.(map[string]interface{})
		s := application.ApplicationSource{}

		if v, ok := as["repo_url"]; ok {
			s.RepoURL = v.(string)
		}

		if v, ok := as["path"]; ok {
			s.Path = v.(string)
		}

		if v, ok := as["ref"]; ok {
			s.Ref = v.(string)
		}

		if v, ok := as["target_revision"]; ok {
			s.TargetRevision = v.(string)
		}

		if v, ok := as["chart"]; ok {
			s.Chart = v.(string)
		}

		if v, ok := as["helm"]; ok {
			s.Helm = expandApplicationSourceHelm(v.([]interface{}))
		}

		if v, ok := as["kustomize"]; ok {
			s.Kustomize = expandApplicationSourceKustomize(v.([]interface{}))
		}

		if v, ok := as["directory"].([]interface{}); ok && len(v) > 0 {
			s.Directory = expandApplicationSourceDirectory(v[0])
		}

		if v, ok := as["plugin"]; ok {
			s.Plugin = expandApplicationSourcePlugin(v.([]interface{}))
		}

		ass[i] = s
	}

	return ass
}

func expandApplicationSourcePlugin(in []interface{}) *application.ApplicationSourcePlugin {
	if len(in) == 0 {
		return nil
	}

	result := &application.ApplicationSourcePlugin{}

	a := in[0].(map[string]interface{})
	if v, ok := a["name"]; ok {
		result.Name = v.(string)
	}

	if env, ok := a["env"]; ok {
		for _, v := range env.(*schema.Set).List() {
			result.Env = append(result.Env, &application.EnvEntry{
				Name:  v.(map[string]interface{})["name"].(string),
				Value: v.(map[string]interface{})["value"].(string),
			})
		}
	}

	return result
}

func expandApplicationSourceDirectory(in interface{}) *application.ApplicationSourceDirectory {
	result := &application.ApplicationSourceDirectory{}

	if in == nil {
		return result
	}

	a := in.(map[string]interface{})
	if v, ok := a["recurse"]; ok {
		result.Recurse = v.(bool)
	}

	if v, ok := a["exclude"]; ok {
		result.Exclude = v.(string)
	}

	if v, ok := a["include"]; ok {
		result.Include = v.(string)
	}

	if aj, ok := a["jsonnet"].([]interface{}); ok {
		jsonnet := application.ApplicationSourceJsonnet{}

		if len(aj) > 0 && aj[0] != nil {
			j := aj[0].(map[string]interface{})
			if evs, ok := j["ext_var"].([]interface{}); ok && len(evs) > 0 {
				for _, v := range evs {
					if vv, ok := v.(map[string]interface{}); ok {
						jsonnet.ExtVars = append(jsonnet.ExtVars, application.JsonnetVar{
							Name:  vv["name"].(string),
							Value: vv["value"].(string),
							Code:  vv["code"].(bool),
						})
					}
				}
			}

			if tlas, ok := j["tla"].(*schema.Set); ok && len(tlas.List()) > 0 {
				for _, v := range tlas.List() {
					if vv, ok := v.(map[string]interface{}); ok {
						jsonnet.TLAs = append(jsonnet.TLAs, application.JsonnetVar{
							Name:  vv["name"].(string),
							Value: vv["value"].(string),
							Code:  vv["code"].(bool),
						})
					}
				}
			}

			if libs, ok := j["libs"].([]interface{}); ok && len(libs) > 0 {
				for _, lib := range libs {
					jsonnet.Libs = append(jsonnet.Libs, lib.(string))
				}
			}
		}

		result.Jsonnet = jsonnet
	}

	return result
}

func expandApplicationSourceKustomize(in []interface{}) *application.ApplicationSourceKustomize {
	if len(in) == 0 {
		return nil
	}

	result := &application.ApplicationSourceKustomize{}

	a := in[0].(map[string]interface{})
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
			result.Images = append(result.Images, application.KustomizeImage(i.(string)))
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

	result := &application.ApplicationSourceHelm{}

	a := in[0].(map[string]interface{})
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

	if v, ok := a["pass_credentials"]; ok {
		result.PassCredentials = v.(bool)
	}

	if v, ok := a["ignore_missing_value_files"]; ok {
		result.IgnoreMissingValueFiles = v.(bool)
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

	if fileParameters, ok := a["file_parameter"]; ok {
		for _, _p := range fileParameters.(*schema.Set).List() {
			p := _p.(map[string]interface{})

			parameter := application.HelmFileParameter{}

			if v, ok := p["name"]; ok {
				parameter.Name = v.(string)
			}

			if v, ok := p["path"]; ok {
				parameter.Path = v.(string)
			}

			result.FileParameters = append(result.FileParameters, parameter)
		}
	}

	if v, ok := a["skip_crds"]; ok {
		result.SkipCrds = v.(bool)
	}

	return result
}

func expandApplicationSyncPolicy(sp interface{}) (*application.SyncPolicy, error) {
	var syncPolicy = &application.SyncPolicy{}

	if sp == nil {
		return syncPolicy, nil
	}

	p := sp.(map[string]interface{})

	if _a, ok := p["automated"].(*schema.Set); ok {
		var automated = &application.SyncPolicyAutomated{}

		list := _a.List()

		if len(list) > 0 {
			a := list[0].(map[string]interface{})
			if v, ok := a["prune"]; ok {
				automated.Prune = v.(bool)
			}

			if v, ok := a["self_heal"]; ok {
				automated.SelfHeal = v.(bool)
			}

			if v, ok := a["allow_empty"]; ok {
				automated.AllowEmpty = v.(bool)
			}

			syncPolicy.Automated = automated
		}
	}

	if _sOpts, ok := p["sync_options"].([]interface{}); ok && len(_sOpts) > 0 {
		var syncOptions application.SyncOptions

		for _, so := range _sOpts {
			syncOptions = append(syncOptions, so.(string))
		}

		syncPolicy.SyncOptions = syncOptions
	}

	if _retry, ok := p["retry"].([]interface{}); ok && len(_retry) > 0 {
		var retry = &application.RetryStrategy{}

		r := (_retry[0]).(map[string]interface{})

		if v, ok := r["limit"]; ok {
			var err error

			retry.Limit, err = convertStringToInt64(v.(string))
			if err != nil {
				return nil, fmt.Errorf("failed to convert retry limit to integer: %w", err)
			}
		}

		if _b, ok := r["backoff"].(*schema.Set); ok {
			retry.Backoff = &application.Backoff{}

			list := _b.List()
			if len(list) > 0 {
				b := list[0].(map[string]interface{})

				if v, ok := b["duration"]; ok {
					retry.Backoff.Duration = v.(string)
				}

				if v, ok := b["max_duration"]; ok {
					retry.Backoff.MaxDuration = v.(string)
				}

				if v, ok := b["factor"]; ok {
					factor, err := convertStringToInt64Pointer(v.(string))
					if err != nil {
						return nil, fmt.Errorf("failed to convert backoff factor to integer: %w", err)
					}

					retry.Backoff.Factor = factor
				}
			}
		}

		syncPolicy.Retry = retry
	}

	if _mnm, ok := p["managed_namespace_metadata"].([]interface{}); ok && len(_mnm) > 0 {
		mnm := _mnm[0].(map[string]interface{})
		syncPolicy.ManagedNamespaceMetadata = &application.ManagedNamespaceMetadata{}

		if a, ok := mnm["annotations"]; ok {
			syncPolicy.ManagedNamespaceMetadata.Annotations = expandStringMap(a.(map[string]interface{}))
		}

		if l, ok := mnm["labels"]; ok {
			syncPolicy.ManagedNamespaceMetadata.Labels = expandStringMap(l.(map[string]interface{}))
		}
	}

	return syncPolicy, nil
}

func expandApplicationIgnoreDifferences(ids []interface{}) (result []application.ResourceIgnoreDifferences) {
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

	return //nolint:nakedret // overriding as function follows pattern in rest of file
}

func expandApplicationInfo(infos *schema.Set) (result []application.Info, err error) {
	for _, i := range infos.List() {
		item := i.(map[string]interface{})
		info := application.Info{}
		fieldSet := false

		if name, ok := item["name"].(string); ok && name != "" {
			info.Name = name
			fieldSet = true
		}

		if value, ok := item["value"].(string); ok && value != "" {
			info.Value = value
			fieldSet = true
		}

		if !fieldSet {
			return result, fmt.Errorf("spec.info: cannot be empty - must only contains 'name' or 'value' fields")
		}

		result = append(result, info)
	}

	return
}

func expandApplicationDestinations(ds *schema.Set) (result []application.ApplicationDestination) {
	for _, dest := range ds.List() {
		result = append(result, expandApplicationDestination(dest))
	}

	return
}

func expandApplicationDestination(dest interface{}) (result application.ApplicationDestination) {
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

func expandSyncWindows(sws []interface{}) (result []*application.SyncWindow) {
	for _, _sw := range sws {
		sw := _sw.(map[string]interface{})

		result = append(result, &application.SyncWindow{
			Applications: expandStringList(sw["applications"].([]interface{})),
			Clusters:     expandStringList(sw["clusters"].([]interface{})),
			Duration:     sw["duration"].(string),
			Kind:         sw["kind"].(string),
			ManualSync:   sw["manual_sync"].(bool),
			Namespaces:   expandStringList(sw["namespaces"].([]interface{})),
			Schedule:     sw["schedule"].(string),
		})
	}

	return
}

// Flatten

func flattenApplication(app *application.Application, d *schema.ResourceData) error {
	metadata := flattenMetadata(app.ObjectMeta, d)
	if err := d.Set("metadata", metadata); err != nil {
		e, _ := json.MarshalIndent(metadata, "", "\t")
		return fmt.Errorf("error persisting metadata: %s\n%s", err, e)
	}

	spec := flattenApplicationSpec(app.Spec)
	if err := d.Set("spec", spec); err != nil {
		e, _ := json.MarshalIndent(spec, "", "\t")
		return fmt.Errorf("error persisting spec: %s\n%s", err, e)
	}

	status := flattenApplicationStatus(app.Status)
	if err := d.Set("status", status); err != nil {
		e, _ := json.MarshalIndent(status, "", "\t")
		return fmt.Errorf("error persisting status: %s\n%s", err, e)
	}

	return nil
}

func flattenApplicationSpec(s application.ApplicationSpec) []map[string]interface{} {
	spec := map[string]interface{}{
		"destination":       flattenApplicationDestinations([]application.ApplicationDestination{s.Destination}),
		"ignore_difference": flattenApplicationIgnoreDifferences(s.IgnoreDifferences),
		"info":              flattenApplicationInfo(s.Info),
		"project":           s.Project,
		"sync_policy":       flattenApplicationSyncPolicy(s.SyncPolicy),
	}

	if s.Source != nil {
		spec["source"] = flattenApplicationSource([]application.ApplicationSource{*s.Source})
	} else {
		spec["source"] = flattenApplicationSource(s.Sources)
	}

	if s.RevisionHistoryLimit != nil {
		spec["revision_history_limit"] = int(*s.RevisionHistoryLimit)
	}

	return []map[string]interface{}{spec}
}

func flattenApplicationSyncPolicy(sp *application.SyncPolicy) []map[string]interface{} {
	if sp == nil {
		return nil
	}

	result := make(map[string]interface{}, 0)
	backoff := make(map[string]interface{}, 0)

	if sp.Automated != nil {
		result["automated"] = []map[string]interface{}{
			{
				"prune":       sp.Automated.Prune,
				"self_heal":   sp.Automated.SelfHeal,
				"allow_empty": sp.Automated.AllowEmpty,
			},
		}
	}

	if sp.ManagedNamespaceMetadata != nil {
		result["managed_namespace_metadata"] = []map[string]interface{}{
			{
				"annotations": sp.ManagedNamespaceMetadata.Annotations,
				"labels":      sp.ManagedNamespaceMetadata.Labels,
			},
		}
	}

	result["sync_options"] = []string(sp.SyncOptions)

	if sp.Retry != nil {
		limit := convertInt64ToString(sp.Retry.Limit)

		if sp.Retry.Backoff != nil {
			backoff = map[string]interface{}{
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
				"backoff": []map[string]interface{}{backoff},
			},
		}
	}

	return []map[string]interface{}{result}
}

func flattenApplicationInfo(infos []application.Info) (result []map[string]string) {
	for _, i := range infos {
		info := map[string]string{}

		if i.Name != "" {
			info["name"] = i.Name
		}

		if i.Value != "" {
			info["value"] = i.Value
		}

		result = append(result, info)
	}

	return
}

func flattenApplicationIgnoreDifferences(ids []application.ResourceIgnoreDifferences) (result []map[string]interface{}) {
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

func flattenApplicationSource(source []application.ApplicationSource) (result []map[string]interface{}) {
	for _, s := range source {
		result = append(result, map[string]interface{}{
			"chart":           s.Chart,
			"directory":       flattenApplicationSourceDirectory([]*application.ApplicationSourceDirectory{s.Directory}),
			"helm":            flattenApplicationSourceHelm([]*application.ApplicationSourceHelm{s.Helm}),
			"kustomize":       flattenApplicationSourceKustomize([]*application.ApplicationSourceKustomize{s.Kustomize}),
			"path":            s.Path,
			"plugin":          flattenApplicationSourcePlugin([]*application.ApplicationSourcePlugin{s.Plugin}),
			"ref":             s.Ref,
			"repo_url":        s.RepoURL,
			"target_revision": s.TargetRevision,
		})
	}

	return
}

func flattenApplicationSourcePlugin(as []*application.ApplicationSourcePlugin) (result []map[string]interface{}) {
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

func flattenApplicationSourceDirectory(as []*application.ApplicationSourceDirectory) (result []map[string]interface{}) {
	for _, a := range as {
		if a != nil && !a.IsZero() {
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

			for _, lib := range a.Jsonnet.Libs {
				jsonnet["libs"] = append(jsonnet["libs"], lib)
			}

			m := map[string]interface{}{
				"recurse": a.Recurse,
				"exclude": a.Exclude,
				"include": a.Include,
			}

			if len(jsonnet) > 0 {
				m["jsonnet"] = []map[string][]interface{}{jsonnet}
			}

			result = append(result, m)
		}
	}

	return //nolint:nakedret // only just breaching - function follows pattern in rest of file
}

func flattenApplicationSourceKustomize(as []*application.ApplicationSourceKustomize) (result []map[string]interface{}) {
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

func flattenApplicationSourceHelm(as []*application.ApplicationSourceHelm) (result []map[string]interface{}) {
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

			var fileParameters []map[string]interface{}
			for _, p := range a.FileParameters {
				fileParameters = append(fileParameters, map[string]interface{}{
					"name": p.Name,
					"path": p.Path,
				})
			}

			result = append(result, map[string]interface{}{
				"parameter":                  parameters,
				"file_parameter":             fileParameters,
				"release_name":               a.ReleaseName,
				"skip_crds":                  a.SkipCrds,
				"value_files":                a.ValueFiles,
				"values":                     a.Values,
				"pass_credentials":           a.PassCredentials,
				"ignore_missing_value_files": a.IgnoreMissingValueFiles,
			})
		}
	}

	return result
}

func flattenApplicationDestinations(ds []application.ApplicationDestination) (result []map[string]string) {
	for _, d := range ds {
		result = append(result, map[string]string{
			"namespace": d.Namespace,
			"server":    d.Server,
			"name":      d.Name,
		})
	}

	return
}

func flattenApplicationStatus(s application.ApplicationStatus) []map[string]interface{} {
	status := map[string]interface{}{
		"conditions": flattenApplicationConditions(s.Conditions),
		"health":     flattenApplicationHealthStatus(s.Health),
		"resources":  flattenApplicationResourceStatuses(s.Resources),
		"summary":    flattenApplicationSummary(s.Summary),
		"sync":       flattenApplicationSyncStatus(s.Sync),
	}

	if s.OperationState != nil {
		status["operation_state"] = flattenApplicationOperationState(*s.OperationState)
	}

	if s.ReconciledAt != nil {
		status["reconciled_at"] = s.ReconciledAt.String()
	}

	return []map[string]interface{}{status}
}

func flattenApplicationConditions(aacs []application.ApplicationCondition) []map[string]interface{} {
	acs := make([]map[string]interface{}, len(aacs))

	for i, v := range aacs {
		acs[i] = map[string]interface{}{
			"message": v.Message,
			"type":    v.Type,
		}

		if v.LastTransitionTime != nil {
			acs[i]["last_transition_time"] = v.LastTransitionTime.String()
		}
	}

	return acs
}

func flattenApplicationHealthStatus(hs application.HealthStatus) []map[string]interface{} {
	h := map[string]interface{}{
		"message": hs.Message,
		"status":  hs.Status,
	}

	return []map[string]interface{}{h}
}

func flattenApplicationSyncStatus(ss application.SyncStatus) []map[string]interface{} {
	s := map[string]interface{}{
		"revision":  ss.Revision,
		"revisions": ss.Revisions,
		"status":    ss.Status,
	}

	return []map[string]interface{}{s}
}

func flattenApplicationResourceStatuses(arss []application.ResourceStatus) []map[string]interface{} {
	rss := make([]map[string]interface{}, len(arss))

	for i, v := range arss {
		rss[i] = map[string]interface{}{
			"group":            v.Group,
			"hook":             v.Hook,
			"kind":             v.Kind,
			"name":             v.Name,
			"namespace":        v.Namespace,
			"requires_pruning": v.RequiresPruning,
			"status":           v.Status,
			"sync_wave":        convertInt64ToString(v.SyncWave),
			"version":          v.Version,
		}

		if v.Health != nil {
			rss[i]["health"] = flattenApplicationHealthStatus(*v.Health)
		}
	}

	return rss
}

func flattenApplicationSummary(as application.ApplicationSummary) []map[string]interface{} {
	s := map[string]interface{}{
		"external_urls": as.ExternalURLs,
		"images":        as.Images,
	}

	return []map[string]interface{}{s}
}

func flattenApplicationOperationState(os application.OperationState) []map[string]interface{} {
	s := map[string]interface{}{
		"message":     os.Message,
		"phase":       os.Phase,
		"retry_count": convertInt64ToString(os.RetryCount),
		"started_at":  os.StartedAt.String(),
	}

	if os.FinishedAt != nil {
		s["finished_at"] = os.FinishedAt.String()
	}

	return []map[string]interface{}{s}
}

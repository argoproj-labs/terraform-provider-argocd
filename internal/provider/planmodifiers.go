package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// UseUnknownOnUpdate returns a plan modifier that sets the value to unknown
// whenever the resource is being updated. This is useful for computed fields like
// resource_version and generation that change on every Kubernetes API call.
//
// Unlike UseStateForUnknown which preserves the prior state value, this modifier
// marks the value as unknown during updates so that Terraform accepts any value
// returned by the provider after apply.
//
// Fixes: https://github.com/argoproj-labs/terraform-provider-argocd/issues/807
func UseUnknownOnUpdateString() planmodifier.String {
	return useUnknownOnUpdateStringModifier{}
}

type useUnknownOnUpdateStringModifier struct{}

func (m useUnknownOnUpdateStringModifier) Description(_ context.Context) string {
	return "Sets the value to unknown during updates since server-managed fields change on every API call."
}

func (m useUnknownOnUpdateStringModifier) MarkdownDescription(_ context.Context) string {
	return "Sets the value to unknown during updates since server-managed fields change on every API call."
}

func (m useUnknownOnUpdateStringModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If there's no state (create), leave as unknown (default behavior)
	if req.State.Raw.IsNull() {
		return
	}

	// If the plan is being destroyed, no need to modify
	if req.Plan.Raw.IsNull() {
		return
	}

	// This is an update - check if any values in the resource are changing
	// by comparing the full plan to the full state using Equal
	if !req.Plan.Raw.Equal(req.State.Raw) {
		// Resource is being modified, mark as unknown so any value is accepted
		resp.PlanValue = types.StringUnknown()
		return
	}

	// No change to the resource, preserve the state value
	resp.PlanValue = req.StateValue
}

// UseUnknownOnUpdateInt64 returns a plan modifier for Int64 attributes
// that sets the value to unknown whenever the resource is being updated.
func UseUnknownOnUpdateInt64() planmodifier.Int64 {
	return useUnknownOnUpdateInt64Modifier{}
}

type useUnknownOnUpdateInt64Modifier struct{}

func (m useUnknownOnUpdateInt64Modifier) Description(_ context.Context) string {
	return "Sets the value to unknown during updates since server-managed fields change on every API call."
}

func (m useUnknownOnUpdateInt64Modifier) MarkdownDescription(_ context.Context) string {
	return "Sets the value to unknown during updates since server-managed fields change on every API call."
}

func (m useUnknownOnUpdateInt64Modifier) PlanModifyInt64(_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	// If there's no state (create), leave as unknown (default behavior)
	if req.State.Raw.IsNull() {
		return
	}

	// If the plan is being destroyed, no need to modify
	if req.Plan.Raw.IsNull() {
		return
	}

	// This is an update - check if any values in the resource are changing
	if !req.Plan.Raw.Equal(req.State.Raw) {
		// Resource is being modified, mark as unknown so any value is accepted
		resp.PlanValue = types.Int64Unknown()
		return
	}

	// No change to the resource, preserve the state value
	resp.PlanValue = req.StateValue
}

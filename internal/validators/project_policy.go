package validators

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/argoproj/argo-cd/v3/util/rbac"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ConfigValidator = projectPolicyValidator{}

// policyComponentCount is the number of comma separated components in a casbin
// policy rule: 'p, sub, res, act, obj, eft'.
const policyComponentCount = 6

// validPolicyActions and validPolicyActionPatterns mirror the equivalent lists
// in Argo CD's AppProject validation. Argo CD does not export them, so they
// have to be kept in sync by hand.
var (
	validPolicyActions = map[string]bool{
		rbac.ActionGet:      true,
		rbac.ActionCreate:   true,
		rbac.ActionUpdate:   true,
		rbac.ActionDelete:   true,
		rbac.ActionSync:     true,
		rbac.ActionOverride: true,
		"*":                 true,
	}

	validPolicyActionPatterns = []*regexp.Regexp{
		regexp.MustCompile("action/.*"),
		regexp.MustCompile("update/.*"),
		regexp.MustCompile("delete/.*"),
	}
)

type projectPolicyValidator struct{}

func (v projectPolicyValidator) Description(_ context.Context) string {
	return "each role policy must be a valid casbin rule of the form 'p, sub, res, act, obj, eft'"
}

func (v projectPolicyValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v projectPolicyValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var metadata types.List

	var spec types.List

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("metadata"), &metadata)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("spec"), &spec)...)

	if resp.Diagnostics.HasError() {
		return
	}

	metadataObject, ok := firstObject(metadata)
	if !ok {
		return
	}

	// Both the expected policy subject and the expected policy object are
	// derived from the project name, so nothing can be checked until it is
	// known.
	projectName, ok := knownString(metadataObject.Attributes()["name"])
	if !ok {
		return
	}

	specObject, ok := firstObject(spec)
	if !ok {
		return
	}

	roles, ok := specObject.Attributes()["role"].(types.Set)
	if !ok || roles.IsNull() || roles.IsUnknown() {
		return
	}

	rolePath := path.Root("spec").AtListIndex(0).AtName("role")

	for _, element := range roles.Elements() {
		role, ok := element.(types.Object)
		if !ok || role.IsNull() || role.IsUnknown() {
			continue
		}

		roleName, ok := knownString(role.Attributes()["name"])
		if !ok {
			continue
		}

		policies, ok := role.Attributes()["policies"].(types.List)
		if !ok || policies.IsNull() || policies.IsUnknown() {
			continue
		}

		for _, policyElement := range policies.Elements() {
			policy, ok := knownString(policyElement)
			if !ok {
				continue
			}

			if err := validateProjectPolicy(projectName, roleName, policy); err != nil {
				resp.Diagnostics.AddAttributeError(rolePath, "Invalid Project Role Policy", err.Error())
			}
		}
	}
}

// ProjectPolicy returns a validator which ensures that every policy configured
// on a project role is a valid Argo CD casbin rule.
func ProjectPolicy() resource.ConfigValidator {
	return projectPolicyValidator{}
}

func firstObject(list types.List) (types.Object, bool) {
	if list.IsNull() || list.IsUnknown() || len(list.Elements()) == 0 {
		return types.Object{}, false
	}

	object, ok := list.Elements()[0].(types.Object)
	if !ok || object.IsNull() || object.IsUnknown() {
		return types.Object{}, false
	}

	return object, true
}

func knownString(value attr.Value) (string, bool) {
	s, ok := value.(types.String)
	if !ok || s.IsNull() || s.IsUnknown() {
		return "", false
	}

	return s.ValueString(), true
}

func validateProjectPolicy(project string, role string, policy string) error {
	policyComponents := strings.Split(policy, ",")
	if len(policyComponents) != policyComponentCount || strings.Trim(policyComponents[0], " ") != "p" {
		return fmt.Errorf("invalid policy rule '%s': must be of the form: 'p, sub, res, act, obj, eft'", policy)
	}

	// subject
	subject := strings.Trim(policyComponents[1], " ")
	expectedSubject := fmt.Sprintf("proj:%s:%s", project, role)

	if subject != expectedSubject {
		return fmt.Errorf("invalid policy rule '%s': policy subject must be: '%s', not '%s'", policy, expectedSubject, subject)
	}

	// resource
	//
	// Project roles may only grant the project-scoped subset of the Argo CD RBAC
	// resources, which Argo CD exports as rbac.ProjectScoped and enforces in
	// AppProject.ValidateProject. Deriving the set from that map keeps the
	// provider from drifting away from the server.
	res := strings.Trim(policyComponents[2], " ")
	if !rbac.ProjectScoped[res] {
		return fmt.Errorf("invalid policy rule '%s': project resource must be one of %s, not '%s'", policy, quotedProjectScopedResources(), res)
	}

	// action
	action := strings.Trim(policyComponents[3], " ")
	if !isValidPolicyAction(action) {
		return fmt.Errorf("invalid policy rule '%s': invalid action '%s'", policy, action)
	}

	// object
	object := strings.Trim(policyComponents[4], " ")

	objectRegexp, err := regexp.Compile(fmt.Sprintf(`^%s(/[*\w-.]+){1,2}$`, project))
	if err != nil || !objectRegexp.MatchString(object) {
		return fmt.Errorf("invalid policy rule '%s': object must be of form '%s/*' or '%s/<APPNAME>' or '%s/<NS>/<APPNAME>', not '%s'", policy, project, project, project, object)
	}

	// effect
	effect := strings.Trim(policyComponents[5], " ")
	if effect != "allow" && effect != "deny" {
		return fmt.Errorf("invalid policy rule '%s': effect must be: 'allow' or 'deny'", policy)
	}

	return nil
}

func isValidPolicyAction(action string) bool {
	if validPolicyActions[action] {
		return true
	}

	for i := range validPolicyActionPatterns {
		if validPolicyActionPatterns[i].MatchString(action) {
			return true
		}
	}

	return false
}

// quotedProjectScopedResources renders the permitted project-scoped resources in
// a stable order so that the error message does not depend on map iteration.
func quotedProjectScopedResources() string {
	quoted := make([]string, 0, len(rbac.ProjectScoped))

	for _, r := range rbac.Resources {
		if rbac.ProjectScoped[r] {
			quoted = append(quoted, "'"+r+"'")
		}
	}

	return strings.Join(quoted, ", ")
}

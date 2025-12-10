package validators

import (
	"context"
	"fmt"
	"regexp"
	"time"

	argocdtime "github.com/argoproj/pkg/time"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/robfig/cron/v3"
)

// GroupNameValidator returns a validator which ensures that any configured
// attribute value is a valid group name (no commas, newlines, carriage returns, or tabs).
func GroupNameValidator() validator.String {
	return groupNameValidator{}
}

type groupNameValidator struct{}

func (v groupNameValidator) Description(ctx context.Context) string {
	return "value must be a valid group name"
}

func (v groupNameValidator) MarkdownDescription(ctx context.Context) string {
	return "value must be a valid group name"
}

func (v groupNameValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	invalidChars := regexp.MustCompile("[,\n\r\t]")

	if invalidChars.MatchString(value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Group Name",
			fmt.Sprintf("Group '%s' contains invalid characters (comma, newline, carriage return, or tab)", value),
		)
	}
}

// RoleNameValidator returns a validator which ensures that any configured
// attribute value is a valid role name.
func RoleNameValidator() validator.String {
	return roleNameValidator{}
}

type roleNameValidator struct{}

func (v roleNameValidator) Description(ctx context.Context) string {
	return "value must be a valid role name"
}

func (v roleNameValidator) MarkdownDescription(ctx context.Context) string {
	return "value must be a valid role name"
}

func (v roleNameValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	roleNameRegexp := regexp.MustCompile(`^[a-zA-Z0-9]([-_a-zA-Z0-9]*[a-zA-Z0-9])?$`)

	if !roleNameRegexp.MatchString(value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Role Name",
			fmt.Sprintf("Invalid role name '%s'. Must consist of alphanumeric characters, '-' or '_', and must start and end with an alphanumeric character", value),
		)
	}
}

// SyncWindowKindValidator returns a validator which ensures that any configured
// attribute value is either "allow" or "deny".
func SyncWindowKindValidator() validator.String {
	return syncWindowKindValidator{}
}

type syncWindowKindValidator struct{}

func (v syncWindowKindValidator) Description(ctx context.Context) string {
	return "value must be either 'allow' or 'deny'"
}

func (v syncWindowKindValidator) MarkdownDescription(ctx context.Context) string {
	return "value must be either 'allow' or 'deny'"
}

func (v syncWindowKindValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value != "allow" && value != "deny" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Sync Window Kind",
			fmt.Sprintf("Kind '%s' mismatch: can only be allow or deny", value),
		)
	}
}

// SyncWindowScheduleValidator returns a validator which ensures that any configured
// attribute value is a valid cron schedule.
func SyncWindowScheduleValidator() validator.String {
	return syncWindowScheduleValidator{}
}

type syncWindowScheduleValidator struct{}

func (v syncWindowScheduleValidator) Description(ctx context.Context) string {
	return "value must be a valid cron schedule"
}

func (v syncWindowScheduleValidator) MarkdownDescription(ctx context.Context) string {
	return "value must be a valid cron schedule"
}

func (v syncWindowScheduleValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	specParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

	if _, err := specParser.Parse(value); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Cron Schedule",
			fmt.Sprintf("cannot parse schedule '%s': %s", value, err.Error()),
		)
	}
}

// SyncWindowDurationValidator returns a validator which ensures that any configured
// attribute value is a valid ArgoCD duration.
func SyncWindowDurationValidator() validator.String {
	return syncWindowDurationValidator{}
}

type syncWindowDurationValidator struct{}

func (v syncWindowDurationValidator) Description(ctx context.Context) string {
	return "value must be a valid ArgoCD duration"
}

func (v syncWindowDurationValidator) MarkdownDescription(ctx context.Context) string {
	return "value must be a valid ArgoCD duration"
}

func (v syncWindowDurationValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if _, err := argocdtime.ParseDuration(value); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Duration",
			fmt.Sprintf("cannot parse duration '%s': %s", value, err.Error()),
		)
	}
}

// SyncWindowTimezoneValidator returns a validator which ensures that any configured
// attribute value is a valid timezone.
func SyncWindowTimezoneValidator() validator.String {
	return syncWindowTimezoneValidator{}
}

type syncWindowTimezoneValidator struct{}

func (v syncWindowTimezoneValidator) Description(ctx context.Context) string {
	return "value must be a valid timezone"
}

func (v syncWindowTimezoneValidator) MarkdownDescription(ctx context.Context) string {
	return "value must be a valid timezone"
}

func (v syncWindowTimezoneValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if _, err := time.LoadLocation(value); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Timezone",
			fmt.Sprintf("cannot parse timezone '%s': %s", value, err.Error()),
		)
	}
}

package argocd

import (
	"fmt"
	"github.com/argoproj/pkg/time"
	"github.com/robfig/cron"
	"golang.org/x/crypto/ssh"
	apiValidation "k8s.io/apimachinery/pkg/api/validation"
	utilValidation "k8s.io/apimachinery/pkg/util/validation"
	"regexp"
	"strings"
)

func validateMetadataLabels(value interface{}, key string) (ws []string, es []error) {
	m := value.(map[string]interface{})
	for k, v := range m {
		for _, msg := range utilValidation.IsQualifiedName(k) {
			es = append(es, fmt.Errorf("%s (%q) %s", key, k, msg))
		}
		val, isString := v.(string)
		if !isString {
			es = append(es, fmt.Errorf("%s.%s (%#v): Expected value to be string", key, k, v))
			return
		}
		for _, msg := range utilValidation.IsValidLabelValue(val) {
			es = append(es, fmt.Errorf("%s (%q) %s", key, val, msg))
		}
	}
	return
}

func validateMetadataAnnotations(value interface{}, key string) (ws []string, es []error) {
	m := value.(map[string]interface{})
	for k := range m {
		errors := utilValidation.IsQualifiedName(strings.ToLower(k))
		if len(errors) > 0 {
			for _, e := range errors {
				es = append(es, fmt.Errorf("%s (%q) %s", key, k, e))
			}
		}
	}
	return
}

func validateMetadataName(value interface{}, key string) (ws []string, es []error) {
	v := value.(string)
	errors := apiValidation.NameIsDNSSubdomain(v, false)
	if len(errors) > 0 {
		for _, err := range errors {
			es = append(es, fmt.Errorf("%s %s", key, err))
		}
	}
	return
}

func validateRoleName(value interface{}, key string) (ws []string, es []error) {
	v := value.(string)
	roleNameRegexp := regexp.MustCompile(`^[a-zA-Z0-9]([-_a-zA-Z0-9]*[a-zA-Z0-9])?$`)
	if !roleNameRegexp.MatchString(v) {
		es = append(es, fmt.Errorf("%s: invalid role name '%s'. Must consist of alphanumeric characters, '-' or '_', and must start and end with an alphanumeric character", key, v))
	}
	return
}

func validateGroupName(value interface{}, key string) (ws []string, es []error) {
	v := value.(string)
	invalidChars := regexp.MustCompile("[,\n\r\t]")
	if strings.TrimSpace(v) == "" {
		es = append(es, fmt.Errorf("%s: group '%s' is empty", key, v))
	}
	if invalidChars.MatchString(v) {
		es = append(es, fmt.Errorf("%s: group '%s' contains invalid characters", key, v))
	}
	return
}

func validateSyncWindowKind(value interface{}, key string) (ws []string, es []error) {
	v := value.(string)
	if v != "allow" && v != "deny" {
		es = append(es, fmt.Errorf("%s: kind '%s' mismatch: can only be allow or deny", key, v))
	}
	return
}

func validateSyncWindowSchedule(value interface{}, key string) (ws []string, es []error) {
	v := value.(string)
	specParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := specParser.Parse(v)
	if err != nil {
		es = append(es, fmt.Errorf("%s: cannot parse schedule '%s': %s", key, v, err))
	}
	return
}

func validateSyncWindowDuration(value interface{}, key string) (ws []string, es []error) {
	v := value.(string)
	_, err := time.ParseDuration(v)
	if err != nil {
		es = append(es, fmt.Errorf("%s: cannot parse duration '%s': %s", key, v, err))
	}
	return
}

func validateDuration(value interface{}, key string) (ws []string, es []error) {
	v := value.(string)
	if _, err := time.ParseDuration(v); err != nil {
		es = append(es, fmt.Errorf("%s: invalid duration '%s': %s", key, v, err))
	}
	return
}

func validateSSHPrivateKey(value interface{}, key string) (ws []string, es []error) {
	v := value.(string)
	if _, err := ssh.ParsePrivateKey([]byte(v)); err != nil {
		es = append(es, fmt.Errorf("%s: invalid ssh private key: %s", key, err))
	}
	return
}

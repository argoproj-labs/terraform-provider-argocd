package argocd

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	_ "time/tzdata"

	apiValidation "k8s.io/apimachinery/pkg/api/validation"
	utilValidation "k8s.io/apimachinery/pkg/util/validation"
)

func validateMetadataLabels(isAppSet bool) func(value interface{}, key string) (ws []string, es []error) {
	return func(value interface{}, key string) (ws []string, es []error) {
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

			if isAppSet && strings.HasPrefix(val, "{{") && strings.HasSuffix(val, "}}") {
				return
			}

			for _, msg := range utilValidation.IsValidLabelValue(val) {
				es = append(es, fmt.Errorf("%s (%q) %s", key, val, msg))
			}
		}

		return
	}
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

func validateDuration(value interface{}, key string) (ws []string, es []error) {
	v := value.(string)

	if _, err := time.ParseDuration(v); err != nil {
		es = append(es, fmt.Errorf("%s: invalid duration '%s': %s", key, v, err))
	}

	return
}

func validateIntOrStringPercentage(value interface{}, key string) (ws []string, es []error) {
	v := value.(string)

	positiveIntegerOrPercentageRegexp := regexp.MustCompile(`^[+]?\d+?%?$`)

	if !positiveIntegerOrPercentageRegexp.MatchString(v) {
		es = append(es, fmt.Errorf("%s: invalid input '%s'. String input must match a positive integer (e.g. '100') or percentage (e.g. '20%%')", key, v))
	}

	return
}

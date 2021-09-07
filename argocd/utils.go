package argocd

import (
	"fmt"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	application "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v2/util/io"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func convertStringToInt64(s string) (i int64, err error) {
	i, err = strconv.ParseInt(s, 10, 64)
	return
}

func convertInt64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}

func convertInt64PointerToString(i *int64) string {
	return strconv.FormatInt(*i, 10)
}

func convertStringToInt64Pointer(s string) (*int64, error) {
	i, err := convertStringToInt64(s)
	if err != nil {
		return nil, fmt.Errorf("not a valid int64: %s", s)
	}
	return &i, nil
}
func isKeyInMap(key string, d map[string]interface{}) bool {
	if d == nil {
		return false
	}
	for k := range d {
		if k == key {
			return true
		}
	}
	return false
}

func expandStringMap(m map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		result[k] = v.(string)
	}
	return result
}

func expandStringList(l []interface{}) (
	result []string) {
	for _, p := range l {
		result = append(result, p.(string))
	}
	return
}

func isValidPolicyAction(action string) bool {
	var validActions = map[string]bool{
		"get":      true,
		"create":   true,
		"update":   true,
		"delete":   true,
		"sync":     true,
		"override": true,
		"*":        true,
	}
	var validActionPatterns = []*regexp.Regexp{
		regexp.MustCompile("action/.*"),
	}

	if validActions[action] {
		return true
	}
	for i := range validActionPatterns {
		if validActionPatterns[i].MatchString(action) {
			return true
		}
	}
	return false
}

func validatePolicy(project string, role string, policy string) error {
	policyComponents := strings.Split(policy, ",")
	if len(policyComponents) != 6 || strings.Trim(policyComponents[0], " ") != "p" {
		return fmt.Errorf("invalid policy rule '%s': must be of the form: 'p, sub, res, act, obj, eft'", policy)
	}
	// subject
	subject := strings.Trim(policyComponents[1], " ")
	expectedSubject := fmt.Sprintf("proj:%s:%s", project, role)
	if subject != expectedSubject {
		return fmt.Errorf("invalid policy rule '%s': policy subject must be: '%s', not '%s'", policy, expectedSubject, subject)
	}
	// resource
	resource := strings.Trim(policyComponents[2], " ")
	if resource != "applications" {
		return fmt.Errorf("invalid policy rule '%s': project resource must be: 'applications', not '%s'", policy, resource)
	}
	// action
	action := strings.Trim(policyComponents[3], " ")
	if !isValidPolicyAction(action) {
		return fmt.Errorf("invalid policy rule '%s': invalid action '%s'", policy, action)
	}
	// object
	object := strings.Trim(policyComponents[4], " ")
	objectRegexp, err := regexp.Compile(fmt.Sprintf(`^%s/[*\w-.]+$`, project))
	if err != nil || !objectRegexp.MatchString(object) {
		return fmt.Errorf("invalid policy rule '%s': object must be of form '%s/*' or '%s/<APPNAME>', not '%s'", policy, project, project, object)
	}
	// effect
	effect := strings.Trim(policyComponents[5], " ")
	if effect != "allow" && effect != "deny" {
		return fmt.Errorf("invalid policy rule '%s': effect must be: 'allow' or 'deny'", policy)
	}
	return nil
}

func isValidToken(token *application.JWTToken, expiresIn int64) error {
	// Check token expiry
	if expiresIn > 0 && token.ExpiresAt < time.Now().Unix() {
		return fmt.Errorf("token has expired")
	}
	// Check that token login works
	opts := apiClientConnOpts
	opts.AuthToken = token.String()
	opts.Insecure = true
	c, err := apiclient.NewClient(&opts)
	if err != nil {
		return err
	}
	closer, _, err := c.NewProjectClient()
	if err != nil {
		return err
	}
	defer io.Close(closer)
	return nil
}

func persistToState(key string, data interface{}, d *schema.ResourceData) error {
	if err := d.Set(key, data); err != nil {
		return fmt.Errorf("error persisting %s: %s", key, err)
	}
	return nil
}

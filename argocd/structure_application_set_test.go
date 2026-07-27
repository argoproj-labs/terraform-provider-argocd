package argocd

import (
	"reflect"
	"testing"
)

func TestExpandFlattenApplicationSetPullRequestGeneratorValues(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"values": map[string]interface{}{
			"foo": "bar",
			"env": "dev",
		},
	}

	asg, err := expandApplicationSetPullRequestGeneratorGenerator(input, false, false)
	if err != nil {
		t.Fatalf("unexpected error expanding pull request generator: %s", err)
	}

	expectedValues := map[string]string{
		"foo": "bar",
		"env": "dev",
	}

	if !reflect.DeepEqual(asg.PullRequest.Values, expectedValues) {
		t.Fatalf("expected expanded Values %v, got %v", expectedValues, asg.PullRequest.Values)
	}

	flattened := flattenApplicationSetPullRequestGenerator(asg.PullRequest)
	if len(flattened) != 1 {
		t.Fatalf("expected exactly one flattened pull request generator, got %d", len(flattened))
	}

	if !reflect.DeepEqual(flattened[0]["values"], expectedValues) {
		t.Fatalf("expected flattened values %v, got %v", expectedValues, flattened[0]["values"])
	}
}

package argocd

import (
	"reflect"
	"testing"
)

func TestExpandApplicationSourceHelmNilValueFile(t *testing.T) {
	t.Parallel()

	input := []interface{}{
		map[string]interface{}{
			"value_files": []interface{}{"values.yaml", nil},
		},
	}

	result := expandApplicationSourceHelm(input)

	expected := []string{"values.yaml"}
	if !reflect.DeepEqual(result.ValueFiles, expected) {
		t.Fatalf("expected ValueFiles %v, got %v", expected, result.ValueFiles)
	}
}

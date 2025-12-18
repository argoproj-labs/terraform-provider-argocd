package argocd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMetadataIsInternalKey(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Key      string
		Expected bool
	}{
		{"", false},
		{"anyKey", false},
		{"any.hostname.io", false},
		{"any.hostname.com/with/path", false},
		{"any.kubernetes.io", true},
		{"kubernetes.io", true},
		{"notified.notifications.argoproj.io", true},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()

			isInternal := metadataIsInternalKey(tc.Key)
			if tc.Expected && isInternal != tc.Expected {
				t.Fatalf("Expected %q to be internal", tc.Key)
			}

			if !tc.Expected && isInternal != tc.Expected {
				t.Fatalf("Expected %q not to be internal", tc.Key)
			}
		})
	}
}

func TestMetadataFilterFinalizers(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                 string
		apiFinalizers        []string
		configuredFinalizers []interface{}
		expected             []string
	}{
		{
			name:                 "empty lists",
			apiFinalizers:        []string{},
			configuredFinalizers: []interface{}{},
			expected:             []string{},
		},
		{
			name:                 "no configured finalizers",
			apiFinalizers:        []string{"system.finalizer", "user.finalizer"},
			configuredFinalizers: []interface{}{},
			expected:             []string{},
		},
		{
			name:                 "only configured finalizers returned",
			apiFinalizers:        []string{"system.finalizer", "user.finalizer", "another.user.finalizer"},
			configuredFinalizers: []interface{}{"user.finalizer", "another.user.finalizer"},
			expected:             []string{"user.finalizer", "another.user.finalizer"},
		},
		{
			name:                 "configured finalizer not in API response",
			apiFinalizers:        []string{"system.finalizer"},
			configuredFinalizers: []interface{}{"user.finalizer"},
			expected:             []string{},
		},
		{
			name:                 "mixed scenario - system and user finalizers",
			apiFinalizers:        []string{"resources.argoproj.io/finalizer", "user.custom/finalizer", "kubernetes.io/finalizer"},
			configuredFinalizers: []interface{}{"user.custom/finalizer"},
			expected:             []string{"user.custom/finalizer"},
		},
		{
			name:                 "invalid type in configured finalizers",
			apiFinalizers:        []string{"system.finalizer", "user.finalizer"},
			configuredFinalizers: []interface{}{"user.finalizer", 123, nil},
			expected:             []string{"user.finalizer"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := metadataFilterFinalizers(tc.apiFinalizers, tc.configuredFinalizers)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestToStringSet(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    interface{}
		expected map[string]bool
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: map[string]bool{},
		},
		{
			name:     "empty list",
			input:    []interface{}{},
			expected: map[string]bool{},
		},
		{
			name:     "list with strings",
			input:    []interface{}{"foo", "bar", "baz"},
			expected: map[string]bool{"foo": true, "bar": true, "baz": true},
		},
		{
			name:     "list with duplicates",
			input:    []interface{}{"foo", "foo", "bar"},
			expected: map[string]bool{"foo": true, "bar": true},
		},
		{
			name:     "list with non-strings ignored",
			input:    []interface{}{"foo", 123, nil, "bar"},
			expected: map[string]bool{"foo": true, "bar": true},
		},
		{
			name:     "non-list input returns empty",
			input:    "not a list",
			expected: map[string]bool{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := toStringSet(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}

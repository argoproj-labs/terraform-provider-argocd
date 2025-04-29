package argocd

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_validateMetadataLabels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		isAppSet bool
		value    interface{}
		key      string
		wantWs   []string
		wantEs   []error
	}{
		{
			name:     "Valid labels",
			isAppSet: false,
			value: map[string]interface{}{
				"valid-key": "valid-value",
			},
			key:    "metadata_labels",
			wantWs: nil,
			wantEs: nil,
		},
		{
			name:     "Invalid label key",
			isAppSet: false,
			value: map[string]interface{}{
				"Invalid Key!": "valid-value",
			},
			key:    "metadata_labels",
			wantWs: nil,
			wantEs: []error{
				fmt.Errorf("metadata_labels (\"Invalid Key!\") name part must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyName',  or 'my.name',  or '123-abc', regex used for validation is '([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]')"),
			},
		},
		{
			name:     "Invalid label value",
			isAppSet: false,
			value: map[string]interface{}{
				"valid-key": "Invalid Value!",
			},
			key:    "metadata_labels",
			wantWs: nil,
			wantEs: []error{
				fmt.Errorf("metadata_labels (\"Invalid Value!\") a valid label must be an empty string or consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyValue',  or 'my_value',  or '12345', regex used for validation is '(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?')"),
			},
		},
		{
			name:     "Non-string label value",
			isAppSet: false,
			value: map[string]interface{}{
				"valid-key": 123,
			},
			key:    "metadata_labels",
			wantWs: nil,
			wantEs: []error{
				fmt.Errorf("metadata_labels.valid-key (123): Expected value to be string"),
			},
		},
		{
			name:     "Valid templated value for AppSet",
			isAppSet: true,
			value: map[string]interface{}{
				"valid-key": "{{ valid-template }}",
			},
			key:    "metadata_labels",
			wantWs: nil,
			wantEs: nil,
		},
		{
			name:     "Invalid templated value for non-AppSet",
			isAppSet: false,
			value: map[string]interface{}{
				"valid-key": "{{ invalid-template }}",
			},
			key:    "metadata_labels",
			wantWs: nil,
			wantEs: []error{
				fmt.Errorf("metadata_labels (\"{{ invalid-template }}\") a valid label must be an empty string or consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyValue',  or 'my_value',  or '12345', regex used for validation is '(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?')"),
			},
		},
		{
			name:     "Empty label key",
			isAppSet: false,
			value: map[string]interface{}{
				"": "valid-value",
			},
			key:    "metadata_labels",
			wantWs: nil,
			wantEs: []error{
				fmt.Errorf("metadata_labels (\"\") name part must be non-empty"),
				fmt.Errorf("metadata_labels (\"\") name part must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyName',  or 'my.name',  or '123-abc', regex used for validation is '([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]')"),
			},
		},
		{
			name:     "Empty label value",
			isAppSet: false,
			value: map[string]interface{}{
				"valid-key": "",
			},
			key:    "metadata_labels",
			wantWs: nil,
			wantEs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotWs, gotEs := validateMetadataLabels(tt.isAppSet)(tt.value, tt.key)

			require.Equal(t, tt.wantWs, gotWs)
			require.Equal(t, tt.wantEs, gotEs)
		})
	}
}

func Test_validateSSHPrivateKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		value  interface{}
		wantWs []string
		wantEs []error
	}{
		{
			name:   "Invalid ssh private key",
			value:  "foo",
			wantWs: nil,
			wantEs: []error{fmt.Errorf("ssh_private_key: invalid ssh private key: ssh: no key found")},
		},
		{
			name:   "Valid ssh private key",
			value:  mustGenerateSSHPrivateKey(t),
			wantWs: nil,
			wantEs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotWs, gotEs := validateSSHPrivateKey(tt.value, "ssh_private_key")

			if !reflect.DeepEqual(gotWs, tt.wantWs) {
				t.Errorf("validateSSHPrivateKey() gotWs = %v, want %v", gotWs, tt.wantWs)
			}

			if !reflect.DeepEqual(gotEs, tt.wantEs) {
				t.Errorf("validateSSHPrivateKey() gotEs = %v, want %v", gotEs, tt.wantEs)
			}
		})
	}
}

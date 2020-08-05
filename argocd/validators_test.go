package argocd

import (
	"fmt"
	"reflect"
	"testing"
)

func Test_validateSSHPrivateKey(t *testing.T) {

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

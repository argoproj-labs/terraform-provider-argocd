package provider

import (
	"testing"

	"github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/stretchr/testify/assert"
)

// A running application's SyncPolicy.Automated fields (Prune, SelfHeal,
// AllowEmpty) are *bool and are frequently nil when not explicitly set by
// the user, e.g. right after an out-of-band sync triggered outside Terraform.
// newApplicationSyncPolicyAutomated must not dereference them directly.
func TestNewApplicationSyncPolicyAutomated_NilFields(t *testing.T) {
	t.Parallel()

	spa := &v1alpha1.SyncPolicyAutomated{}

	assert.NotPanics(t, func() {
		result := newApplicationSyncPolicyAutomated(spa)

		assert.False(t, result.AllowEmpty.ValueBool())
		assert.False(t, result.Prune.ValueBool())
		assert.False(t, result.SelfHeal.ValueBool())
	})
}

func TestNewApplicationSyncPolicyAutomated_SetFields(t *testing.T) {
	t.Parallel()

	allowEmpty, prune, selfHeal := true, true, true
	spa := &v1alpha1.SyncPolicyAutomated{
		AllowEmpty: &allowEmpty,
		Prune:      &prune,
		SelfHeal:   &selfHeal,
	}

	result := newApplicationSyncPolicyAutomated(spa)

	assert.True(t, result.AllowEmpty.ValueBool())
	assert.True(t, result.Prune.ValueBool())
	assert.True(t, result.SelfHeal.ValueBool())
}

func TestNewApplicationSyncPolicyAutomated_Nil(t *testing.T) {
	t.Parallel()

	assert.Nil(t, newApplicationSyncPolicyAutomated(nil))
}

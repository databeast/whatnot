package whatnot

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNamespaceManager(t *testing.T) {
	t.Run("register namespace manager with options", newManagerWithOptions)
}

func newManagerWithOptions(t *testing.T) {
	manager := NewNamespaceManager(WithAcls, WithRaft, WithTrace, WithGossip, WithDeadlockBreak, WithLogger{createTestLogger(t)})
	err := manager.RegisterNamespace(NewNamespace("test"))
	if !assert.Nil(t, err, "registering namespace failed") {
		t.Error(err.Error())
		return
	}
}

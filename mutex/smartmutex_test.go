package mutex

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeadlock(t *testing.T) {
	Opts.Tracing = true
	m1 := New("m1")

	m1.Lock()
	assert.True(t, m1.IsLocked(), "mutex does not declare itself as locked")
	m1.Unlock()
	assert.False(t, m1.IsLocked(), "mutex does not declare itself as unlocked")

}

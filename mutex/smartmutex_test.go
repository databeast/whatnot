package mutex

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestLockQueingStats(t *testing.T) {
	//TODO: Still have a race condition here - not critical since this functionality is informational only
	Opts.Tracing = true
	m1 := New("m1")

	go func() {
		m1.Lock()
	}()

	go func() {
		m1.Lock()
	}()

	go func() {
		m1.Lock()
	}()

	go func() {
		m1.Lock()
	}()

	time.Sleep(time.Second)
	assert.Equal(t, 3, m1.Queue(), "did not indicate 3 waiting locks in queue")

}

func TestLockStatus(t *testing.T) {
	Opts.Tracing = true
	m1 := New("m1")

	m1.Lock()
	assert.True(t, m1.IsLocked(), "mutex does not declare itself as locked")
	m1.Unlock()
	assert.False(t, m1.IsLocked(), "mutex does not declare itself as unlocked")
	m1.SoftLock()
	assert.False(t, m1.IsLocked(), "mutex does not declare itself as unlocked")
}

func TestDeadlock(t *testing.T)  {
	var detected bool
	Opts.Tracing = true
	Opts.OnPotentialDeadlock= func(){t.Log("received deadlock detection") ; detected = true}
	m1 := New("m1")
	t.Log("forcing a recursive lock")
	go func() {
		m1.Lock()
		m1.Lock()
	}()
	time.Sleep(time.Second)
	assert.True(t, detected, "deadlock not detected")
}
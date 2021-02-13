package mutex

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDeadlock(t *testing.T) {
	Opts.Tracing = true
	m1 := New("m1")

	m1.Lock()
	assert.True(t, m1.IsLocked(), "mutex does not declare itself as locked")
	m1.Unlock()
	assert.False(t, m1.IsLocked(), "mutex does not declare itself as unlocked")


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

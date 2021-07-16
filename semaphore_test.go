package whatnot

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSemaphoreClaim(t *testing.T) {
	elem := PathElement{
		logsupport:       logsupport{},
		mu:               nil,
		section:          "",
		parent:           nil,
		children:         nil,
		parentnotify:     nil,
		subevents:        nil,
		selfnotify:       nil,
		reslock:          resourceLock{},
		resval:           ElementValue{},
		subscriberNotify: nil,
		prunetracker:     nil,
		prunectx:         nil,
		prunefunc:        nil,
		semaphores:       nil,
	}
	err := elem.CreateSemaphorePool(false, false, SemaphorePoolOpts{PoolSize: 10})

	if !assert.Nil(t, err) {
		t.Error(err)
	}
	claims := []*SemaphoreClaim{}

	for i := 1 ; i <= 10 ; i ++ {
		claim, err := elem.semaphores.ClaimSingle(time.Second)
		if !assert.Nil(t, err) {
			t.Error(err)
		}
		t.Logf("claiming semaphore %d", i)
		claims = append(claims, claim)
	}
	t.Log("waiting for semaphore claim 11 to fail after timeout")
	_, err = elem.semaphores.ClaimSingle(time.Second)
	assert.NotNilf(t, err, "semaphore claim did not time out")
	for i, c := range claims {
		err = c.Return()
		if !assert.Nil(t, err) {
			t.Error(err)
		}
		t.Logf("Releasing Semaphore %d", i+1)
	}


}
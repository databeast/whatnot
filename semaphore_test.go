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
		claims = append(claims, claim)
	}
	for _, c := range claims {
		err = c.Return()
		if !assert.Nil(t, err) {
			t.Error(err)
		}
	}


}
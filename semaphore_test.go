package whatnot

import (
	"context"
	"testing"
	"time"

	"github.com/databeast/whatnot/mutex"
	"github.com/stretchr/testify/assert"
)

func TestClaimOverMax(t *testing.T) {
	elem := PathElement{}
	err := elem.CreateSemaphorePool(false, false, SemaphorePoolOpts{PoolSize: 10})
	assert.Nil(t, err)
	claim, err := elem.semaphores.Claim(context.Background(), 11)
	assert.Nil(t, claim)
	assert.NotNil(t, err)
}

func TestSemaphoreClaim(t *testing.T) {
	elem := PathElement{}
	err := elem.CreateSemaphorePool(false, false, SemaphorePoolOpts{PoolSize: 10})

	if !assert.Nil(t, err) {
		t.Error(err)
	}
	claims := []*SemaphoreClaim{}
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for i := 1; i <= 10; i++ {
		claim, err := elem.semaphores.Claim(timeout, 1)
		if !assert.Nil(t, err) {
			t.Error(err)
		}
		t.Logf("claiming semaphore %d", i)
		claims = append(claims, claim)
	}
	t.Log("waiting for semaphore claim 11 to fail after timeout")
	_, err = elem.semaphores.Claim(timeout, 1)
	assert.NotNilf(t, err, "semaphore claim did not time out")
	for i, c := range claims {
		err = c.Return()
		if !assert.Nil(t, err) {
			t.Error(err)
		}
		t.Logf("Releasing Semaphore %d", i+1)
	}
}

func TestPoolAcrossPrefixChildren(t *testing.T) {
	mutex.Opts.DisableDeadlockDetection = true
	nsm, _ := NewNamespaceManager()
	_ = nsm.RegisterNamespace(NewNamespace("test"))
	ns, _ := nsm.FetchNamespace("test")

	elem, err := ns.FetchOrCreateAbsolutePath("/path/to/test/data")

	//THIS IS DEADLOCKING

	t.Log("creating children from prefix")
	e1, err := elem.Add("sub1")
	assert.NotNil(t, e1)
	assert.Nil(t, err)
	e2, err := elem.Add("sub2")
	assert.NotNil(t, e2)
	assert.Nil(t, err)
	e3, err := elem.Add("sub3")
	assert.NotNil(t, e3)
	assert.Nil(t, err)

	err = elem.CreateSemaphorePool(true, false, SemaphorePoolOpts{PoolSize: 10})
	if !assert.Nil(t, err) {
		t.Error(err)
	}
	t.Log("created shared semaphore pool across children")

}

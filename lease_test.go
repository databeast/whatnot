package whatnot

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestLeaseCreationAndExpiration(t *testing.T) {
	t.Run("Create a new lease", createNewLeaseOnPathElement)
	t.Run("Test that Lease expires after set time", leaseExpiresAsExpected)
}

func createNewLeaseOnPathElement(t *testing.T) {
	t.Log("Creating a new lease on an existing Path Element")
	gns := createTestNamespace(t)
	err := gns.RegisterAbsolutePath(PathString("/testelement").ToAbsolutePath())
	if !assert.Nil(t, err, "registering path returned error") {
		t.Error(err.Error())
		return
	}
	elem := gns.FetchAbsolutePath("/testelement")
	_, cancel, err := elem.LockWithLease(time.Second * 2)
	if !assert.Nil(t, err, "creating lease returned error") {
		t.Error(err.Error())
		return
	}
	t.Log("created one-second lease on test element")
	cancel()
	t.Log("cancelled one-second lease on test element")
}

func leaseExpiresAsExpected(t *testing.T) {
	t.Log("Checking that lease expires at expected duration")

	gns := createTestNamespace(t)
	err := gns.RegisterAbsolutePath(PathString("/testelement").ToAbsolutePath())
	if !assert.Nil(t, err, "registering path returned error") {
		t.Error(err.Error())
		return
	}
	elem := gns.FetchAbsolutePath("/testelement")

	leaseFor := time.Second

	ctx, cancel, err := elem.LockWithLease(leaseFor)
	if !assert.Nil(t, err, "creating lease returned error") {
		t.Error(err.Error())
		return
	}
	defer cancel()
	t.Log("created one-second lease on test element")

	t.Log("measuring time for lease to expire")

	start := time.Now()
	<-ctx.Done()
	t.Log("lease returned before time expired")
	took :=  time.Now().Sub(start)
	if !assert.Equal(t, leaseFor.Round(time.Second).Seconds(), took.Round(time.Second).Seconds(), "did not expire in expected time") {
		t.Error("did not expire in expected time")
		return
	}
	t.Logf("took %f seconds to expire", took.Round(time.Second).Seconds())
	return

}

func leaseAcceptsCancelation(t *testing.T) {
	t.Log("Creating a lease and then cancelling it prematurely")

}

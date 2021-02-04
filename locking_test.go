package whatnot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestElementLocking(t *testing.T) {
	t.Run("Lock Single Element", lockSingleElement)
	t.Run("Lock Element Prefix", lockElementPrefix)
	t.Run("Queue Locks on single element", lockandUnlockSingleElement)
}

func lockSingleElement(t *testing.T) {
	t.Log("creating a Path Element and applying a lock to it")
	gns := createTestNamespace(t)

	testPathString := PathString("/path/to/test/data")
	testpath := testPathString.ToAbsolutePath()
	err := gns.RegisterAbsolutePath(testpath)
	if !assert.Nil(t, err, "registerabsolute path returned error") {
		t.Error(err.Error())
		return
	}
	lastElement := gns.FetchAbsolutePath(testPathString)
	if !assert.NotNil(t, lastElement, "did not return last element in absolute path") {
		t.Error("did not find registered path element")
		return
	}
	t.Log("Locking Path Element")
	lastElement.Lock()
	t.Log("Unlocking Path Element")
	lastElement.UnLock()
}

func lockElementPrefix(t *testing.T) {
	t.Log("creating an Absolute Path of Elements, locking the entire prefix")

	gns := createTestNamespace(t)

	testPathString := PathString("/path/to/test/data")
	testpath := testPathString.ToAbsolutePath()
	err := gns.RegisterAbsolutePath(testpath)
	if !assert.Nil(t, err, "registerabsolute path returned error") {
		t.Error(err.Error())
		return
	}
	// grab the top-level element from our created path
	firstElement := gns.FetchAbsolutePath("/path")
	if !assert.NotNil(t, firstElement, "did not return first element in absolute path") {
		t.Error("did not find registered path element")
		return
	}
	t.Log("Locking Path Element and children")
	firstElement.LockSubs()
	t.Log("Unlocking Path Element and children")
	firstElement.UnLockSubs()
}

func lockandUnlockSingleElement(t *testing.T) {
	t.Log("Creating a Path Element, locking it, and testing it remains locked until manually unlocked")
}

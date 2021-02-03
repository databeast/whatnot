package whatnot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWatchNotifications(t *testing.T) {
	t.Run("Create new watch subscription on PathElement", createNewWatchOnElement)
	t.Run("changing a path element creates notification", changeElementAndNotify)
}

func createNewWatchOnElement(t *testing.T) {
	t.Log("creating a new watch on existing path element")
	const testNameSpace = "globaltest"
	t.Log("Creating Namespace Manager")
	manager = NewNamespaceManager()
	gns := NewNamespace(testNameSpace)
	err := manager.RegisterNamespace(gns)
	if !assert.Nil(t, err, "RegisterNamespace returned error") {
		return
	}
	testPathString := PathString("/path/to/test/data")
	testpath := testPathString.ToAbsolutePath()
	err = gns.RegisterAbsolutePath(testpath)
	if !assert.Nil(t, err, "registerabsolute path returned error") {
		t.Error(err.Error())
		return
	}

	elem := gns.FetchAbsolutePath(testPathString)

	sub := elem.SubscribeToEvents(false)
	select {
	case <-sub.Events():
	default:
		// we instantly pass here, because there is no waiting message on the channel
	}

}

func changeElementAndNotify(t *testing.T) {
	t.Log("testing that a notification is received when modifying a watched element")

}

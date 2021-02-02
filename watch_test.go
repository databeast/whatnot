package whatnot

import (
	"testing"

	"github.com/databeast/whatnot/mutex"
	"github.com/stretchr/testify/assert"
)

func TestWatchNotifications(t *testing.T) {
	t.Run("changing a path element creates notification", changeElementAndNotify)
}

func createNewWatchOnElement(t *testing.T) {
	t.Log("creating a new watch on existing path element")
}

func changeElementAndNotify(t *testing.T) {
	t.Log("testing that a notification is received when modifying a watched element")
	elem := PathElement{
		section:          SubPath("testelement"),
		parent:           nil,
		parentnotify:     nil,
		mu:               &mutex.SmartMutex{},
		subscriberNotify: nil,
	}
	sub := elem.SubscribeToEvents(false)
	select {
	case e := <-sub.Events():
		assert.Equal(t, sub, e.sub, "returned modified element was not original element")
	default:

	}

}

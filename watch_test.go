package whatnot

import (
	"testing"
)

func TestWatchNotifications(t *testing.T) {
	t.Run("changing a path element creates notification", changeElementAndNotify)
}

func createNewWatchOnElement(t *testing.T) {
	t.Log("creating a new watch on existing path element")
}

func changeElementAndNotify(t *testing.T) {
	t.Log("testing that a notification is received when modifying a watched element")

}

package whatnot

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestElementIsPrunedAfterDuration(t *testing.T) {
	gns := createTestNamespace(t)

	testPathString := PathString("/path/to/test/data")
	testpath := testPathString.ToAbsolutePath()
	err := gns.RegisterAbsolutePath(testpath)
	if !assert.Nil(t, err, "registerabsolute path returned error") {
		t.Error(err.Error())
		return
	}
	elem := gns.FetchAbsolutePath(testPathString)
	if !assert.NotNil(t, elem, "did not return last element in absolute path") {
		t.Error("did not find registered path element")
		return
	}
	elem.EnablePruningAfter(time.Second)

	events := elem.SubscribeToEvents(true)
	testtimeout, cancel := context.WithTimeout(context.Background(), pruneInterval+time.Minute) // we need to wait for the pruning interval to trigger
	defer cancel()

	select {
	case e := <-events.Events():
		if e.Change == ChangePruned {
			t.Log("element signaled pruning")
		} else {
			t.Errorf("incorrect element change event %d received", e.Change)
		}

	case <-testtimeout.Done():
		t.Error("test timed out before pruning signal received")
	}
}

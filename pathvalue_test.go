package whatnot

import (
	"testing"
	"time"

	"github.com/databeast/whatnot/access"
	"github.com/stretchr/testify/assert"
)

func TestPathElementValues(t *testing.T) {
	t.Run("Set and Retrieve a value", setAndRetrieveValueOnElement)
}

func setAndRetrieveValueOnElement(t *testing.T) {
	t.Log("Creating a Path Element, setting a Value on it and then retrieving that value")
	gns := createTestNamespace(t)
	err := gns.RegisterAbsolutePath(PathString("/testelem").ToAbsolutePath())
	if !assert.Nil(t, err, "registering path returned error") {
		t.Error(err.Error())
		return
	}
	elem := gns.FetchAbsolutePath("/testelem")
	if !assert.NotNil(t, elem, "registered path element is nil") {
		t.Error("couldnt fetch test path element")
		return
	}
	time.Sleep(time.Second)
	elem.SetValue(ElementValue{Val: "test value"}, ChangeAdded, access.Role{})
	val := elem.GetValue()
	assert.Equal(t, ElementValue{Val: "test value"}, val, "retrieved value did not match original value")
}

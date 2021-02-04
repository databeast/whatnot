package whatnot

import (
	"testing"

	"github.com/databeast/whatnot/access"
	"github.com/stretchr/testify/assert"
)

func TestPathElementValues(t *testing.T) {
	t.Run("Set and Retrieve a value", setAndRetrieveValueOnElement)
}

func setAndRetrieveValueOnElement(t *testing.T) {
	t.Log("Creating a Path Element, setting a Value on it and then retrieving that value")
	gns := createTestNamespace(t)
	gns.RegisterAbsolutePath(PathString("testelem").ToAbsolutePath())
	elem := gns.FetchAbsolutePath(PathString("testelem"))
	if !assert.NotNil(t, elem, "registered path element is nil") {
		t.Error("couldnt fetch test path element")
		return
	}
	elem.SetValue(ElementValue{Val: "test value"}, ADDED, access.Role{})
	val := elem.GetValue()
	assert.Equal(t, ElementValue{Val: "test value"}, val)
}

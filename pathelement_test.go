package whatnot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathElements(t *testing.T) {
	t.Run("Append Element to Existing Element", appendPathElement)

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
	if !assert.Equal(t, lastElement.AbsolutePath(), testpath, "returned element absolutepath did not match original") {
		t.Error("mismatched return path")
		return
	}

	extraElement, err := lastElement.AppendRelativePath(PathString("subdata"))
	if !assert.Nil(t, err, "appending another sub-element returned error") {
		t.Error(err.Error())
		return
	}

	if !assert.Equal(t, SubPath("subdata"), extraElement.SubPath(), "did not match original subpath string") {
		t.Error("mismatch between provided and created subpath")
	}

	if !assert.Equal(t, lastElement, extraElement.Parent(), "incorrect parent returned") {
		t.Error("new element did not obtain correct parent element")
		return
	}

	if !assert.Equal(t, extraElement, lastElement.fetchSubElement(SubPath("subdata")), "sub-element not reachable from parent element") {
		t.Error("parent element of created sub-element not linked to new sub element correctly")
	}

}

func appendPathElement(t *testing.T) {
	t.Log("Testing that appending a path element to an existing path element succeeds")
}

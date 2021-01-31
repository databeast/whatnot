package whatnot

import (
	"testing"
)

func TestPathElements(t *testing.T) {
	t.Run("Append Element to Existing Element", appendPathElement)
}

func appendPathElement(t *testing.T) {
	t.Log("Testing that appending a path element to an existing path element succeeds")
}

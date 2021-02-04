package whatnot

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPathConstruction(t *testing.T) {
	t.Run("Test that path string is parsed correctly", parsePathStringToAbsolutePath)
}

func parsePathStringToAbsolutePath(t *testing.T) {
	origpath := PathString("/path/to/my/test/element")
	abspath := origpath.ToAbsolutePath()
	if !assert.Equal(t, origpath, abspath.ToPathString(), "conversion back to string was different") {
		t.Error("converted the absolute path back to a path string produced a difference result")
		return
	}
	assert.Equal(t, SubPath("path"), abspath[0])
	assert.Equal(t, SubPath("to"), abspath[1])
	assert.Equal(t, SubPath("my"), abspath[2])
	assert.Equal(t, SubPath("test"), abspath[3])
	assert.Equal(t, SubPath("element"), abspath[4])

}
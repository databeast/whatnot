package whatnot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var manager *NameSpaceManager

func TestNamespaces(t *testing.T) {
	t.Run("Register a Namespace", registerNewNamespace)
}

func registerNewNamespace(t *testing.T) {
	t.Log("Creating Namespace Manager")
	manager = NewNamespaceManager()
	gns := NewNamespace("global")
	err := manager.RegisterNamespace(gns)
	assert.Nil(t, err, "RegisterNamespace returned error")

}

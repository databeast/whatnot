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
	const testNameSpace = "globaltest"
	t.Log("Creating Namespace Manager")
	manager, err := NewNamespaceManager()
	if !assert.Nil(t, err, "NewNamespaceManager returned error") {
		return
	}
	gns := NewNamespace(testNameSpace)
	err = manager.RegisterNamespace(gns)
	if !assert.Nil(t, err, "RegisterNamespace returned error") {
		return
	}

	fns, err := manager.FetchNamespace(testNameSpace)
	if !assert.Nil(t, err, "FetchNamespace returned error") {
		return
	}
	assert.Equal(t, gns.name, fns.name, "namespace identifiers did not match")

}


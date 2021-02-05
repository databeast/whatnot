package whatnot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testNameSpace = "globaltest"
)

func createTestNamespace(t *testing.T) *Namespace {
	t.Log("Creating Namespace Manager")
	manager = NewNamespaceManager()
	gns := NewNamespace(testNameSpace)
	err := manager.RegisterNamespace(gns)
	if !assert.Nil(t, err, "RegisterNamespace returned error") {
		t.Error("failed to register Test Namespace")
		return nil
	}
	return gns
}

func createTestLogger(t *testing.T) Logger {
	return testlogger{t: t}
}

type testlogger struct {
	t *testing.T
}

func (t testlogger) Debug(msg string) {
	t.t.Log(msg)
}

func (t testlogger) Info(msg string) {
	t.t.Log(msg)
}

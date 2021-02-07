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
	manager, err := NewNamespaceManager( WithLogger{createTestLogger(t)})
	if !assert.Nil(t, err, "NewNamespaceManager returned error") {
		t.Error("failed to create Namespace Manager")
		return nil
	}
	gns := NewNamespace(testNameSpace)
	err = manager.RegisterNamespace(gns)
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

func (t testlogger) Debugf(format string, a ...interface{}) {
	t.t.Logf(format, a...)
}

func (t testlogger) Infof(format string, a ...interface{}) {
	t.t.Logf(format, a...)
}

func (t testlogger) Warn(msg string) {
	t.t.Log(msg)
}

func (t testlogger) Warnf(format string, a ...interface{}) {
	t.t.Logf(format, a...)
}

func (t testlogger) Error(msg string) {
	t.t.Log(msg)
}

func (t testlogger) Errorf(format string, a ...interface{}) {
	t.t.Logf(format, a...)
}


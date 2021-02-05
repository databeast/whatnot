package whatnot

import (
	"fmt"

	"github.com/databeast/whatnot/mutex"
	"github.com/pkg/errors"
)

// NameSpaceManager provides top-level management of unique element namespaces
type NameSpaceManager struct {
	namespaces map[string]*Namespace
	mu         *mutex.SmartMutex
	log        Logger
}

// NewNamespaceManager create a top-level namespace manager, to contain multiple subscribable namespaces
// you probably only want to call this once, to initialize WhatNot, but who am I to tell you what your use cases are
func NewNamespaceManager(opts ...ManagerOption) *NameSpaceManager {
	return &NameSpaceManager{
		mu:         mutex.New(fmt.Sprintf("NameSpace Manager mutex")),
		namespaces: make(map[string]*Namespace),
		log:        nilLogger{},
	}
}

// RegisterNamespace actives a name Namespace into the list of actively available and
// subscribable namespaces
func (m *NameSpaceManager) RegisterNamespace(ns *Namespace) error {

	if _, ok := m.namespaces[ns.name]; ok { // fail if already registered
		return errors.Errorf("refusing to register already-registered name")
	}

	m.mu.Lock()
	m.namespaces[ns.name] = ns
	m.mu.Unlock()
	m.log.Info(fmt.Sprintf("registered new namespace %q", ns.name))

	return nil
}

// UnRegisterNamespace will completely remove a given namespace
// all the properties, leases, subscriptions, etc within it.
func (m *NameSpaceManager) UnRegisterNamespace(ns *Namespace) error {

	// fail if not present
	if _, ok := m.namespaces[ns.name]; !ok { // fail if not present
		return errors.Errorf("refusing to unregister unknown namespace")
	}

	m.mu.Lock()
	delete(m.namespaces, ns.name) // TODO: this needs a better collapsing method than just deleting this reference
	m.mu.Unlock()

	return nil
}

// FetchNamespace gets you access to the requested namespace
// understandably most other operations involving a namespace's contents begin here
func (m *NameSpaceManager) FetchNamespace(name string) (ns *Namespace, err error) {

	if ns, ok := m.namespaces[name]; ok { // fail if not present
		return ns, nil
	} else {
		return nil, errors.Errorf("no such namespaces: %q", name)
	}
}

// WithLogger attaches a Logger to the Namespace manager
// allowing you to insert your own logging solution into Whatnot
func (m *NameSpaceManager) WithLogger(l Logger) *NameSpaceManager {
	m.log = l
	return m

}

type optionName string

const (
	optionDiscoverGossip optionName = "gossip cluster discovery"
	optionSyncRaft       optionName = "raft cluster synchronization"
	optionTrace          optionName = "trace mutex locks"
	optionBreak          optionName = "break mutex deadlocking"
	optionAcls           optionName = "enable element permissions"
)

type ManagerOption func() optionName

var WithGossip = func() optionName {
	return optionDiscoverGossip
}

var WithRaft = func() optionName {
	return optionSyncRaft
}

var WithTrace = func() optionName {
	return optionTrace
}

var WithDeadlockBreak = func() optionName {
	return optionBreak
}

var WithAcls = func() optionName {
	return optionAcls
}

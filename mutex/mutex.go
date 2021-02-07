package mutex

import (
	"fmt"
	"sync"
)

var mutexTracing = true

// SmartMutex is a more extensive Mutex structure
// with Deadlock-detection and metrics of lock aquisition queues
// while most Mutexes in Go should be extremely localized in Scope
// and retain locks for a minimal time
// this Mutex structure is optimized for uses by a great many goroutines
// from multiple code scopes
type SmartMutex struct {
	mu         Mutex
	name       string
	locked     bool
	count      int
	statuslock *sync.Mutex
	countlock  *sync.Mutex
}

// Generate a new Mutex on the given element path
func New(elementpath string) *SmartMutex {
	if elementpath == "" {
		panic("refusing to create a mutex on an empty path")
	}
	m := &SmartMutex{
		mu:         Mutex{},
		countlock:  &sync.Mutex{},
		statuslock: &sync.Mutex{},
	}

	m.name = elementpath
	return m
}

func (m *SmartMutex) releaseDeadlock() {
	m.mu = Mutex{}
}

// SoftLock is intended to test if object is locked, usually to wait
// for a mutex'ed resource to stabilize
// but then just to wait for changes to complete before proceeding
// if you wish to make modifications during during the lock, call
// Lock() and Unlock() explicitly.
func (m *SmartMutex) SoftLock() {
	m.statuslock.Lock()
	status := m.locked

	if status == false {
		m.statuslock.Unlock()
		return
	} else {
		m.statuslock.Unlock()
		if mutexTracing == true {
			m.countlock.Lock()
			m.trace(fmt.Sprintf("pass-through softlock on %s is waiting for %d existing locks to release", m.name, m.count))
			m.countlock.Unlock()
		}
		m.Lock()
		m.Unlock()
		return
	}
}

func (m *SmartMutex) Name() string {
	n := m.name
	return n
}

func (m *SmartMutex) Locked() bool {
	return m.locked
}

// Lock provides deadlock-aware, queue-tracking Mutex Locks
func (m *SmartMutex) Lock() {
	if m == nil {
		panic("tried to lock a nil mutex")
	}

	m.statuslock.Lock()
	if m.locked == true {
		m.countlock.Lock()
		if m.count > 0 {
			m.trace(fmt.Sprintf("waiting for lock on %s to release (%d already in queue)", m.name, m.count))
		}
		m.count += 1
		m.countlock.Unlock()
	}
	m.statuslock.Unlock()

	m.mu.Lock()

	m.statuslock.Lock()
	m.locked = true
	m.statuslock.Unlock()
}

// Queue returns the number of goroutines currently waiting to obtain a lock on this mutex
func (m *SmartMutex) Queue() int {
	return m.count
}

// Unlock implements standard Mutex Unlocking, with lock wait queue tracking support
func (m *SmartMutex) Unlock() {
	if m == nil {
		panic("tried to unlock a nil mutex")
	}
	if mutexTracing == true {
		if m.count > 1 {
			m.trace(fmt.Sprintf("releasing mutex lock on %s (%d still in queue)", m.name, m.count))
		}
	}

	// unlike most inline code mutexes, the amount of sources of concurrency here, and the likely
	// relative unimportance of unlocking a prior lock as an indicator of code integrity
	// the normal panic here is recovered and logged, but not made fatal
	defer func() {
		if r := recover(); r != nil {
			m.trace(fmt.Sprintf("Recovered after unlocking unlocked mutex %s in f %v", m.name, r))
		}
	}()

	if m.count > 0 {
		m.countlock.Lock()
		m.count--
		m.countlock.Unlock()
	}

	m.statuslock.Lock()
	if m.locked == true {
		m.mu.Unlock()
		m.locked = false
	}
	m.statuslock.Unlock()
}

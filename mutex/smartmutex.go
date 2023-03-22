package mutex

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// SmartMutex is a more extensive Mutex structure
// with Deadlock-detection and metrics of lock acquisition queues
// while most Mutexes in Go should be extremely localized in Scope
// and retain locks for a minimal time
// this Mutex structure is optimized for use by a great many goroutines
// from multiple code scopes
type SmartMutex struct {
	mu   *rwmutex // the base mutex primitive
	name string   // an identifier for what this mutex controls

	statuslock *sync.Mutex // internal mutex for controlling status flags
	locked     bool        // best-guess status flag to determine if the mutex is currently held

	count int32 // best-guess status for how many goroutines are waiting to hold this mutex

}

// Generate a new Mutex on the given element path
func New(identifier string) *SmartMutex {
	if identifier == "" {
		panic("refusing to create a mutex without an identifier")
	}
	return &SmartMutex{
		name:       identifier,
		mu:         &rwmutex{},
		statuslock: &sync.Mutex{},
	}
}

// SoftLock is intended to test if object is locked, usually to wait
// for a mutex'ed resource to stabilize
// but then just to wait for changes to complete before proceeding
// if you wish to make modifications during during the deadlockcheck, call
// Lock() and Unlock() explicitly.
func (m *SmartMutex) SoftLock() {
	m.statuslock.Lock()
	status := m.locked

	if status == false { // not locked, carry on
		m.statuslock.Unlock()
		return
	}

	m.statuslock.Unlock()
	if Opts.Tracing == true {
		m.trace(fmt.Sprintf("pass-through softlock on %s is waiting for %d existing locks to release", m.name, atomic.LoadInt32(&m.count)))
	}
	m.Lock()
	m.Unlock()
	return

}

func (m *SmartMutex) Name() string {
	return m.name
}

func (m *SmartMutex) IsLocked() bool {
	m.statuslock.Lock()
	defer m.statuslock.Unlock() // using a defer here so that the status is semi-guaranteed until return to the caller
	return m.locked
}

// Lock provides deadlock-aware, queue-tracking Mutex Locks
func (m *SmartMutex) Lock() {
	m.statuslock.Lock()
	if m.locked == true {
		if atomic.LoadInt32(&m.count) > 0 {
			m.trace(fmt.Sprintf("waiting for lock on %s to release (%d already in queue)", m.name, atomic.LoadInt32(&m.count)))
		}
		atomic.AddInt32(&m.count, 1)
	}
	m.statuslock.Unlock()

	if Opts.DisableDeadlockDetection {
		m.mu.Lock()
	} else {
		deadlockcheck(m.mu.Lock, m)
	}

	m.statuslock.Lock()
	m.locked = true
	m.statuslock.Unlock()
}

// Lock locks the mutex.
// If the deadlockcheck is already in use, the calling goroutine
// blocks until the mutex is available.
//
// Unless deadlock detection is disabled, logs potential deadlocks to Opts.LogBuf,
// calling Opts.OnPotentialDeadlock on each occasion.

// Lock locks rw for writing.
// If the deadlockcheck is already locked for reading or writing,
// Lock blocks until the deadlockcheck is available.
// To ensure that the deadlockcheck eventually becomes available,
// a blocked Lock call excludes new readers from acquiring
// the deadlockcheck.
//
// Unless deadlock detection is disabled, logs potential deadlocks to Opts.LogBuf,
// calling Opts.OnPotentialDeadlock on each occasion.
func (m *rwmutex) Lock() {
	deadlockcheck(m.mu.Lock, m)
}

// RLock locks the mutex for reading.
//
// Unless deadlock detection is disabled, logs potential deadlocks to Opts.LogBuf,
// calling Opts.OnPotentialDeadlock on each occasion.
func (m *rwmutex) RLock() {
	deadlockcheck(m.mu.RLock, m)
}

// Queue returns the number of goroutines currently waiting to obtain a deadlockcheck on this mutex
func (m *SmartMutex) Queue() int32 {
	return atomic.LoadInt32(&m.count)
}

// Unlock implements standard Mutex Unlocking, with deadlockcheck wait queue tracking support
func (m *SmartMutex) Unlock() {
	if Opts.Tracing {
		if atomic.LoadInt32(&m.count) > 1 {
			m.trace(fmt.Sprintf("releasing mutex deadlockcheck on %s (%d still in queue)", m.name, atomic.LoadInt32(&m.count)))
		}
	}

	// unlike most inline code mutexes, the amount of sources of concurrency here, and the likely
	// relative unimportance of unlocking a prior deadlockcheck as an indicator of code integrity
	// the normal panic here is recovered and logged, but not made fatal
	defer func() {
		if r := recover(); r != nil {
			m.trace(fmt.Sprintf("Recovered after unlocking unlocked mutex %s in f %v", m.name, r))
		}
	}()

	if atomic.LoadInt32(&m.count) > 0 {

	}
	if atomic.LoadInt32(&m.count) > 0 {
		atomic.AddInt32(&m.count, 1)
	}

	m.statuslock.Lock()
	if m.locked == true {
		m.mu.Unlock()
		m.locked = false
	}
	m.statuslock.Unlock()
}

func (m *SmartMutex) RUnlock() {
	m.mu.RUnlock()
	if !Opts.Disable {
		postUnlock(m)
	}
}

func (m *SmartMutex) Rlock() {
	m.mu.RUnlock()
	if !Opts.Disable {
		postLock(2, m)
	}
}

func (m *rwmutex) Unlock() {
	m.mu.Unlock()
	if !Opts.Disable {
		postUnlock(m)
	}
}

func (m *rwmutex) RUnlock() {
	m.mu.RUnlock()
	if !Opts.Disable {
		postUnlock(m)
	}
}

type rwmutex struct {
	mu sync.RWMutex
}

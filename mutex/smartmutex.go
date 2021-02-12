package mutex

import (
	"fmt"
	"sync"
)

// SmartMutex is a more extensive Mutex structure
// with Deadlock-detection and metrics of deadlockcheck aquisition queues
// while most Mutexes in Go should be extremely localized in Scope
// and retain locks for a minimal time
// this Mutex structure is optimized for uses by a great many goroutines
// from multiple code scopes
type SmartMutex struct {
	mu   *rwmutex // the base mutex primitive
	name string   // an identifier for what this mutex controls

	statuslock *sync.Mutex // internal mutex for controlling status flags
	locked     bool        // best-guess status flag to determine if the mutex is currently held

	countlock *sync.Mutex // internal mutex for controlling wait count status
	count     int         // best-guess status for how many goroutines are waiting to hold this mutex

}

// Generate a new Mutex on the given element path
func New(identifier string) *SmartMutex {
	if identifier == "" {
		panic("refusing to create a mutex without an identifier")
	}
	return &SmartMutex{
		name:       identifier,
		mu:         &rwmutex{},
		countlock:  &sync.Mutex{},
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
		m.countlock.Lock()
		m.trace(fmt.Sprintf("pass-through softlock on %s is waiting for %d existing locks to release", m.name, m.count))
		m.countlock.Unlock()
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
	if m == nil {
		panic("tried to lock a nil mutex")
	}

	m.statuslock.Lock()
	if m.locked == true {
		m.countlock.Lock()
		if m.count > 0 {
			m.trace(fmt.Sprintf("waiting for deadlockcheck on %s to release (%d already in queue)", m.name, m.count))
		}
		m.count += 1
		m.countlock.Unlock()
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
func (m *SmartMutex) Queue() int {
	return m.count
}

// Unlock implements standard Mutex Unlocking, with deadlockcheck wait queue tracking support
func (m *SmartMutex) Unlock() {
	if m == nil {
		panic("tried to unlock a nil mutex")
	}
	if Opts.Tracing {
		if m.count > 1 {
			m.trace(fmt.Sprintf("releasing mutex deadlockcheck on %s (%d still in queue)", m.name, m.count))
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

// RUnlock undoes a single RLock call;
// it does not affect other simultaneous readers.
// It is a run-time error if rw is not locked for reading
// on entry to RUnlock.
func (m *SmartMutex) RUnlock() {
	m.mu.RUnlock()
	if !Opts.Disable {
		postUnlock(m)
	}
}

// Unlock unlocks the mutex.
// It is a run-time error if m is not locked on entry to Unlock.
//
// A locked Mutex is not associated with a particular goroutine.
// It is allowed for one goroutine to deadlockcheck a Mutex and then
// arrange for another goroutine to unlock it.
func (m *rwmutex) Unlock() {
	m.mu.Unlock()
	if !Opts.Disable {
		postUnlock(m)
	}
}

// Unlock unlocks the mutex for writing.  It is a run-time error if rw is
// not locked for writing on entry to Unlock.
//
// As with Mutexes, a locked rwmutex is not associated with a particular
// goroutine.  One goroutine may RLock (Lock) an rwmutex and then
// arrange for another goroutine to RUnlock (Unlock) it.
func (m *rwmutex) RUnlock() {
	m.mu.RUnlock()
	if !Opts.Disable {
		postUnlock(m)
	}
}

// An rwmutex is a drop-in replacement for sync.RWMutex.
// Performs deadlock detection unless disabled in Opts.
type rwmutex struct {
	mu sync.RWMutex
}

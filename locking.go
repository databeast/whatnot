package whatnot

import (
	"context"
	"sync"
	"time"

	"github.com/databeast/whatnot/access"
)

// reslock operations that take longer that this are considered failed
const defaultLockAttemptTimeout = time.Second * 1

// resourceLock is a Temporary Locking Semaphore on an namespace element resource
type resourceLock struct {
	logsupport
	selfmu    *sync.Mutex // mutex for modifying myself
	resmu     *sync.Mutex // mutex for modifying my attached Path Element
	islocked  bool		  // readable state flag to making mutex state knowable
	recursive bool		  // does this resource lock cover child Path Elements?
	Role      access.Role // APi Role that is keeping this locked
	deadline  context.Context
}

// unlockAfterExpire sets the given Path Element to remove any leases on it after the given duration
func (m *PathElement) unlockAfterExpire() {
	go func() {
		select {
			case <-m.reslock.deadline.Done():
				if m.reslock.recursive {
					unlockWg := &sync.WaitGroup{}
					unlockWg.Add(1)
					go m.asyncRecursiveUnLockSelfAndSubs(unlockWg)
					unlockWg.Wait()
				} else {
					m.reslock.resmu.Unlock()
				}
		}
	}()
}

func (r *resourceLock) lock(recursive bool) {
	r.selfmu.Lock()
	if r.islocked {
		r.Debug("waiting to claim additional lock")
	}

	r.resmu.Lock()
	r.recursive = recursive
	r.islocked = true

	r.selfmu.Unlock()
}

func (r *resourceLock) unlock() {
	r.selfmu.Lock()
	if r.islocked == false {
		r.Warn("ignoring call to unlock already unlocked reslock")
		r.resmu.Unlock()
		return
	}
	r.resmu.Unlock()
	r.islocked = false
	r.selfmu.Unlock()
}

// Lock places a Mutex on this pathElement
// and sends a notification of this lock to its chain of parent elements
func (m *PathElement) Lock() {
	m.reslock.lock(false)
	m.parentnotify <- elementChange{elem: m, change: ChangeLocked}
	m.selfevents <- elementChange{elem: m, change: ChangeLocked}
}

// UnLock will release the Mutex Lock on this path element
// Note that it will NOT unlock mutexes on sub-element
// unlocking will sent a notification event to the chain of parent elements
func (m *PathElement) UnLock() {
	//NOTE: Subs will Remain Locked when doing this.
	m.reslock.unlock()
	m.parentnotify <- elementChange{elem: m, change: ChangeUnlocked}
	m.selfevents <- elementChange{elem: m, change: ChangeUnlocked}
}

// LockSubs will lock this Path Element and every Path Element it is a parent to
func (m *PathElement) LockSubs() {
	// TODO:  Implement deadlock timeouts and recovery
	lockwg := &sync.WaitGroup{}
	lockwg.Add(1)
	m.asyncRecursiveLockSelfAndSubs(lockwg)
	lockwg.Wait()
	m.parentnotify <- elementChange{elem: m, change: ChangeLocked}
	m.selfevents  <- elementChange{elem: m, change: ChangeLocked}
}

func (m *PathElement) UnLockSubs() {
	// TODO:  Implement deadlock timeouts and recovery
	unlockwg := &sync.WaitGroup{}
	unlockwg.Add(1)
	m.asyncRecursiveUnLockSelfAndSubs(unlockwg)
	unlockwg.Wait()
	m.parentnotify <- elementChange{elem: m, change: ChangeUnlocked}
	m.selfevents  <- elementChange{elem: m, change: ChangeUnlocked}
}

func (m *PathElement) asyncRecursiveLockSelfAndSubs(parentwg *sync.WaitGroup) {
	m.reslock.lock(true) // reslock myself first

	if len(m.children) > 0 {
		subLockWg := &sync.WaitGroup{}
		subLockWg.Add(len(m.children)) // always increment the waitgroup delta before allowing anything to start

		for _, v := range m.children {
			go v.asyncRecursiveLockSelfAndSubs(subLockWg)
		}
		subLockWg.Wait()
	}
	parentwg.Done()
}

func (m *PathElement) asyncRecursiveUnLockSelfAndSubs(parentwg *sync.WaitGroup) {

	m.reslock.unlock() // unlock myself first

	if len(m.children) > 0 {
		subUnlockwg := &sync.WaitGroup{}
		subUnlockwg.Add(len(m.children)) // always increment the waitgroup delta before allowing anything to start

		for _, v := range m.children {
			go v.asyncRecursiveUnLockSelfAndSubs(subUnlockwg)
		}
	}
	parentwg.Done()
}

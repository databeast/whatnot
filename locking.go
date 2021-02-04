package whatnot

import (
	"context"
	"sync"
	"time"

	"github.com/databeast/whatnot/access"
)

const LOCKING_FAIL_TIMEOUT = time.Second * 1 // reslock operations that take longer that this are considered failed

// resourceLock is a Temporary Locking Semaphore on an namespace element resource
type resourceLock struct {
	selfmu    *sync.Mutex // mutex for modifying myself
	resmu     *sync.Mutex // mutex for modifying my attached Path Element
	islocked  bool
	recursive bool
	Role      access.Role // APi Role that is keeping this locked
	deadline  context.Context
}

// unlockAfterExpire sets the given Path Element to remove any leases on it after the given duration
func (r *PathElement) unlockAfterExpire(ttl time.Duration) {
	// assign a deadline context to this reslock
	r.reslock.selfmu.Lock()
	dl, cancelFunc := context.WithTimeout(context.Background(), ttl)
	r.reslock.deadline = dl
	r.reslock.selfmu.Unlock()

	go func() {
		select {
		case <-r.reslock.deadline.Done():
			if r.reslock.recursive {
				unlockWg := &sync.WaitGroup{}
				unlockWg.Add(1)
				go r.asyncRecursiveUnLockSelfAndSubs(unlockWg)
				unlockWg.Wait()
				cancelFunc()
			} else {
				r.reslock.resmu.Unlock()
				defer cancelFunc()
			}
		}

	}()

}

func (r *resourceLock) lock(recursive bool) {
	r.selfmu.Lock()
	if r.islocked {
		namespaceLogging.Debug("waiting to claim additional lock")
	}
	r.resmu.Lock()
	r.recursive = recursive
	r.islocked = true
	r.selfmu.Unlock()
}

func (r *resourceLock) unlock() {
	r.selfmu.Lock()
	if r.islocked == false {
		namespaceLogging.Info("ignoring call to unlock already unlocked reslock")
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
	go func() { m.parentnotify <- elementChange{elem: m, change: LOCKED} }()
}

// UnLock will release the Mutex Lock on this path element
// Note that it will NOT unlock mutexes on sub-element
// unlocking will sent a notification event to the chain of parent elements
func (m *PathElement) UnLock() {
	//NOTE: Subs will Remain Locked when doing this.
	m.reslock.unlock()
	go func() { m.parentnotify <- elementChange{elem: m, change: UNLOCKED} }()
}

// LockSubs will lock this Path Element and every Path Element it is a parent to
func (m *PathElement) LockSubs() {
	// TODO:  Implement deadlock timeouts and recovery
	lockwg := &sync.WaitGroup{}
	lockwg.Add(1)
	m.asyncRecursiveLockSelfAndSubs(lockwg)
	lockwg.Wait()
	go func() { m.parentnotify <- elementChange{elem: m, change: LOCKED} }()
}

func (m *PathElement) UnLockSubs() {
	// TODO:  Implement deadlock timeouts and recovery
	unlockwg := &sync.WaitGroup{}
	unlockwg.Add(1)
	m.asyncRecursiveUnLockSelfAndSubs(unlockwg)
	unlockwg.Wait()
	go func() { m.parentnotify <- elementChange{elem: m, change: UNLOCKED} }()
}

func (m *PathElement) asyncRecursiveLockSelfAndSubs(parentwg *sync.WaitGroup) {
	m.reslock.lock(true) // reslock myself first

	if len(m.subs) > 0 {
		subLockWg := &sync.WaitGroup{}
		subLockWg.Add(len(m.subs)) // always increment the waitgroup delta before allowing anything to start

		for _, v := range m.subs {
			go v.asyncRecursiveLockSelfAndSubs(subLockWg)
		}
		subLockWg.Wait()
	}
	parentwg.Done()
}

func (m *PathElement) asyncRecursiveUnLockSelfAndSubs(parentwg *sync.WaitGroup) {

	m.reslock.unlock() // unlock myself first

	if len(m.subs) > 0 {
		subUnlockwg := &sync.WaitGroup{}
		subUnlockwg.Add(len(m.subs)) // always increment the waitgroup delta before allowing anything to start

		for _, v := range m.subs {
			go v.asyncRecursiveUnLockSelfAndSubs(subUnlockwg)
		}
	}
	parentwg.Done()
}

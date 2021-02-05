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
func (m *PathElement) unlockAfterExpire(ttl time.Duration) {
	// assign a deadline context to this reslock
	m.reslock.selfmu.Lock()
	dl, cancelFunc := context.WithTimeout(context.Background(), ttl)
	m.reslock.deadline = dl
	m.reslock.selfmu.Unlock()

	go func() {
		select {
		case <-m.reslock.deadline.Done():
			if m.reslock.recursive {
				unlockWg := &sync.WaitGroup{}
				unlockWg.Add(1)
				go m.asyncRecursiveUnLockSelfAndSubs(unlockWg)
				unlockWg.Wait()
				cancelFunc()
			} else {
				m.reslock.resmu.Unlock()
				//defer cancelFunc()
				cancelFunc()
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
	select {
		case m.parentnotify <- elementChange{elem: m, change: LOCKED}:
			// notification was sent
		default:
			// nobody was listening
	}
}

// UnLock will release the Mutex Lock on this path element
// Note that it will NOT unlock mutexes on sub-element
// unlocking will sent a notification event to the chain of parent elements
func (m *PathElement) UnLock() {
	//NOTE: Subs will Remain Locked when doing this.
	m.reslock.unlock()
	go func() {
		select {
			case m.parentnotify <- elementChange{elem: m, change: UNLOCKED}:
				// nofitication has been sent
			default:
				// nobody was listening
		}
	}()
}

// LockSubs will lock this Path Element and every Path Element it is a parent to
func (m *PathElement) LockSubs() {
	// TODO:  Implement deadlock timeouts and recovery
	lockwg := &sync.WaitGroup{}
	lockwg.Add(1)
	m.asyncRecursiveLockSelfAndSubs(lockwg)
	lockwg.Wait()
	go func() {
		select {
			case m.parentnotify <- elementChange{elem: m, change: LOCKED}:
				// notification has been sent
			default:
				// nobody is listening
		}
	}()
}

func (m *PathElement) UnLockSubs() {
	// TODO:  Implement deadlock timeouts and recovery
	unlockwg := &sync.WaitGroup{}
	unlockwg.Add(1)
	m.asyncRecursiveUnLockSelfAndSubs(unlockwg)
	unlockwg.Wait()
	go func() {
		select {
			case m.parentnotify <- elementChange{elem: m, change: UNLOCKED}:
				// notification has been sent
			default:
				// nobody is listening
		}
	}()
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

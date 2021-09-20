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
	islocked  bool        // readable state flag to making mutex state knowable
	recursive bool        // does this resource lock cover child Path Elements?
	Role      access.Role // APi Role that is keeping this locked
	deadline  context.Context
}

// unlockAfterExpire sets the given Path Element to remove any leases on it after the given duration
func (p *PathElement) unlockAfterExpire() {
	go func() {
		select {
		case <-p.reslock.deadline.Done():
			if p.reslock.recursive {
				unlockWg := &sync.WaitGroup{}
				unlockWg.Add(1)
				go p.asyncRecursiveUnLockSelfAndSubs(unlockWg)
				unlockWg.Wait()
			} else {
				p.reslock.resmu.Unlock()
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
// this also fulfills the interface Sync.Locker
func (p *PathElement) Lock() {
	p.reslock.lock(false)
	p.selfnotify <- elementChange{id: randid.Uint64(), elem: p, change: ChangeLocked}
}

// UnLock will release the Mutex Lock on this path element
// Note that it will NOT unlock mutexes on sub-element
// unlocking will sent a notification event to the chain of parent elements
// this also fulfills the interface Sync.Locker
func (p *PathElement) UnLock() {
	//NOTE: Subs will Remain Locked when doing this.
	p.reslock.unlock()
	p.selfnotify <- elementChange{id: randid.Uint64(), elem: p, change: ChangeUnlocked}
}

// LockSubs will lock this Path Element and every Path Element it is a parent to
func (p *PathElement) LockSubs() {
	// TODO:  Implement deadlock timeouts and recovery
	lockwg := &sync.WaitGroup{}
	lockwg.Add(1)
	p.asyncRecursiveLockSelfAndSubs(lockwg)
	lockwg.Wait()
	p.selfnotify <- elementChange{id: randid.Uint64(), elem: p, change: ChangeLocked}
}

func (p *PathElement) UnLockSubs() {
	// TODO:  Implement deadlock timeouts and recovery
	unlockwg := &sync.WaitGroup{}
	unlockwg.Add(1)
	p.asyncRecursiveUnLockSelfAndSubs(unlockwg)
	unlockwg.Wait()
	p.selfnotify <- elementChange{elem: p, change: ChangeUnlocked}
}

func (p *PathElement) asyncRecursiveLockSelfAndSubs(parentwg *sync.WaitGroup) {
	p.reslock.lock(true) // reslock myself first

	if len(p.children) > 0 {
		subLockWg := &sync.WaitGroup{}
		subLockWg.Add(len(p.children)) // always increment the waitgroup delta before allowing anything to start

		for _, v := range p.children {
			go v.asyncRecursiveLockSelfAndSubs(subLockWg)
		}
		subLockWg.Wait()
	}
	parentwg.Done()
}

func (p *PathElement) asyncRecursiveUnLockSelfAndSubs(parentwg *sync.WaitGroup) {
	p.reslock.unlock() // unlock myself first

	subUnlockwg := &sync.WaitGroup{}
 
	for _, v := range p.children {
		subUnlockwg.Add(1) // always increment the waitgroup delta before allowing anything to start
		go v.asyncRecursiveUnLockSelfAndSubs(subUnlockwg)
	}
	subUnlockwg.Wait()
	p.Debugf("%s has finished unlocking all its children", p.section)

	parentwg.Done()
}

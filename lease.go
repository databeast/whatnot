package whatnot

/*
Leases allow time-limited Mutex control over an element in a namespace
and optionally, all elements beneath it
*/

import (
	"context"
	"time"
)

// LeaseContext implements Element Locking Lease control as a Context Interface object
// it is heavily recommend to use this as the context object for the rest of your functions
// lifetime to keep it in sync with the accordant lease it was generated with to enable
// your code to control and react to lease expiration
type LeaseContext struct {
	logsupport
	ctx       context.Context
	elem      *PathElement
	recursive bool
	cancel    func()
}

// Deadline implements the Context interface
func (l *LeaseContext) Deadline() (time.Time, bool) {
	return l.ctx.Deadline()
}

// Done implements the Context interface
func (l *LeaseContext) Done() <-chan struct{} {
	return l.ctx.Done()
}

// Err implements the Context interface
func (l *LeaseContext) Err() error {
	return l.ctx.Err()
}

// Value implements the Context interface
func (l *LeaseContext) Value(key interface{}) interface{} {
	return l.ctx.Value(key)
}

// Cancel implements the Context interface
func (l *LeaseContext) Cancel() {
	if l.recursive == true {
		go l.elem.UnLockSubs()
	} else {
		go l.elem.UnLock()
	}
}

// LockWithLease will lock a single path element with a timed lease on the lock
// it uses a a background context so cannot be cancelled before the lease expires
func (p *PathElement) LockWithLease(ttl time.Duration) (ctx *LeaseContext, release func()) {
	return p.generateLease(context.Background(), ttl, false)
}

// ContextLockWithLease will lock a single path element with a timed lease on the lock
// you provide the context instance to have external control to cancel it before timeout
func (p *PathElement) ContextLockWithLease(octx context.Context, ttl time.Duration) (ctx *LeaseContext, release func()) {
	return p.generateLease(octx, ttl, false)
}

// LockPrefixWithLease will lock a path element and all sub-elements with a timed lease on the lock
// it uses a a background context so cannot be cancelled before the lease expires
func (p *PathElement) LockPrefixWithLease(ttl time.Duration) (ctx *LeaseContext, release func()) {
	return p.generateLease(context.Background(), ttl, true)
}

// ContextLockPrefixWithLease will lock a path element and all sub-elements with a timed lease on the lock
// you provide the context instance to have external control to cancel it before timeout
func (p *PathElement) ContextLockPrefixWithLease(octx context.Context, ttl time.Duration) (ctx *LeaseContext, release func()) {
	return p.generateLease(octx, ttl, true)
}

func (p *PathElement) generateLease(octx context.Context, ttl time.Duration, recursive bool) (ctx *LeaseContext, release func()) {

	dl, cancel := context.WithTimeout(octx, ttl)

	// lock the resource lock structure itself while changing it
	p.reslock.selfmu.Lock()
	p.reslock.deadline = dl
	p.reslock.selfmu.Unlock()

	ctx = &LeaseContext{
		ctx:       dl,
		elem:      p,
		recursive: recursive,
		cancel:    cancel,
	}

	if recursive {
		p.LockSubs()
	} else {
		p.Lock()

	}

	p.unlockAfterExpire()

	return ctx, cancel
}

package mutex

//support for tracking which goroutines are currently waiting to acquire the mutex

import (
	"fmt"
	"github.com/petermattis/goid"
	"sync"
)

type lockOrder struct {
	mu    sync.Mutex
	cur   map[interface{}]stackGID // stacktraces + gids for the locks currently taken.
	order map[beforeAfter]ss       // expected order of locks.
}

// Under lo.mu Locked.
func (l *lockOrder) other(ptr interface{}) {
	empty := true
	for k := range l.cur {
		if k == ptr {
			continue
		}
		empty = false
	}
	if empty {
		return
	}
	fmt.Fprintln(Opts.LogBuf, "Other goroutines holding locks:")
	for k, pp := range l.cur {
		if k == ptr {
			continue
		}
		fmt.Fprintf(Opts.LogBuf, "goroutine %v deadlockcheck %p\n", pp.gid, k)
		dumpDeadlock(Opts.LogBuf, pp.stack)
	}
	fmt.Fprintln(Opts.LogBuf)
}

var lo = newLockOrder()

func newLockOrder() *lockOrder {
	return &lockOrder{
		cur:   map[interface{}]stackGID{},
		order: map[beforeAfter]ss{},
	}
}

func (l *lockOrder) postUnlock(p interface{}) {
	l.mu.Lock()
	delete(l.cur, p)
	l.mu.Unlock()
}

func preLock(skip int, p interface{}) {
	lo.preLock(skip, p)
}

func postLock(skip int, p interface{}) {
	lo.postLock(skip, p)
}

func postUnlock(p interface{}) {
	lo.postUnlock(p)
}

func (l *lockOrder) postLock(skip int, p interface{}) {
	stack := callers(skip)
	gid := goid.Get()
	l.mu.Lock()
	l.cur[p] = stackGID{stack, gid}
	l.mu.Unlock()
}

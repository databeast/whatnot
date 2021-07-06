package whatnot

import (
	"errors"
	"fmt"
	"github.com/databeast/whatnot/access"
	"sync"
	"time"
)

/*
Semaphore pool support for PathElements, controlling uses of a particular element and its children

setting a Semaphore pool on a Prefix can set total pool availability for all sub-elements.

Semaphores allow for finer control than the Mutex-style write locking, and allow more control over concurrent access
limits
*/

// SemaphorePool is a combined semaphore for use by a PathElement and all its sub Elements
type SemaphorePool struct {
	onElement *PathElement
	mu        *sync.RWMutex
	maxslots  float64
	usedslots float64
	waiting   *EventMultiplexer
}

type SemaphoreClaim struct {
	fromPool *SemaphorePool
	slots    float64
	returned bool // returning amn already returned claim is bad
}

func (p *SemaphorePool) returnclaim(claim *SemaphoreClaim) (err error) {
	if claim.returned {
		return errors.New("this claim has already been returned")
	}
	defer func() {
		// signal out that a claim has been returned, to any waiting claims
	}()

	p.mu.Lock()

	if p.usedslots+claim.slots > p.maxslots {
		p.usedslots = p.maxslots
		p.mu.Unlock()
		return nil
	}

	if p.usedslots-claim.slots < 0 {
		p.usedslots = 0
		p.mu.Unlock()
		return nil
	}

	p.usedslots -= claim.slots
	id := randid.Uint64()
	p.onElement.parentnotify <- elementChange{
		id:     id,
		elem:   p.onElement,
		change: ChangeReleased,
		actor:  access.Role{},
	}

	p.waiting.Broadcast <- WatchEvent{
		id:     randid.Uint64(),
		elem:   nil,
		TS:     time.Now(),
		Change: ChangeReleased,
		Actor:  access.Role{},
		Note:   "",
	}
	p.mu.Unlock()
	return err
}

// ClaimSingle claims a single unweighted semaphore unit
func (p *SemaphorePool) ClaimSingle(timeout time.Duration) (claim *SemaphoreClaim, err error) {
	p.mu.RLock()
	if p.usedslots + 1 <= p.maxslots { // free slot, lets go
		p.mu.RUnlock()

		claim = &SemaphoreClaim{
			fromPool: p,
			slots:    1,
			returned: false,
		}

		p.mu.Lock()
		p.usedslots ++
		p.mu.Unlock()
		return claim, nil
	}
	// and now we play the waiting game..
	return p.waitForSlot(1, timeout)

}

// loop that waits for a signal from another Claim being released, to check if there's enough space left in the pool
func (p *SemaphorePool) waitForSlot(slots float64, timeout time.Duration) (claim *SemaphoreClaim, err error) {
	notify := make(chan WatchEvent)
	p.waiting.Register(notify, true)
	tick := time.NewTimer(timeout)
	for {
		select {
		case <-notify:
			// ok, we've got a release notification, will this give us enough space?
			p.mu.RLock()
			if p.usedslots + slots <= p.maxslots {
				p.mu.RUnlock()
				p.mu.Lock()
				p.usedslots += slots
				p.mu.Unlock()
				claim = &SemaphoreClaim{
					fromPool: p,
					slots:    slots,
					returned: false,
				}
				p.waiting.Unregister(notify)
				return claim, nil
			}
			p.mu.RUnlock()
			// otherwise, just loop around and wait for the next signal
		case <-tick.C:
			// timeout has occurred
			p.waiting.Unregister(notify)
			return nil, errors.New("timeout passed waiting for available semaphore slots")
		}
	}
}



// ClaimWeighted claims a numerically weighted semaphore unit,
func (p *SemaphorePool) ClaimWeighted() (claim *SemaphoreClaim, err error) {
	return claim, err
}

// ClaimPercentageWeighted claims a semaphore unit, weighted as a percentage of the total semaphore pool
func (p *SemaphorePool) ClaimPercentageWeighted(poolpercentage float64, timeout time.Duration) (claim *SemaphoreClaim, err error) {
	p.mu.RLock()
	claimslots := p.maxslots / poolpercentage
	if p.usedslots + claimslots <= p.maxslots {
		p.mu.RUnlock()
		p.mu.Lock()
		// hurrah, we have the pool space available, lets grant it to them and let them get outta here
		p.usedslots += claimslots
		claim = &SemaphoreClaim{
			fromPool: p,
			slots:    claimslots,
			returned: false,
		}
		p.mu.Unlock()
		return claim, nil
	}
	// got to wait in line for what we want
	return p.waitForSlot(claimslots, timeout)
}

// Return releases the semaphore claim back to the pool
func (p *SemaphoreClaim) Return() error {
	return p.fromPool.returnclaim(p)
}

type SemaphorePoolOpts struct {
	PoolSize float64 // Total Pool Weight available to divide amongst claims in this pool
}

// CreateSemaphorePool instantiates a semaphore pool on this path element.
// prefix will attach the pool to all child elements
// purge will remove any existing semaphore pool, including from all children if prefix is true
func (p *PathElement) CreateSemaphorePool(prefix bool, purge bool, opts SemaphorePoolOpts) (err error) {
	if !purge && p.semaphores != nil {
		return errors.New(fmt.Sprintf("semaphore pool already exists for %s", p.AbsolutePath()))
	}
	p.semaphores = &SemaphorePool{
		onElement: p,
		mu:        &sync.RWMutex{},
		maxslots:  opts.PoolSize,
		usedslots: 0,
		waiting:   NewEventsMultiplexer(),
	}


	if prefix {
		for _, c := range p.children {
			if c.semaphores == nil || purge {
				c.semaphores = p.semaphores
			}
		}
	}
	return nil
}

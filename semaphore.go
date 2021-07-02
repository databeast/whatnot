package whatnot

import (
	"errors"
	"fmt"
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
	mu        *sync.RWMutex
	maxslots  float64
	usedslots float64
	waiting   EventMultiplexer
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

	p.mu.Unlock()
	return err
}

// ClaimSingle claims a single unweighted semaphore unit
func (p *SemaphorePool) ClaimSingle() (claim *SemaphoreClaim, err error) {
	return claim, err
}

// ClaimWeighted claims a numerically weighted semaphore unit,
func (p *SemaphorePool) ClaimWeighted() (claim *SemaphoreClaim, err error) {
	return claim, err
}

// ClaimPercentageWeighted claims a semaphore unit, weighted as a percentage of the total semaphore pool
func (p *SemaphorePool) ClaimPercentageWeighted(poolpercentage float64, timeout time.Duration) (claim *SemaphoreClaim, err error) {
	p.mu.Lock()
	claimslots := p.maxslots / poolpercentage
	if p.usedslots + claimslots < p.maxslots {
		// hurrah, we have the pool space available, lets grant it to them and let them get outta here
		p.usedslots += claimslots
	}
	p.mu.Lock()
	claim = &SemaphoreClaim{}


	return claim, err
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
		mu:        &sync.RWMutex{},
		maxslots:  opts.PoolSize,
		usedslots: 0,
		waiting:   EventMultiplexer{},
	}
	if prefix {
		for _, c := range p.children {
			err = c.CreateSemaphorePool(true, purge, opts)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

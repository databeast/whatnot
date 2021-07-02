package whatnot

import (
	"errors"
	"fmt"
	"sync"
)

/*
Semaphore pool support for PathElements, controlling uses of a particular element and its children

setting a Semaphore pool on a Prefix can set total pool availability for all sub-elements.

Semaphores allow for finer control than the Mutex-style write locking, and allow more control over concurrent access
limits
*/

// SemaphorePool is a combined semaphore for use by a PathElement and all its sub Elements
type SemaphorePool struct {
	mu *sync.RWMutex
	maxslots float64
	usedslots float64
	slots 	float64
}

type SemaphoreClaim struct {
	fromPool *SemaphorePool
	slots	 float64
	returned bool // returning amn already returned claim is bad
}

func (p *SemaphorePool) returnclaim(claim *SemaphoreClaim) (err error) {
	if claim.returned {
		return errors.New("this claim has already been returned")
	}
	p.mu.Lock()

	// has something terribly wrong happened while we weren't looking?
	if p.slots + claim.slots > p.maxslots + 0.00000001 {
		// slightly extending the maxslots because floats are weird.
		p.slots = p.maxslots
		p.mu.Unlock()
		return nil
	}

	if p.slots - claim.slots < -0.000000001 {
		// slightly extending the minslots because floats are weird.
		p.slots = 0
		p.mu.Unlock()
		return nil
	}

	p.slots -= claim.slots

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
func (p *SemaphorePool) ClaimPercentageWeighted(poolpercentage int) (claim *SemaphoreClaim, err error) {
	// 8 significant digits as maximum, to avoid float weirdness overflows
	return claim, err
}

// Return releases the semaphore claim back to the pool
func (p *SemaphoreClaim) Return() {
	p.fromPool.
}

type SemaphorePoolOpts struct {
	PoolWeight int   // Total Pool Weight available to divide amongst claims in this pool

}

// CreateSemaphorePool instantiates a semaphore pool on this path element.
// prefix will attach the pool to all child elements
// purge will remove any existing semaphore pool, including from all children if prefix is true
func (p *PathElement) CreateSemaphorePool(prefix bool, purge bool, opts SemaphorePoolOpts) (err error) {
	if !purge && p.semaphores != nil {
		return errors.New(fmt.Sprintf("semaphore pool already exists for %s", p.AbsolutePath()))
	}
	p.semaphores = &SemaphorePool{}
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

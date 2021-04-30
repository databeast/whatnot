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
}

type SemaphoreClaim struct {
	fromPool *SemaphorePool

}

// ClaimSingle claims a single unweighted semaphore unit
func (p *SemaphorePool) ClaimSingle() {

}

// ClaimWeighted claims a numerically weighted semaphore unit,
func (p *SemaphorePool) ClaimWeighted() {

}

// ClaimPercentageWeighted claims a semaphore unit, weighted as a percentage of the total semaphore pool
func (p *SemaphorePool) ClaimPercentageWeighted() {

}

func (p *SemaphorePool) Return() {

}

type SemaphorePoolOpts struct {
	PoolWeight int   // Total Pool Weight available to divide amongst claims in this pool

}

// CreateSemaphorePool instantiates a semaphore pool on this path element.
// prefix will attach the pool to all child elements
// purge will remove any existing semaphore pool, including from all children if prefix is true
func (p *PathElement) CreateSemaphorePool(prefix bool, purge bool, opts SemaphorePoolOpts) (err error) {
	if !purge && p.semaphores != nil {
		return errors.New(fmt.Sprintf("semaphore pool already exists for %s", p.AbsolutePath())
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

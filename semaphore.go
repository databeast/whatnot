package whatnot

/*
Semaphore pool support for PathElements, controlling uses of a particular element and its children

setting a Semaphore pool on a Prefix can set total pool availability for all sub-elements.

Semaphores allow for finer control than the Mutex-style write locking, and allow more control over concurrent access
limits
*/

// SemaphorePool is a combined semaphore for use by a PathElement and all its sub Elements
type SemaphorePool struct {

}

func (p *SemaphorePool) Claim() {

}

func (p *SemaphorePool) Return() {

}

// CreateSemaphorePool instantiates a semaphore pool on this path element.
// prefix will attach the pool to all child elements
// purge will remove any existing semaphore pool, including from all children if prefix is true
func (p *PathElement) CreateSemaphorePool(prefix bool, purge bool) {

}
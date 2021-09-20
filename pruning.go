package whatnot

import (
	"context"
	"time"
)

/*
Support for pruning off elements from a namespace once they haven't been active in a given amount of time

highly recommended for elements that contain no attached data, as they will be recreated once referenced again
at a small cost in additional latency
*/

// tracking information for LRU pruning of path elements
type pruningTracker struct {
	// how long to wait until pruning this element after its most recent usage
	pruneAfter time.Duration

	// the last time this element itself was accessed
	lastSelfUsed time.Time

	// the last time any of this elements children were accessed
	lastChildUsed time.Time

	// do not prune this element if it, or any of its childre, have a Value set
	retainData bool
}

func (p *PathElement) EnablePruningAfter(age time.Duration) {
	p.prunectx, p.prunefunc = context.WithCancel(context.Background())
	p.prunetracker = &pruningTracker{
		pruneAfter:    age,
		lastSelfUsed:  time.Now(),
		lastChildUsed: time.Now(),
		retainData:    false,
	}
}

func (p *PathElement) PreventPruning() {
	if p.prunetracker != nil {
		p.prunetracker.retainData = true
	}
}

func (p *PathElement) prune() {
	p.prunechildren()

	if p.prunetracker == nil {
		return
	}
	if p.prunetracker.retainData {
		return // this element is not prunable
	}
	// if the children are in use, then this element is not prunable
	if time.Now().Sub(p.prunetracker.lastChildUsed) < p.prunetracker.pruneAfter {
		return
	}
	// if the children are no longer in use, or this element has no children, test if it can be pruned away
	if time.Now().Sub(p.prunetracker.lastSelfUsed) > p.prunetracker.pruneAfter {
		p.Delete() // TODO: what happens when only partial deletes occur?
	}

}

func (p *PathElement) prunechildren() {
	p.mu.Lock()
	prunelist := p.children // take a copy to get around having to mutex this for long times
	p.mu.Unlock()
 	for _, element := range prunelist {
		element.prune()
	}
}

const pruneInterval = time.Second * 60

func (n *Namespace) pruningcheck() {
	startpruning := time.Tick(pruneInterval)
	select {
	case <-startpruning:
		n.root.prunechildren()
	}

	//TODO : WHAT HAPPENS WHEN PRUNING TAKES LONG THAN PRUNING INTERVAL?
}

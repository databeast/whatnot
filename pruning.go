package whatnot

import "time"

/*
Support for pruning off elements from a namespace once they haven't been active in a given amount of time

highly recommended for elements that contain no attached data, as they will be recreated once referenced again
at a small cost in additional latency
*/

// tracking information for LRU pruning of path elements
type pruningTracker struct {
	// the last time this element itself was accessed
	lastSelfUsed	time.Time

	// the last time any of this elements children were accessed
	lastChildUsed	time.Time

	// do not prune this element if it, or any of its childre, have a Value set
	retainData		bool
}

func (p *PathElement) EnablePruningAfter(age time.Duration) {

}

func (p *PathElement) checkForPruning() bool {
	return false
}

func (p *PathElement) prune() {
	if p.prunetracker == nil {
		return
	}j

}

func (p *PathElement) prunechildren() {
	if p.prunetracker == nil {
		return
	}
}





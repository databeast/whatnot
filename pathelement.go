package whatnot

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/databeast/whatnot/mutex"
	"github.com/pkg/errors"
)

// PathElement is an individual section of a complete path
type PathElement struct {
	logsupport // construct in

	// internal mutex for synchronizing modifications to this structure itself
	mu *mutex.SmartMutex

	// the individual name of this Path Element
	section SubPath

	// The parent Path Element
	parent *PathElement

	// sub Path-elements directly beneath this PathElement
	children map[SubPath]*PathElement

	// channel to notify parent element of changes to this element
	// or any of its sub-elements
	parentnotify chan elementChange

	// channel for incoming change notifications from any of the
	// children of this Path Elements
	subevents chan elementChange

	// channel for events directly on this element itself
	selfnotify chan elementChange

	// reslock is the mutex-like structure governing the leasable lock
	// on the resources represented by this Path Element
	reslock resourceLock

	// additional keyval data attached to this pathelement
	resval ElementValue

	// Channel Multiplexer for sending watch events to subscriptions
	// on this Path Element or any of its children
	subscriberNotify *EventMultiplexer

	// pruning support for shutting down unused areas of the namespace after a duration
	prunetracker *pruningTracker
	prunectx     context.Context
	prunefunc    context.CancelFunc

	// semaphore pool support
	semaphores *SemaphorePool
}

// SubPath returns the name of this Path Element
// without the parent section of the path
// eg the 'file' portion of the path
func (p *PathElement) SubPath() (path SubPath) {
	return p.section
}

// Parent returns the parent PathElement of this PathElement
func (p *PathElement) Parent() *PathElement {
	return p.parent
}

// ParentChain returns a slice of this Path Elements
// parent Pathelements, in order of parentage
// i.e, the first item is this elements immediate parent
// the last item is always the top-level (leftmost) pathelement
func (p *PathElement) ParentChain() (parents []*PathElement) {
	var nextParent *PathElement
	if p.parent == nil {
		return
	} else {
		nextParent = p.parent
	}

	for {
		// dont return the root node
		if nextParent.section == rootId {
			break
		}
		parents = append([]*PathElement{nextParent}, parents...) // prepend it into the list, to the items are in path order
		nextParent = nextParent.Parent()
	}
	return parents
}

// Chain returns the full Path of this Element, as a slice of individual PathElements
func (p *PathElement) Chain() (chain []*PathElement) {
	return append(p.ParentChain(), p)
}

// AbsolutePath returns the full Path of this Element, as a single AbsolutePath instance
func (p *PathElement) AbsolutePath() (path AbsolutePath) {
	for _, element := range p.Chain() {
		path = append(path, element.SubPath())
	}
	return path
}

// fetchSubElement fetches named sub element, if it exists
// returns nil if no sub element by that name exists
func (p PathElement) fetchSubElement(path SubPath) *PathElement {
	sub, ok := p.children[path]
	if ok {
		return sub
	} else {
		return nil
	}
}

func (p *PathElement) logChange(e elementChange) {
	if p.prunetracker != nil {
		if e.change != ChangeDeleted {
			if e.elem == p {
				p.prunetracker.lastSelfUsed = time.Now()
			} else {
				p.prunetracker.lastChildUsed = time.Now()
			}
		}
	}
	switch e.change {
	case ChangeAdded:
		// TODO: call hook function
	case ChangeEdited:
		// TODO: call hook function
	case ChangeLocked:
		// TODO: call hook function
	case ChangeUnlocked:
		// TODO: call hook function
	case ChangeDeleted:
		// TODO: call hook function
	case ChangeUnknown:
		// subscriberStats for now - placeholder for later audit logging
	}

}

// FetchClosestSubPathTail finds the last element in a path chain that most closely resembles the requested path
func (p PathElement) FetchClosestSubPathTail(subPath PathString) *PathElement {
	elemChain := p.FetchClosestSubPath(subPath)
	if len(elemChain) > 0 {
		return elemChain[len(elemChain)-1]
	} else {
		return nil
	}
}

// FetchClosestSubPath will attempt to find the final Path Element that has the
// leading subpath string - this is relative to the pathelement itself, and is not an absolute
// path.
func (p *PathElement) FetchClosestSubPath(subPath PathString) (pathchain []*PathElement) {
	elems := splitPath(subPath)
	var finalElement *PathElement
	var currentElement = p
	var nextElement *PathElement
	for _, e := range elems {
		nextElement = currentElement.fetchSubElement(SubPath(e))
		if nextElement == nil { // this is the closest match we can get
			finalElement = currentElement
			break
		} else {
			currentElement = nextElement
		}
	}
	if finalElement == nil { // we matched every element exactly
		finalElement = currentElement
	}

	pathchain = finalElement.Chain()
	return pathchain
}

// Add a Single subpath Element to this Element
func (p *PathElement) Add(path SubPath) (elem *PathElement, err error) {

	err = path.Validate()
	if err != nil {
		return nil, err
	}

	p.mu.Lock()

	if v, ok := p.children[path]; ok {
		// be safely re-entry - element already exists, do not overwrite
		elem = v
		p.mu.Unlock()
		return elem, nil
	}

	elem = &PathElement{
		section:      path,
		parent:       p,
		parentnotify: p.subevents,
		mu:           mutex.New(fmt.Sprintf("internal mutex for %s", path)),
		children:     make(map[SubPath]*PathElement),
		subevents:    make(chan elementChange, 2),
		selfnotify:   make(chan elementChange, 2),
	}
	p.children[path] = elem
	elem.reslock = resourceLock{
		selfmu:    &sync.Mutex{},
		resmu:     &sync.Mutex{},
		recursive: false,
	}
	// begin the broadcaster for watch subscriptions to function
	elem.initEventBroadcast()

	p.mu.Unlock()
	return elem, nil
}

// attach an existing PathElement to a parent PathElement
func (p *PathElement) attach(elem *PathElement) (err error) {

	// check this is a properly formed element
	if elem.children == nil {
		elem.children = make(map[SubPath]*PathElement)
	}
	// propagate our pruning information down to this element as well
	elem.prunetracker = p.prunetracker

	p.mu.Lock()
	p.children[elem.SubPath()] = elem
	p.mu.Unlock()

	return nil
}

func (p *PathElement) Delete() (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// cascade the context-cancel signal that our event-watching goroutine needs to exit, along with that of all child elements
	p.prunefunc()
	deleteEvent := elementChange{id: randid.Uint64(), elem: p, change: ChangeDeleted}
	p.parentnotify <- deleteEvent
	p.selfnotify <- deleteEvent

	// recursively delete all children
	for _, elem := range p.children {
		err = elem.Delete()
		if err != nil {
			return err // TODO: what happens in this half-deleted state?
		}
	}

	return nil
}

// AppendRelativePath constructs an element-relative subpath, append it to an Existing PathElement,
// creating all PathElements along the way
func (p *PathElement) AppendRelativePath(subPath PathString) (*PathElement, error) {
	// subpaths cannot be absolute, so they cannot start with the delimeter
	if strings.HasPrefix(string(subPath), pathDelimeter) {
		return nil, errors.Errorf("cannot use an absolute path as a subpath")
	}

	pathElems := subPath.ToRelativePath()
	var cur = p
	var err error
	for _, e := range pathElems {
		cur, err = p.Add(e)
		if err != nil {
			return nil, err
		}
		if cur == nil {
			return nil, errors.Errorf("path does not exist")
		}
	}
	return cur, nil
}

// remove the leading match portion of an Absolute Path, return only the portion that is SubPath to this Element
func (p *PathElement) subtractPathToSubPaths(path PathString) (newSubPath []SubPath) {
	return
}

func (p *PathElement) FetchSubPath(subPath PathString) (*PathElement, error) {

	// subpaths cannot be absolute, so they cannot start with the delimeter
	if strings.HasPrefix(string(subPath), pathDelimeter) {
		return nil, errors.Errorf("cannot use an absolute path as a subpath")
	}

	pathElems := subPath.ToRelativePath()
	var cur = p
	for _, e := range pathElems {
		cur = cur.fetchSubElement(e)
		if cur == nil {
			return nil, errors.Errorf("path does not exist")
		}
	}

	return cur, nil
}

// FetchAllSubPaths returns the SubPath location of all descendent PathElements
// underneath this PathElement
func (p *PathElement) FetchAllSubPaths() (allpaths [][]SubPath, err error) {
	for _, s := range p.children {
		elempaths := [][]SubPath{} // all Normalized SubPaths of this Element

		elemsubs, err := s.FetchAllSubPaths()
		if err != nil {
			return nil, err
		}

		if len(elemsubs) == 0 { // this element is a terminal path, just add it directly
			elempaths = append(elempaths, []SubPath{s.SubPath()})
		} else {
			for _, sp := range elemsubs { // each object is a []SubPath
				adjustedpath := []SubPath{s.SubPath()}
				completedpath := append(adjustedpath, sp...)

				elempaths = append(elempaths, completedpath)
			}
		}

		allpaths = append(allpaths, elempaths...)
	}
	return allpaths, nil

}

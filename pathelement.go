package whatnot

import (
	"fmt"
	"strings"
	"sync"

	"github.com/databeast/whatnot/mutex"
	"github.com/pkg/errors"
)

// PathElement is an individual section of a complete path
type PathElement struct {
	// the individual name of this Path Element
	section SubPath

	// The parent Path Element
	parent *PathElement

	// channel to notify parent element of changes to this element
	// or any of its sub-elements
	parentnotify chan elementChange
	mu           *mutex.SmartMutex
	reslock      resourceLock
	val          ElementValue

	// sub Path-elements directly beneath this Path Elements
	subs map[SubPath]*PathElement

	// channel for incoming change notifications from any of the
	// children of this Path Elements
	subevents chan elementChange

	// channel for events directly on this element itself
	events chan elementChange

	// Channel Multiplexer for sending watch events to things subscribed
	// to events on this Path Element or any of its children
	subscriberNotify *EventMultiplexer
}

// SubPath returns the name of this Path Element
func (m *PathElement) SubPath() (path SubPath) {
	return m.section
}

// Parent returns the parent PathElement of this PathElement
func (m *PathElement) Parent() *PathElement {
	return m.parent
}

func (m *PathElement) ParentChain() (parents []*PathElement) {
	var nextParent *PathElement
	if m.parent == nil {
		return
	} else {
		nextParent = m.parent
	}

	for {
		if nextParent.section == rootId {
			break
		}
		parents = append([]*PathElement{nextParent}, parents...) // prepend it into the list, to the items are in path order
		nextParent = nextParent.Parent()
	}
	return parents
}

// Chain returns the full Path of this Element, as a slice of individual PathElements
func (m *PathElement) Chain() (chain []*PathElement) {
	return append(m.ParentChain(), m)
}

// AbsolutePath returns the full Path of this Element, as a single AbsolutePath instance
func (m *PathElement) AbsolutePath() (path AbsolutePath) {
	for _, element := range m.Chain() {
		path = append(path, element.SubPath())
	}
	return path
}

// fetchSubElement fetches named sub element, if it exists
// returns nil if no sub element by that name exists
func (m PathElement) fetchSubElement(path SubPath) *PathElement {
	sub, ok := m.subs[path]
	if ok {
		return sub
	} else {
		return nil
	}
}

func (m PathElement) logChange(e elementChange) {
	switch e.change {
	default:
		// nothing for now - placeholder for later audit logging
	}

}

// FetchClosestSubPathTail finds the last element in a path chain that most closely resembles the requested path
func (m PathElement) FetchClosestSubPathTail(subPath PathString) *PathElement {
	elemChain := m.FetchClosestSubPath(subPath)
	if len(elemChain) > 0 {
		return elemChain[len(elemChain)-1]
	} else {
		return nil
	}
}

// FetchClosestSubPath
func (m *PathElement) FetchClosestSubPath(subPath PathString) (pathchain []*PathElement) {
	elems := splitPath(subPath)
	var finalElement *PathElement
	var currentElement = m
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
func (m *PathElement) Add(path SubPath) (elem *PathElement, err error) {

	err = path.Validate()
	if err != nil {
		return nil, err
	}

	m.mu.Lock()

	if v, ok := m.subs[path]; ok { // element already exists, do not overwrite
		elem = v
	} else {
		elem = &PathElement{
			section:          path,
			parent:           m,
			parentnotify:     m.subevents,
			mu:               mutex.New(fmt.Sprintf("internal mutex for %s", path)),
			subs:             make(map[SubPath]*PathElement),
			subevents:        make(chan elementChange),
			subscriberNotify: newEventsMultiplexer(),
		}
		m.subs[path] = elem
		elem.reslock = resourceLock{
			selfmu:    &sync.Mutex{},
			resmu:     &sync.Mutex{},
			recursive: false,
		}
	}
	elem.watchChildren() // start the signal handler going

	m.mu.Unlock()
	return elem, nil
}

// attach an existing PathElement to a parent PathElement
func (m *PathElement) attach(elem *PathElement) (err error) {

	// check this is a properly formed element
	if elem.subs == nil {
		elem.subs = make(map[SubPath]*PathElement)
	}

	m.mu.Lock()
	m.subs[elem.SubPath()] = elem
	m.mu.Unlock()

	return nil
}

// AppendRelativePath constructs an element-relative subpath, append it to an Existing PathElement,
// creating all PathElements along the way
func (m *PathElement) AppendRelativePath(subPath PathString) (*PathElement, error) {
	// subpaths cannot be absolute, so they cannot start with the delimeter
	if strings.HasPrefix(string(subPath), pathDelimeter) {
		return nil, errors.Errorf("cannot use an absolute path as a subpath")
	}

	pathElems := subPath.ToRelativePath()
	var cur = m
	var err error
	for _, e := range pathElems {
		cur, err = m.Add(e)
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
func (m *PathElement) subtractPathToSubPaths(path PathString) (newSubPath []SubPath) {
	return
}

func (m *PathElement) FetchSubPath(subPath PathString) (*PathElement, error) {

	// subpaths cannot be absolute, so they cannot start with the delimeter
	if strings.HasPrefix(string(subPath), pathDelimeter) {
		return nil, errors.Errorf("cannot use an absolute path as a subpath")
	}

	pathElems := subPath.ToRelativePath()
	var cur = m
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
func (m *PathElement) FetchAllSubPaths() (allpaths [][]SubPath, err error) {
	for _, s := range m.subs {
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

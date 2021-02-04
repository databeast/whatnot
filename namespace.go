package whatnot

import (
	"fmt"

	"github.com/databeast/whatnot/mutex"
)

// Namespace provides unique namespaces for keyval trees
type Namespace struct {
	root     *PathElement
	name     string
	globalmu *mutex.SmartMutex
	events   chan elementChange
}

// NewNamespace creates a new Namespace Instance. If this is intended to be persisted
// it should be registed to a NamespaceManager via RegisterNameSpace
func NewNamespace(name string) (ns *Namespace) {
	ns = &Namespace{
		name:     name,
		globalmu: mutex.New(fmt.Sprintf("Global mutex for namespace %q", name)),
		events:   make(chan elementChange),
	}

	ns.root = &PathElement{
		section:      ROOT_ID,
		mu:           mutex.New("Namespace Root Element mutex"),
		subs:         make(map[SubPath]*PathElement),
		subevents:    make(chan elementChange),
		parentnotify: ns.events,
	}
	return ns
}

// RegisterAbsolutePath constructs a complete path in the Namespace, with all required
// structure instances to make the path immediately available and active
func (m *Namespace) RegisterAbsolutePath(path AbsolutePath) error {
	var currentElement = m.root
	var err error
	for _, p := range path {
		currentElement, err = currentElement.Add(p)
		if err != nil {
			return err
		}
	}
	return nil
}

// FetchAbsolutePath will return the PathElement instance at the end of the provided Path
// assuming it exists, otherwise it returns Nil
func (m *Namespace) FetchAbsolutePath(path PathString) *PathElement {
	abspath := path.ToAbsolutePath()
	lastElem := m.FindPathTail(path)

	// if the lengths dont match, we definitely dont have a match
	lastElemPath := lastElem.AbsolutePath()
	if len(lastElemPath) != len(abspath) {
		return nil
	}
	// if they do, lets be certain they are identical
	for i, p := range lastElemPath {
		if abspath[i] != p {
			return nil // path mismatch
		}
	}
	return lastElem
}

// FindPathTailFetch attempts to locate the last element that most closely matches the given path fragment
// if no suitable match can be found, it returns Nil, if multiple elements are found, it returns the first
// one going from alphabetically-sorted pathing
func (m *Namespace) FindPathTail(path PathString) *PathElement {
	return m.root.FetchClosestSubPathTail(path)
}

// FetchAllAbsolutePaths returns an array of all distinct terminayted absolute paths
// effectively dumping all possible paths in the entire namespace
func (m *Namespace) FetchAllAbsolutePaths() (allpaths []AbsolutePath, err error) {
	all, err := m.root.FetchAllSubPaths()
	if err != nil {
		return nil, err
	}
	for _, a := range all {
		allpaths = append(allpaths, a)
	}
	return allpaths, nil

}

type NamespaceManagerOpt interface {
	opt()
}

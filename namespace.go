package whatnot

import (
	"fmt"

	"github.com/databeast/whatnot/mutex"
)

/////////

// Namespace provides unique namespaces for keyval trees
type Namespace struct {
	root     *PathElement
	name     string
	globalmu *mutex.SmartMutex
	events   chan elementChange
}

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

func (m *Namespace) RegisterAbsolutePath(path AbsolutePath) error {
	var currentElement *PathElement = m.root
	var err error
	for _, p := range path {
		currentElement, err = currentElement.Add(p)
		if err != nil {
			return err
		}
	}
	return nil
}

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

// Fetch the last element that most closely matches the given path
func (m *Namespace) FindPathTail(path PathString) *PathElement {
	return m.root.FetchClosestSubPathTail(path)
}

// return an array of all distinct terminal absolute  paths
func (m *Namespace) FetchAllAbsolutePaths() (allpaths []AbsolutePath, err error) {
	all, err := m.root.FetchAllSubPaths()
	if err != nil {
		return nil, err
	}
	for _, a := range all {
		allpaths = append(allpaths, AbsolutePath(a))
	}
	return allpaths, nil

}

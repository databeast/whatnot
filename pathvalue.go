package whatnot

import "github.com/databeast/whatnot/access"

type ElementValue struct {
	Val interface{}
}

func (m *PathElement) SetValue(value ElementValue, change changeType, actor access.Role) {
	if m == nil {
		panic("SetValue called on nil PathElement")
	}
	m.mu.Lock()
	m.resval = value
	m.mu.Unlock()

	m.parentnotify <- elementChange{elem: m, change: change, actor: actor}
}

func (m *PathElement) GetValue() (value ElementValue) {
	if m == nil {
		panic("GetValue called on nil PathElement")
	}
	return m.resval
}

package whatnot

import "github.com/databeast/whatnot/access"

type ElementValue struct {
	Val interface{}
}

func (p *PathElement) SetValue(value ElementValue, change changeType, actor access.Role) {
	if p == nil {
		panic("SetValue called on nil PathElement")
	}
	p.mu.Lock()
	p.resval = value
	p.mu.Unlock()

	p.parentnotify <- elementChange{elem: p, change: change, actor: actor}
}

func (p *PathElement) GetValue() (value ElementValue) {
	if p == nil {
		panic("GetValue called on nil PathElement")
	}
	return p.resval
}

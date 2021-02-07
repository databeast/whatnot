package whatnot

import (
	"time"

	"github.com/databeast/whatnot/access"
)

type changeType int

const (
	ChangeUnknown changeType = iota + 1
	ChangeLocked
	ChangeUnlocked
	ChangeAdded
	ChangeEdited
	ChangeDeleted
)

// elementChange is a notification channel structure
// for communicating changes to individual elements to subscribed watchers
type elementChange struct {
	elem   *PathElement
	change changeType
	actor  access.Role
}

// new key additions recheck those lists, add watches on keys ?

// PathSubscription implements a notification Subscription
// to a given PathElement (and potentially all its descendents)
type PathSubscription struct {
	baseElement *PathElement
}

type ElementWatchSubscription struct {
	onElement   *PathElement // the Path Element this is a subscription to
	isRecursive bool         // is this subscription for this element alone, or its children as well?
	events      chan WatchEvent
}

type WatchEvents chan WatchEvent

type WatchEvent struct {
	elem   *PathElement
	TS     time.Time
	Change changeType
	Actor  access.Role
	Note   string
}

func (e WatchEvent) OnElement() *PathElement {
	return e.elem
}
func (m *PathElement) SubscribeToEvents(prefix bool) *ElementWatchSubscription {
	sub := &ElementWatchSubscription{
		onElement:   m,
		events:      make(chan WatchEvent),
		isRecursive: prefix,
	}

	m.subscriberNotify.Register(sub.events) // this is the part that will allow us to receive channel messages

	return sub
}

func (m *PathElement) UnSubscribeFromEvents() error {
	return nil
}

// watchChildren is a PathElement-specific goroutine to handle event channels
// in default mode, this means that goroutine load scales 1-to-1 with total
// number of distinct pathelements
func (m *PathElement) watchChildren() {
	go func() {
		var e elementChange
		for {
			select {
			// process events from our children
			case e = <-m.subevents:
				if e.elem == nil {
					panic("elementChange event passed with nil PathElement")
				}

			// process events from ourself
			case e = <-m.selfevents:
				if e.elem == nil {
					panic("elementChange event passed with nil PathElement")
				}
			}

			if e.elem == nil {
				panic("elementChange event passed with nil PathElement")
			}
			pe := e // clone our event to send upwards, make the data race analyzer happy
			go func() { m.parentnotify <- pe }()
			// then do what we need to do with the event ourselves now
			m.logChange(e)

			// Broadcast the change out to all subscribers
			m.subscriberNotify.Broadcast <- WatchEvent{
				elem:   e.elem,
				TS:     time.Now(),
				Note:   "",
				Change: e.change,
				Actor:  e.actor,
			}
			// TODO: needs close handler
		}

	}()
}

// Events returns a channel of subscriberNotify occurring to this Key (or its subKeys
func (m *ElementWatchSubscription) Events() <-chan WatchEvent {
	return m.events
}

type subscriberStats struct{}

const (
	defaultMultiplexerBuffer = 100
)

// client creates a channel
// client registers channel with Topic
// client recieves messages over that channel

func (m *Namespace) WatchEventsOnPath() {

	// find Topic for

}

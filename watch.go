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

// ElementWatchSubscription is a contract to be notified
// of all events on a given Path Element, and optionally
// all its child elements
type ElementWatchSubscription struct {
	logsupport
	onElement   *PathElement // the Path Element this is a subscription to
	isRecursive bool         // is this subscription for this element alone, or its children as well?
	events      chan WatchEvent
}

type WatchEvents chan WatchEvent

// WatchEvent describes an event on a Path Element or optionally
// any of its children, obtained and consumed via an ElementWatchSubscription
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

// SubscribeToEvents generates a Watch Subscription that produces a single channel
// of notification events on the accompanying Path Element, and optionally all of its
// child path elements
func (m *PathElement) SubscribeToEvents(prefix bool) *ElementWatchSubscription {
	sub := &ElementWatchSubscription{
		onElement:   m,
		events:      make(chan WatchEvent),
		isRecursive: prefix,
	}

	m.subscriberNotify.Register(sub.events, prefix) // this is the part that will allow us to receive channel messages

	return sub
}

// UnSubscrulibeFromEvents will unregister the notification channel
// and then nil out the watch subscription that is passed to it
func (m *PathElement) UnSubscribeFromEvents(sub *ElementWatchSubscription) {
	m.subscriberNotify.Unregister(sub.events)
	sub = nil
}

// watchChildren is a PathElement-specific goroutine to handle event channels
// in default mode, this means that goroutine load scales 1-to-1 with total
// number of distinct pathelements
func (m *PathElement) watchChildren() {
	go func() {
		var e elementChange
		for {
			select {
			case e = <-m.subevents:
				// process events from our children
				m.Debugf("%s received change notify from child: %s", m.AbsolutePath().ToPathString(), e.elem.AbsolutePath().ToPathString())
			case e = <-m.selfevents:
				// process events from ourself
				m.Debugf("%s received change notify on self", m.AbsolutePath().ToPathString())
			}

			if e.elem == nil {
				panic("elementChange event passed with nil PathElement")
			}
			pe := e // clone our event to send upwards, make the data race analyzer happy
			if m.parent.section !=  rootId {
				m.parentnotify <- pe
			}

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



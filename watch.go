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
	id     uint64
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
	id     uint64
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

// UnSubscribeFromEvents will unregister the notification channel
// and then nil out the watch subscription that is passed to it.
// preventing any further reception of events
func (m *PathElement) UnSubscribeFromEvents(sub *ElementWatchSubscription) {
	m.subscriberNotify.Unregister(sub.events)
	sub = nil
}

// Events returns a channel of subscriberNotify occurring to this Key (or its subKeys
func (m *ElementWatchSubscription) Events() <-chan WatchEvent {
	return m.events
}

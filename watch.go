package whatnot

import (
	"sync"
	"time"

	"github.com/databeast/whatnot/access"
)

type changeType int

const (
	UNKNOWN  changeType = 0
	LOCKED   changeType = 1
	UNLOCKED changeType = 2
	ADDED    changeType = 3
	EDITED   changeType = 4
	DELETED  changeType = 5
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

	m.subscriberNotify.Register(sub.events) // this is the part that will allow us to recieve channel messages

	return sub
}

func (m *PathElement) UnSubscribeFromEvents() error {
	return nil
}

// Each PathElement gets its own goroutine to handle event channels
func (m *PathElement) watchChildren() {
	go func() {
		var e elementChange
		for {
			select {
			case e = <-m.subevents:
				// first pass the message on
				if e.elem == nil {
					panic("elementChange subevent passed with nil PathElement")
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
			}
			// TODO: needs close handler
		}

	}()
}

// Events returns a channel of subscriberNotify occurring to this Key (or its subKeys
func (m *ElementWatchSubscription) Events() <-chan WatchEvent {
	return m.events
}

type nothing struct{}

// Topic is a pub-sub mechanism where consumers can Register to
// receive messages sent to Broadcast.
type EventMultiplexer struct {
	// Producer sends messages on this channel. Close the channel
	// to shutdown the topic.
	Broadcast chan<- WatchEvent

	lock        sync.Mutex
	connections map[chan<- WatchEvent]nothing
}

// New creates a new topic. Messages can be broadcast on this topic,
// and registered consumers are guaranteed to either receive them, or
// see a channel close.
func newEventsMultiplexer() *EventMultiplexer {
	t := &EventMultiplexer{}
	broadcast := make(chan WatchEvent, 100)
	t.Broadcast = broadcast
	t.connections = make(map[chan<- WatchEvent]nothing)
	go t.run(broadcast)
	return t
}

func (t *EventMultiplexer) run(broadcast <-chan WatchEvent) {
	for msg := range broadcast {
		func() {
			t.lock.Lock()
			defer t.lock.Unlock()
			for ch, _ := range t.connections {
				select {
				case ch <- msg:
				default:
					delete(t.connections, ch)
					close(ch)
				}
			}
		}()
	}

	t.lock.Lock()
	defer t.lock.Unlock()
	for ch, _ := range t.connections {
		delete(t.connections, ch)
		close(ch)
	}
}

// Register starts receiving messages on the given channel. If a
// channel close is seen, either the topic has been shut down, or the
// consumer was too slow, and should re-register.
func (t *EventMultiplexer) Register(ch chan<- WatchEvent) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.connections[ch] = nothing{}
}

// Unregister stops receiving messages on this channel.
func (t *EventMultiplexer) Unregister(ch chan<- WatchEvent) {
	t.lock.Lock()
	defer t.lock.Unlock()

	// double-close is not safe, so make sure we didn't already
	// drop this consumer as too slow
	_, ok := t.connections[ch]
	if ok {
		delete(t.connections, ch)
		close(ch)
	}
}

// client creates a channel
// client registers channel with Topic
// client recieves messages over that channel

func (m *Namespace) WatchEventsOnPath() {

	// find Topic for

}

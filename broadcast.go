package whatnot

import (
	"github.com/databeast/whatnot/access"
	"sync"
	"time"
)

/*
Whatnot depends on multiple subscribers to a notification channel

The following is an implementation of GO channel multiplexing
*/

const (
	defaultMultiplexerBuffer = 100
)

// EventMultiplexer is a pub-sub mechanism where consumers can Register to
// receive messages sent to Broadcast.
type EventMultiplexer struct {
	logsupport
	// Broadcast is the channel to set events to for them to be multiplexed out
	Broadcast chan<- WatchEvent

	lock *sync.Mutex

	onElement *PathElement

	connections map[chan<- WatchEvent]bool
}

// Register starts receiving messages on the given channel. If a
// channel close is seen, either the topic has been shut down, or the
// consumer was too slow, and should re-register.
func (t *EventMultiplexer) Register(ch chan<- WatchEvent, recursive bool) {
	t.lock.Lock()
	t.connections[ch] = recursive
	t.lock.Unlock()
}

// Unregister stops receiving messages on this channel.
func (t *EventMultiplexer) Unregister(ch chan<- WatchEvent) {
	t.lock.Lock()

	// double-close is not safe, so make sure we didn't already
	// drop this consumer as too slow
	_, ok := t.connections[ch]
	if ok {
		delete(t.connections, ch)
		close(ch)
	}
	t.lock.Unlock()
}

// initEventBroadcast creates a new event multiplexer
// Messages can be broadcast on this topic,
// and registered consumers are guaranteed to either receive them, or
// see a channel close.
func (p *PathElement) initEventBroadcast() {
	if p.subscriberNotify != nil {
		// simple reentrance
		return
	}
	p.subscriberNotify = NewEventsMultiplexer()
	p.subscriberNotify.onElement = p
	// start reading incoming events from child elements
	p.watchChildren()
}

// NewEventsMultiplexer creates a new event multiplexer
// that will duplicate incoming WatchEvents to multiple watcher
// channels
func NewEventsMultiplexer() *EventMultiplexer {

	var broadcast = make(chan WatchEvent, defaultMultiplexerBuffer)
	t := &EventMultiplexer{
		Broadcast:   broadcast,
		lock:        &sync.Mutex{},
		onElement:   nil,
		connections: make(map[chan<- WatchEvent]bool, 5), //give this a bit of a buffer so slow subscribers can respond
	}

	go t.run(broadcast)
	return t
}

// BroadcastAsync has the multiplexer submit the WatchEvent
// instead of the caller attaching directly to a channel
// delivery is not guaranteed in this case and the goroutine
// will eventually exit if it deadlocks
func (t *EventMultiplexer) BroadcastAsync(evt WatchEvent) {
	go func() {
		t.Broadcast <- evt
	}()
	// TODO: kill goroutine after max wait time
}

// run is the primary goroutine loop for each Multiplexer
// to shut it down, send a channel close to the multiplexer's Broadcast channel
func (t *EventMultiplexer) run(broadcastchan <-chan WatchEvent) {
	for msg := range broadcastchan {
		func() {
			// send our broadcast event to every subscriber
			for ch, rec := range t.connections {
				if rec == false {
					if msg.elem != t.onElement {
						continue // this connect only wants notifications about itself
					}
				}
				select {
				case ch <- msg:
					// sends event to individual multiplexer subscriber
					t.Debugf("transmitted event %d to broadcast subscriber", msg.id)
				default:
					// cannot send message, listener has closed the channel
					t.lock.Lock()
					t.Debugf("%s removing disconnected subscriber", t.onElement.AbsolutePath().ToPathString())
					delete(t.connections, ch)
					close(ch)
					t.lock.Unlock()
				}
			}
		}()
	}

	// broadcast channel has been closed at this point
	t.lock.Lock()
	for ch := range t.connections {
		delete(t.connections, ch)
		close(ch)
	}
	t.lock.Unlock()
}

// watchChildren is a PathElement-specific goroutine to handle event channels
// in default mode, this means that goroutine load scales 1-to-1 with total
// number of distinct pathelements
func (p *PathElement) watchChildren() {
	go func() {
		var e elementChange
		for {
			if p.prunectx != nil { // need to watch for pruning cancellation
				select {
				case e = <-p.subevents:
					// process events from our children
					p.Debugf("%s received change notify %d from child: %s", p.AbsolutePath().ToPathString(), e.id, e.elem.AbsolutePath().ToPathString())
				case e = <-p.selfnotify:
					// process events from ourself
					p.Debugf("%s received change notify on self", p.AbsolutePath().ToPathString())
				case <-p.prunectx.Done():
					// element is being pruned, we need to shut down this monitoring goroutine
					// but first send this final notification upwards
					p.Debugf("pruning signal received - shutting down event watch goroutine")
					p.parentnotify <- elementChange{
						id:     randid.Uint64(),
						elem:   p,
						change: ChangePruned,
						actor:  access.Role{},
					}

					// then do what we need to do with the event ourselves now
					p.logChange(e)

					// Broadcast the change out to all subscribers
					p.subscriberNotify.Broadcast <- WatchEvent{
						id:     randid.Uint64(),
						elem:   p,
						TS:     time.Now(),
						Note:   "",
						Change: ChangePruned,
						Actor:  e.actor,
					}
					return
				}
			} else { // just watch for path element change events
				select {
				case e = <-p.subevents:
					// process events from our children
					p.Debugf("%s received change notify %d from child: %s", p.AbsolutePath().ToPathString(), e.id, e.elem.AbsolutePath().ToPathString())
				case e = <-p.selfnotify:
					// process events from ourself
					p.Debugf("%s received change notify on self", p.AbsolutePath().ToPathString())
				}
			}

			if e.elem == nil {
				panic("elementChange event passed with nil PathElement")
			}
			pe := e // clone our event to send upwards, make the data race analyzer happy
			p.parentnotify <- pe

			// then do what we need to do with the event ourselves now
			p.logChange(e)

			// Broadcast the change out to all subscribers
			p.subscriberNotify.Broadcast <- WatchEvent{
				id:     e.id,
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

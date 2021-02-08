package whatnot

import "sync"

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
func (m *PathElement) initEventBroadcast() {
	if m.subscriberNotify != nil {
		// simple reentrance
		return
	}
	m.subscriberNotify = NewEventsMultiplexer()
	m.subscriberNotify.onElement = m
	// start reading incoming events from child elements
	m.watchChildren()
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
		connections: make(map[chan<- WatchEvent]bool),
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
					t.Debug("transmitted event to broadcast subscriber")
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

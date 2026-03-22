package stream

import (
	"sync"
)

type ClientChannel chan []byte

type Dispatcher struct {
	sync.RWMutex
	subscribers map[string][]ClientChannel // key = topic:symbol
}

// NewDispatcher creates a new dispatcher instance
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		subscribers: make(map[string][]ClientChannel),
	}
}

// Subscribe registers a client to receive messages for a specific topic and symbol
func (d *Dispatcher) Subscribe(topic, symbol string) ClientChannel {
	key := topic + ":" + symbol
	ch := make(ClientChannel, 100) // buffered channel

	d.Lock()
	d.subscribers[key] = append(d.subscribers[key], ch)
	d.Unlock()

	return ch
}

// Broadcast sends the message to all clients subscribed to a topic and symbol
func (d *Dispatcher) Broadcast(topic, symbol string, msg []byte) {
	key := topic + ":" + symbol

	d.RLock()
	defer d.RUnlock()

	if chans, ok := d.subscribers[key]; ok {
		for _, ch := range chans {
			select {
			case ch <- msg:
			default:
				// drop message if channel is full (avoiding block)
			}
		}
	}
}

// Unsubscribe removes a client channel from the subscribers list
func (d *Dispatcher) Unsubscribe(topic, symbol string, ch ClientChannel) {
	key := topic + ":" + symbol

	d.Lock()
	defer d.Unlock()

	if chans, ok := d.subscribers[key]; ok {
		newChans := make([]ClientChannel, 0, len(chans))
		for _, c := range chans {
			if c != ch {
				newChans = append(newChans, c)
			}
		}
		d.subscribers[key] = newChans
	}
}

package broadcaster

import (
	"sync"
)

// Broadcaster sends string values to subscribers by topic.
type Broadcaster struct {
	mu   sync.RWMutex
	subs map[string]map[chan string]struct{}
}

// New creates a new Broadcaster.
func New() *Broadcaster {
	return &Broadcaster{
		subs: make(map[string]map[chan string]struct{}),
	}
}

// Subscribe returns a channel that receives values broadcast to the topic.
func (b *Broadcaster) Subscribe(topic string) chan string {
	ch := make(chan string, 1)
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.subs[topic] == nil {
		b.subs[topic] = make(map[chan string]struct{})
	}
	b.subs[topic][ch] = struct{}{}
	return ch
}

// Unsubscribe removes the channel and closes it.
func (b *Broadcaster) Unsubscribe(topic string, ch chan string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if m, ok := b.subs[topic]; ok {
		delete(m, ch)
		if len(m) == 0 {
			delete(b.subs, topic)
		}
	}
	close(ch)
}

// Broadcast sends value to all subscribers of the topic.
func (b *Broadcaster) Broadcast(topic, value string) {
	b.mu.RLock()
	subs := make([]chan string, 0, len(b.subs[topic]))
	for ch := range b.subs[topic] {
		subs = append(subs, ch)
	}
	b.mu.RUnlock()

	for _, ch := range subs {
		select {
		case ch <- value:
		default:
		}
	}
}

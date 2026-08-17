package events

import (
	"errors"
	"io"
	"sync"
)

var _ = Reader[any](&Bus[any]{})

var TopicAll string = "*"

type Event[T any] struct {
	Topic string
	Data  T
}

type Bus[T any] struct {
	blocking      bool
	subscribersMu sync.RWMutex
	subscribers   map[string]map[chan Event[T]]struct{}
}

type BusConfig struct {
	// Blocking defines whether or not publishing needs to make sure all listeners consumed it.
	// This is dangerous when used carelessly as it could lead in infinite waiting.
	//
	// Default: false
	Blocking bool
}

func NewBus[T any](cfg ...BusConfig) *Bus[T] {
	if len(cfg) == 0 {
		cfg = []BusConfig{{}}
	}

	return &Bus[T]{
		blocking:    cfg[0].Blocking,
		subscribers: make(map[string]map[chan Event[T]]struct{}),
	}
}

func (b *Bus[T]) Publish(topic string, p T) (err error) {
	if topic == "*" || len(topic) == 0 {
		return errors.New("topic must be of higher specificity")
	}

	b.subscribersMu.RLock()
	defer b.subscribersMu.RUnlock()

	targets := make(map[chan Event[T]]struct{})
	for ch := range b.subscribers[topic] {
		targets[ch] = struct{}{}
	}
	for ch := range b.subscribers["*"] {
		targets[ch] = struct{}{}
	}

	event := Event[T]{
		Topic: topic,
		Data:  p,
	}

	for target := range targets {
		if b.blocking {
			target <- event
			continue
		}

		// Non-blocking publush, if a client is taking too long we dismiss it
		select {
		case target <- event:
		default:
		}
	}

	return nil
}

func (b *Bus[T]) Writer(topic string) io.Writer {
	return &topicWriter[T]{
		bus:   b,
		topic: topic,
	}
}

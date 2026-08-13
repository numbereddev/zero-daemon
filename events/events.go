package events

import (
	"fmt"
	"io"
	"sync"
)

const DefaultBufferSize = 10

// make sure Bus implements io.Writer and ReaderBus
var (
	_ = io.Writer(&Bus[any]{})
	_ = Reader[any](&Bus[any]{})
)

type Bus[T any] struct {
	blocking      bool
	subscribersMx sync.RWMutex
	// TODO: topics
	subscribers map[chan T]bool
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
		subscribers: make(map[chan T]bool),
	}
}

func (b *Bus[T]) Publish(p T) (sent int, total int) {
	b.subscribersMx.RLock()
	defer b.subscribersMx.RUnlock()

	total = len(b.subscribers)
	sent = total

	for sub := range b.subscribers {
		if b.blocking {
			sub <- p
			continue
		}

		// Non-blocking publush, if a client is taking too long we dismiss it
		select {
		case sub <- p:
		default:
			sent--
		}
	}

	return
}

// Write implements [io.Writer].
func (b *Bus[T]) Write(p []byte) (n int, err error) {
	var zero T

	switch any(zero).(type) {
	case []byte:
		b.Publish(any(p).(T))
	case string:
		b.Publish(any(string(p)).(T))
	default:
		return 0, fmt.Errorf("Bus.Write only supports []byte or string, got %T", zero)
	}

	return len(p), nil
}

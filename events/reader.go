package events

const DefaultBufferSize = 10

type Reader[T any] interface {
	On(topics []string, c ...int) chan Event[T]
	Off(channel chan Event[T])
}

func (b *Bus[T]) On(topics []string, buffer ...int) chan Event[T] {
	b.subscribersMu.Lock()
	defer b.subscribersMu.Unlock()

	// Ensure the array is never empty
	if len(buffer) == 0 {
		buffer = []int{DefaultBufferSize}
	}

	// Create channel and provide it
	channel := make(chan Event[T], buffer[0])
	for _, topic := range topics {
		if len(topic) == 0 {
			topic = "*"
		}

		if _, exists := b.subscribers[topic]; !exists {
			b.subscribers[topic] = make(map[chan Event[T]]struct{})
		}

		b.subscribers[topic][channel] = struct{}{}
	}

	return channel
}

func (b *Bus[T]) Off(channel chan Event[T]) {
	b.subscribersMu.Lock()
	defer b.subscribersMu.Unlock()

	for topic := range b.subscribers {
		delete(b.subscribers[topic], channel)
	}

	close(channel)
}

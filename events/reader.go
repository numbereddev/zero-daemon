package events

type Reader[T any] interface {
	On(c ...OnOpts) chan T
	Off(channel chan T)
}

type OnOpts struct {
	// TODO: topics
	Buffer *int
}

func (b *Bus[T]) On(opts ...OnOpts) chan T {
	b.subscribersMx.Lock()
	defer b.subscribersMx.Unlock()

	// Ensure the array is never empty
	if len(opts) == 0 {
		opts = []OnOpts{{}}
	}

	// Initialize defaults

	BufferSize := opts[0].Buffer
	if BufferSize == nil {
		BufferSize = new(DefaultBufferSize)
	}

	// Create channel and provide it
	channel := make(chan T, *BufferSize)
	b.subscribers[channel] = true
	return channel
}

func (b *Bus[T]) Off(channel chan T) {
	b.subscribersMx.Lock()
	defer b.subscribersMx.Unlock()

	delete(b.subscribers, channel)
	close(channel)
}

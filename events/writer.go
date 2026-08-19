package events

import (
	"fmt"
	"io"
)

var _ = io.Writer(&topicWriter[any]{})

type topicWriter[T any] struct {
	bus   *Bus[T]
	topic string
}

// Write implements [io.Writer].
func (w *topicWriter[T]) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	// Make copy of byte slice so caller buffer reuse (e.g. inside stdcopy
	// or scanners) won't mutate data asynchronously
	buf := make([]byte, len(p))
	copy(buf, p)

	var typedBuf T
	switch any(typedBuf).(type) {
	case []byte:
		typedBuf = any(buf).(T)
	case string:
		typedBuf = any(string(buf)).(T)
	default:
		return 0, fmt.Errorf("Write() only supports []byte or string, got %T", typedBuf)
	}

	w.bus.Publish(w.topic, typedBuf)
	return len(buf), nil
}

package server

import (
	"sync"

	"github.com/moby/moby/client"
	"github.com/numbereddev/zero-daemon/events"
	"github.com/numbereddev/zero-daemon/internal/system"
)

const (
	EventImagePullProgress = "image.pull.progress" // event.Data is of type ImagePullProgress
	EventImagePullBegin    = "image.pull.begin"
	EventImagePullDone     = "image.pull.done"
	EventStateChange       = "state.change"
	EventStateError        = "state.error"
	EventConsoleOut        = "console.output"
)

// TODO: crash detector
// TODO: start/running detector
// TODO: historical logs getting

type ImagePullProgress struct {
	Status  string
	Current uint64
	Total   uint64
}

type Server struct {
	mu  sync.RWMutex
	cli *client.Client

	id     string
	status system.AtomicString

	stream *client.HijackedResponse
	bus    *events.Bus[string]
}

func New(cli *client.Client, id string) *Server {
	// TODO: Configuration
	return &Server{
		id:     id,
		cli:    cli,
		status: *system.NewAtomicString("offline"),
		bus:    events.NewBus[string](),
	}
}

func (s *Server) ID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.id
}

func (s *Server) Events() events.Reader[string] {
	return s.bus
}

func (s *Server) IsAttached() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stream != nil
}

func (s *Server) Status() string {
	return s.status.Load()
}

func (s *Server) setStatus(state string) {
	if s.Status() == state {
		return
	}

	s.status.Store(state)
	s.bus.Publish(EventStateChange, state)
}

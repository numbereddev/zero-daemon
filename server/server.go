package server

import (
	"sync"

	"github.com/moby/moby/client"
	"github.com/numbereddev/zero-daemon/events"
	"github.com/numbereddev/zero-daemon/internal/system"
)

const (
	EventImagePullProgress = "image.pull.progress" // Data is of type ImagePullProgress
	EventImagePullBegin    = "image.pull.begin"
	EventImagePullDone     = "image.pull.done"
	EventStateChange       = "state.change"
	EventStateError        = "state.error"
	EventConsoleOut        = "console.output"
)

// TODO: crash detector
// TODO: start/running detector
// TODO: historical logs getting

type Server struct {
	mu  sync.RWMutex
	cli *client.Client

	id    string
	state system.AtomicString

	stream *client.HijackedResponse
	bus    *events.Bus[string]
}

func New(cli *client.Client, id string) *Server {
	// TODO: Configuration
	return &Server{
		id:    id,
		cli:   cli,
		state: *system.NewAtomicString("offline"),
		bus:   events.NewBus[string](),
	}
}

func (r *Server) ID() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.id
}

func (r *Server) Events() events.Reader[string] {
	return r.bus
}

func (r *Server) State() string {
	return r.state.Load()
}

func (r *Server) SetState(state string) {
	if r.State() == state {
		return
	}

	r.state.Store(state)
	r.bus.Publish(EventStateChange, state)
}

func (r *Server) IsAttached() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.stream != nil
}

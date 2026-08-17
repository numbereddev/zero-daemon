package runtime

import (
	"sync"

	"github.com/moby/moby/client"
	"github.com/numbereddev/zero-daemon/events"
	"github.com/numbereddev/zero-daemon/internal/system"
)

const (
	EventImagePullProgress = "image.pull.progress"
	EventImagePullError    = "image.pull.error"
	EventImagePullBegin    = "image.pull.begin"
	EventImagePullDone     = "image.pull.done"
	EventStateChange       = "state.change"
	EventConsoleOut        = "console.output"
)

// TODO: crash detector
// TODO: start/running detector
// TODO: historical logs getting

type Runtime struct {
	mu  sync.RWMutex
	cli *client.Client

	id     string
	state  system.AtomicString
	stream *client.HijackedResponse

	bus *events.Bus[string]
}

// TODO: Configuration
func New(cli *client.Client, id string) *Runtime {
	return &Runtime{
		id:    id,
		cli:   cli,
		state: *system.NewAtomicString("offline"),
		bus:   events.NewBus[string](),
	}
}

func (r *Runtime) ID() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.id
}

func (r *Runtime) Events() events.Reader[string] {
	return r.bus
}

func (r *Runtime) State() string {
	return r.state.Load()
}

func (r *Runtime) SetState(state string) {
	if r.State() == state {
		return
	}

	r.state.Store(state)
	_ = r.bus.Publish(EventStateChange, state)
}

func (r *Runtime) IsAttached() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.stream != nil
}

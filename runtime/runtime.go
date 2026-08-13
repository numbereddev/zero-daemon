package runtime

import (
	"context"
	"sync"
	"time"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/client"
	"github.com/numbereddev/zero-daemon/events"
)

type Runtime struct {
	mu     sync.RWMutex
	client *client.Client

	id string

	// TODO: needs to be reworked, just temp
	state   string
	stateMx sync.Mutex

	stream   *client.HijackedResponse
	eventBus *events.Bus[[]byte]
}

// TODO: Configuration
func New(cli *client.Client, id string) *Runtime {
	return &Runtime{
		id:       id,
		client:   cli,
		eventBus: events.NewBus[[]byte](),
	}
}

func (r *Runtime) ID() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.id
}

func (r *Runtime) Events() events.Reader[[]byte] {
	return r.eventBus
}

func (r *Runtime) IsAttached() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.stream != nil
}

func (r *Runtime) setStream(stream *client.HijackedResponse) {
	r.mu.Lock()
	r.stream = stream
	r.mu.Unlock()
}

func (r *Runtime) State() string {
	r.stateMx.Lock()
	defer r.stateMx.Unlock()
	return r.state
}

func (r *Runtime) SetState(action string) {
	r.stateMx.Lock()
	r.state = action
	r.stateMx.Unlock()
}

func (r *Runtime) Exists() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if _, err := r.client.ContainerInspect(ctx, r.ID(), client.ContainerInspectOptions{}); err != nil {
		if errdefs.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

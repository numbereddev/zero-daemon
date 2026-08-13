package container

import (
	"sync"

	"github.com/moby/moby/client"
	"github.com/numbereddev/zero-daemon/internal/events"
)

type Container struct {
	sync.RWMutex
	client *client.Client
	exists bool

	id string

	// TODO: needs to be reworked, just temp
	state   string
	stateMx sync.Mutex

	// TODO: do something with this
	stream   client.HijackedResponse
	eventBus *events.Bus[[]byte]
}

type Config struct{}

// TODO: Configuration
func New(cli *client.Client, cfg Config) *Container {
	return &Container{
		client:   cli,
		eventBus: events.NewBus[[]byte](),
	}
}

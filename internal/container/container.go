package container

import (
	"sync"

	"github.com/moby/moby/client"
	"github.com/numbereddev/zero-daemon/internal/events"
)

type Container struct {
	sync.RWMutex
	client *client.Client
	// TODO: there gotta be a better way
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

func (c *Container) SetID(id string) {
	c.Lock()
	c.id = id
	c.Unlock()
}

func (c *Container) ID() string {
	c.RLock()
	defer c.RUnlock()
	return c.id
}

// TODO: there gotta be a better way
func (c *Container) Exists() bool {
	c.RLock()
	defer c.RUnlock()
	return c.exists
}

// TODO: there gotta be a better way
func (c *Container) SetExists(exists bool) {
	c.Lock()
	c.exists = exists
	c.Unlock()
}

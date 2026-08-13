package container

import "github.com/numbereddev/zero-daemon/internal/events"

func (c *Container) Emitter() events.ReadBus[[]byte] {
	return c.eventBus
}

func (c *Container) Exists() bool {
	c.RLock()
	defer c.RUnlock()
	return c.exists
}

func (c *Container) SetState(action string) {
	c.stateMx.Lock()
	c.state = action
	c.stateMx.Unlock()
}

func (c *Container) State() string {
	c.stateMx.Lock()
	defer c.stateMx.Unlock()
	return c.state
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
